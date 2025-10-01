//go:build linux

package watcher

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/s3rj1k/go-fanotify/fanotify"
	"golang.org/x/sys/unix"
	"vm-server/models"
	scanService "vm-server/services/scan"
)

type Manager struct {
	notify    *fanotify.NotifyFD
	service   *scanService.Service
	mu        sync.RWMutex
	active    map[string]int // ref count of activated absolute paths (directories)
	stopCh    chan struct{}
	stopped   bool
	dedupeMu  sync.Mutex
	recent    map[string]time.Time
	recentTTL time.Duration
	opts      Options
}

type Options struct {
	Mount   bool
	Verbose bool
}

func NewManager(svc *scanService.Service, opts Options) (*Manager, error) {
	// Proactive environment checks with suggestions
	warnEnv(opts)

	n, err := fanotify.Initialize(
		unix.FAN_CLOEXEC|
			unix.FAN_CLASS_NOTIF|
			unix.FAN_UNLIMITED_QUEUE|
			unix.FAN_UNLIMITED_MARKS,
		os.O_RDONLY|
			unix.O_LARGEFILE|
			unix.O_CLOEXEC,
	)
	if err != nil {
		return nil, err
	}
	m := &Manager{
		notify:    n,
		service:   svc,
		active:    make(map[string]int),
		stopCh:    make(chan struct{}),
		recent:    make(map[string]time.Time),
		recentTTL: 2 * time.Second,
		opts:      opts,
	}
	go m.loop()
	return m, nil
}

func (m *Manager) isRecent(key string) bool {
	m.dedupeMu.Lock()
	defer m.dedupeMu.Unlock()
	now := time.Now()
	for k, t := range m.recent {
		if now.Sub(t) > m.recentTTL {
			delete(m.recent, k)
		}
	}
	if t, ok := m.recent[key]; ok && now.Sub(t) <= m.recentTTL {
		return true
	}
	m.recent[key] = now
	return false
}

// ActivatePath marks a directory for watching (non-recursive, child events only).
func (m *Manager) ActivatePath(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return fmt.Errorf("watcher stopped")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return err
	}
	if count := m.active[abs]; count > 0 {
		m.active[abs] = count + 1
		return nil
	}
	// Add mark
	mask := uint64(unix.FAN_OPEN | unix.FAN_ACCESS | unix.FAN_CLOSE_WRITE | unix.FAN_OPEN_EXEC)
	markFlags := uint(unix.FAN_MARK_ADD)
	if m.opts.Mount {
		markFlags |= uint(unix.FAN_MARK_MOUNT)
		if m.opts.Verbose {
			fmt.Printf("[watcher] mount-mark on %s mask=0x%x\n", abs, mask)
		}
	} else {
		mask |= uint64(unix.FAN_EVENT_ON_CHILD)
		if m.opts.Verbose {
			fmt.Printf("[watcher] dir-mark on %s mask=0x%x\n", abs, mask)
		}
	}
	if err := m.notify.Mark(markFlags, mask, unix.AT_FDCWD, abs); err != nil {
		return err
	}
	m.active[abs] = 1
	return nil
}

// DeactivatePath decreases refcount and removes mark when zero.
func (m *Manager) DeactivatePath(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	abs, err := filepath.Abs(p)
	if err != nil {
		return err
	}
	if count := m.active[abs]; count > 1 {
		m.active[abs] = count - 1
		return nil
	}
	if _, ok := m.active[abs]; ok {
		// Remove mark
		_ = m.notify.Mark(uint(unix.FAN_MARK_REMOVE), 0, unix.AT_FDCWD, abs)
		delete(m.active, abs)
	}
	return nil
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return
	}
	close(m.stopCh)
	m.stopped = true
}

