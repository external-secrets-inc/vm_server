package models

type AlreadyExistErr struct {
}

func (e *AlreadyExistErr) Error() string {
	return "already exists"
}

type NotFoundErr struct {
}

func (e *NotFoundErr) Error() string {
	return "not found"
}
