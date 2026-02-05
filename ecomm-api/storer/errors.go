package storer

import (
	"fmt"
	"time"
)

type NotFoundError struct {
	Op        string      // Операция, например "storer.GetProduct"
	Resource  string      // Тип ресурса, например "product"
	ID        interface{} // Идентификатор ресурса
	Timestamp time.Time   // Время возникновения ошибки
	Err       error       // Оборачиваемая ошибка (может быть nil)
}

func NewNotFoundError(op, resource string, id interface{}, err error) *NotFoundError {
	return &NotFoundError{
		Op:        op,
		Resource:  resource,
		ID:        id,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func (e *NotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("operation %s: %s with id %v not found: %s", e.Op, e.Resource, e.ID, e.Err.Error())
	}
	return fmt.Sprintf("operation %s: %s with id %v not found", e.Op, e.Resource, e.ID)
}

// Unwrap позволяет использовать errors.Is и errors.As для обернутой ошибки.
func (e *NotFoundError) Unwrap() error {
	return e.Err
}

type EmailAlreadyExistsError struct {
	Op        string
	Resource  string
	ID        interface{}
	Timestamp time.Time
	Err       error
}

func (e *EmailAlreadyExistsError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("operation %s: %s with id %v not found: %s", e.Op, e.Resource, e.ID, e.Err.Error())
	}
	return fmt.Sprintf("operation %s: %s with id %v not found", e.Op, e.Resource, e.ID)
}

// Unwrap позволяет использовать errors.Is и errors.As для обернутой ошибки.
func (e *EmailAlreadyExistsError) Unwrap() error {
	return e.Err
}
