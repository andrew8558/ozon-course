package storage

import (
	"Homework/internal/model"
	"encoding/json"
	"errors"
	"os"
	"sync"
)

const pickupPointStorageName = "pickup_point_storage"

type PickupPointStorage struct {
	mapPickupPoint map[string]model.PickupPoint
	mx             sync.RWMutex
}

func NewPickupPointStorage() (PickupPointStorage, error) {
	_, err := os.OpenFile(pickupPointStorageName, os.O_CREATE, 0777)
	if err != nil {
		return PickupPointStorage{}, err
	}

	rawBytes, err := os.ReadFile(pickupPointStorageName)
	if err != nil {
		return PickupPointStorage{}, err
	}

	pickupPoints := map[string]model.PickupPoint{}
	if len(rawBytes) == 0 {
		return PickupPointStorage{mapPickupPoint: pickupPoints}, nil
	}
	err = json.Unmarshal(rawBytes, &pickupPoints)
	if err != nil {
		return PickupPointStorage{}, err
	}
	return PickupPointStorage{mapPickupPoint: pickupPoints}, nil
}

func (p *PickupPointStorage) Save() error {
	rawBytes, err := json.Marshal(p.mapPickupPoint)
	if err != nil {
		return err
	}

	err = os.WriteFile(pickupPointStorageName, rawBytes, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (p *PickupPointStorage) Write(name string, address string, contactDetails string) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.mapPickupPoint[name] = model.PickupPoint{
		Name:           name,
		Address:        address,
		ContactDetails: contactDetails,
	}
}

func (p *PickupPointStorage) Get(name string) (model.PickupPoint, error) {
	p.mx.RLock()
	defer p.mx.RUnlock()

	pickupPoint, ok := p.mapPickupPoint[name]

	if ok {
		return pickupPoint, nil
	}

	return model.PickupPoint{}, errors.New("пвз не существует")
}

func (p *PickupPointStorage) List() []model.PickupPoint {
	p.mx.RLock()
	defer p.mx.RUnlock()
	pickupPointSlice := make([]model.PickupPoint, 0)
	for _, pickupPoint := range p.mapPickupPoint {
		pickupPointSlice = append(pickupPointSlice, pickupPoint)
	}
	return pickupPointSlice
}
