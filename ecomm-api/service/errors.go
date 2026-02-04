package service

import (
	"fmt"
	"time"
)

type ErrNotFound struct {
	Op        string      // Операция, например "storer.GetProduct"
	Resource  string      // Тип ресурса, например "product"
	ID        interface{} // Идентификатор ресурса
	Timestamp time.Time   // Время возникновения ошибки
	Err       error       // Оборачиваемая ошибка (может быть nil)
}

func NewErrNotFound(op, resource string, id interface{}, err error) *ErrNotFound {
	return &ErrNotFound{
		Op:        op,
		Resource:  resource,
		ID:        id,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func (e *ErrNotFound) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("operation %s: %s with id %v not found: %s", e.Op, e.Resource, e.ID, e.Err.Error())
	}
	return fmt.Sprintf("operation %s: %s with id %v not found", e.Op, e.Resource, e.ID)
}

// Unwrap позволяет использовать errors.Is и errors.As для обернутой ошибки.
func (e *ErrNotFound) Unwrap() error {
	return e.Err
}
