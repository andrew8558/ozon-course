package storage

import (
	"Homework/internal/model"
	"encoding/json"
	"os"
	"sync"
)

const storageName = "order_storage"

type Storage struct {
	storage *os.File
	mx      sync.RWMutex
}

func NewOrderStorage() (Storage, error) {
	file, err := os.OpenFile(storageName, os.O_CREATE, 0777)
	if err != nil {
		return Storage{}, err
	}
	return Storage{storage: file}, nil
}

func (s *Storage) Save(newOrder model.Order) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	all, err := s.listAll()
	if err != nil {
		return err
	}

	var flag = 0
	for index, order := range all {
		if newOrder.OrderID == order.OrderID {
			all[index] = newOrder
			flag = 1
			break
		}
	}
	if flag == 0 {
		all = append(all, newOrder)
	}

	err = writeBytes(all)
	if err != nil {
		return err
	}
	return nil
}

func writeBytes(orders []model.Order) error {
	rawBytes, err := json.Marshal(orders)
	if err != nil {
		return err
	}

	err = os.WriteFile(storageName, rawBytes, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) Delete(orderID string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	all, err := s.listAll()
	if err != nil {
		return err
	}

	for i, order := range all {
		if order.OrderID == orderID {
			all = remove(all, i)
			break
		}
	}
	err = writeBytes(all)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) List() ([]model.Order, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	all, err := s.listAll()
	if err != nil {
		return nil, err
	}
	return all, nil
}

func (s *Storage) listAll() ([]model.Order, error) {
	rawBytes, err := os.ReadFile(storageName)
	if err != nil {
		return nil, err
	}

	var orders []model.Order
	if len(rawBytes) == 0 {
		return orders, nil
	}
	err = json.Unmarshal(rawBytes, &orders)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func remove(slice []model.Order, s int) []model.Order {
	return append(slice[:s], slice[s+1:]...)
}
