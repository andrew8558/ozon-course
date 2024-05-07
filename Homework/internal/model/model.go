package model

import (
	"errors"
	"time"
)

type OrderStatus string

const (
	Accepted OrderStatus = "accepted"
	Issued   OrderStatus = "issued"
	Refunded OrderStatus = "refunded"
)

type Order struct {
	OrderID     string
	CustomerId  string
	TermKeeping time.Time
	DateIssue   time.Time
	Status      OrderStatus
	Weight      float32
	Price       float32
}

type OrderOutputDTO struct {
	OrderId     string
	CustomerId  string
	TermKeeping time.Time
	Status      OrderStatus
	Weight      float32
	Price       float32
}

type OrderInputDTO struct {
	OrderId     string
	CustomerId  string
	TermKeeping time.Time
	Weight      float32
	Price       float32
}

type PickupPoint struct {
	Name           string
	Address        string
	ContactDetails string
}

type BatchPacker struct {
}

func (p BatchPacker) Pack(order OrderInputDTO) (OrderInputDTO, error) {
	if order.Weight > 10 {
		return order, errors.New("вес заказа больше 10 кг")
	}
	order.Price += 5
	return order, nil
}

type BoxPacker struct {
}

func (p BoxPacker) Pack(order OrderInputDTO) (OrderInputDTO, error) {
	if order.Weight > 30 {
		return order, errors.New("вес заказа больше 30 кг")
	}
	order.Price += 20
	return order, nil

}

type EnvelopePacker struct {
}

func (p EnvelopePacker) Pack(order OrderInputDTO) (OrderInputDTO, error) {
	order.Price += 1
	return order, nil
}