func (m *Manager) loop() {
	for {
		select {
		case <-m.stopCh:
			return
		default:
		}
		data, err := m.notify.GetEvent(os.Getpid())
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			// Sleep a bit on errors to avoid tight loop
			time.Sleep(25 * time.Millisecond)
			continue
		}
		if data == nil {
			continue
		}
		func() {
			defer data.Close()
			p, err := data.GetPath()
			if err != nil || p == "" {
				return
			}
			// Filter by active prefixes
			m.mu.RLock()
			match := false
			for base := range m.active {
				if strings.HasPrefix(p, base) {
					match = true
					break
				}
			}
			m.mu.RUnlock()
			if !match {
				return
			}

			// Build event key for dedupe
			pid := data.GetPID()
			var kinds []string
			if data.MatchMask(unix.FAN_OPEN) {
				kinds = append(kinds, "OPEN")
			}
			if data.MatchMask(unix.FAN_ACCESS) {
				kinds = append(kinds, "ACCESS")
			}
			if data.MatchMask(unix.FAN_CLOSE_WRITE) {
				kinds = append(kinds, "CLOSE_WRITE")
			}
			if data.MatchMask(unix.FAN_OPEN_EXEC) {
				kinds = append(kinds, "OPEN_EXEC")
			}
			if len(kinds) == 0 {
				kinds = append(kinds, "OTHER")
			}
			comm := readProcComm(pid)
			exe := readProcExe(pid)
			ruid, euid, _, _, ok := readProcUIDs(pid)
			if !ok {
				ruid, euid = -1, -1
			}
			dedupeKey := strings.Join(kinds, "|") + "|" + p + "|" + comm + "|" + strconv.Itoa(ruid)
			if m.isRecent(dedupeKey) {
				return
			}
			if m.opts.Verbose {
				fmt.Printf("[watcher] %s pid=%d uid=%d euid=%d comm=%q path=%q\n", strings.Join(kinds, "|"), pid, ruid, euid, comm, p)
			}
			// Upsert
			_ = m.service.UpsertConsumer(&models.Consumer{
				FilePath: p,
				Comm:     comm,
				Exe:      exe,
				RUID:     ruid,
				EUID:     euid,
			})
		}()
	}
}

func readProcComm(pid int) string {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func readProcExe(pid int) string {
	p := fmt.Sprintf("/proc/%d/exe", pid)
	target, err := os.Readlink(p)
	if err != nil {
		return ""
	}
	return target
}

func readProcUIDs(pid int) (int, int, int, int, bool) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, 0, 0, 0, false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				r, _ := strconv.Atoi(fields[1])
				e, _ := strconv.Atoi(fields[2])
				s, _ := strconv.Atoi(fields[3])
				fs, _ := strconv.Atoi(fields[4])
				return r, e, s, fs, true
			}
			break
		}
	}
	return 0, 0, 0, 0, false
}

// ---- Environment checks ----
func warnEnv(opts Options) {
	// Only meaningful on linux
	if runtime.GOOS != "linux" {
		return
	}
	euid := os.Geteuid()
	capEff := readSelfCapEff()
	// Check mount marks capability if requested
	if opts.Mount && !capEff.has(21) && euid != 0 { // 21 = CAP_SYS_ADMIN
		fmt.Printf("[watcher] warning: -watch-mount requested but CAP_SYS_ADMIN is missing (euid=%d)\n", euid)
		fmt.Printf("[watcher] suggestion: run with sudo/root, or setcap 'cap_sys_admin+ep' on the binary, or run container with --cap-add SYS_ADMIN or --privileged\n")
	}
	// Check /proc readability for other processes (comm/exe/uid)
	if !canReadProcPID(1) && euid != 0 {
		fmt.Printf("[watcher] warning: cannot read /proc/<pid>/status for other processes (hidepid or permissions)\n")
		fmt.Printf("[watcher] suggestion: run as root, or grant 'cap_dac_read_search', or remount /proc with hidepid=0 if acceptable\n")
	}
}

type capMask uint64

func readSelfCapEff() capMask {
	f, err := os.Open("/proc/self/status")
	if err != nil {
		return 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "CapEff:") {
			// CapEff: 	0000000000000000
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// hex
				if v, err := strconv.ParseUint(parts[1], 16, 64); err == nil {
					return capMask(v)
				}
			}
			break
		}
	}
	return 0
}

func (m capMask) has(capNum int) bool {
	if capNum < 0 || capNum >= 64 {
		return false
	}
	return (uint64(m) & (1 << uint(capNum))) != 0
}

func canReadProcPID(pid int) bool {
	// Try a cheap read: open status
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return false
	}
	f.Close()
	return true
}
