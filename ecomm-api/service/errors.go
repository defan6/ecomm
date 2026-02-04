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

type ErrNotEnoughStock struct {
	Op        string
	Resource  string
	ID        interface{}
	Requested int64
	Available int64
	Timestamp time.Time
	Err       error
}

func NewNotEnoughStock(op string, resource string, id interface{}, requested int64, available int64, err error) *ErrNotEnoughStock {
	return &ErrNotEnoughStock{
		Op:        op,
		Resource:  resource,
		ID:        id,
		Requested: requested,
		Available: available,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func (e *ErrNotEnoughStock) Error() string {
	return fmt.Sprintf("not enough stock for product with id %d. Requested: %d, Available: %d", e.ID, e.Requested, e.Available)
}

func (e *ErrNotEnoughStock) Unwrap() error {
	return e.Err
}

type ErrNotFoundProductForOrder struct {
	Op        string
	Resource  string
	ID        interface{}
	Timestamp time.Time
	Err       error
}

func NewErrNotFoundProductForOrder(op string, resource string, err error) *ErrNotFoundProductForOrder {
	return &ErrNotFoundProductForOrder{
		Op:        op,
		Resource:  resource,
		Timestamp: time.Now(),
		Err:       err,
	}
}

func (e *ErrNotFoundProductForOrder) Error() string {
	return fmt.Sprintf("operation %s: Some %s not found", e.Op, e.Resource)
}

func (e *ErrNotFoundProductForOrder) Unwrap() error {
	return e.Err
}
