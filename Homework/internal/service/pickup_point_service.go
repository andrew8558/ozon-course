//go:generate mockgen -source ./pickup_point_service.go -destination=./mocks/pickup_point_service.go -package=mocks_pickup_point_service
package service

import (
	"Homework/internal/model"
	"errors"
)

type pickupPointStorage interface {
	Write(name string, address string, contactDetails string)
	List() []model.PickupPoint
	Get(name string) (model.PickupPoint, error)
}

type PickupPointService struct {
	p pickupPointStorage
}

func NewPickupPointService(p pickupPointStorage) PickupPointService { return PickupPointService{p: p} }

func (p PickupPointService) Write(name string, address string, contactDetails string) error {
	pickupPoints := p.p.List()
	for _, pickupPoint := range pickupPoints {
		if name == pickupPoint.Name {
			return errors.New("пвз существует")
		}
	}
	p.p.Write(name, address, contactDetails)
	return nil
}

func (p PickupPointService) GetPickupPoint(name string) (model.PickupPoint, error) {
	pickupPoint, err := p.p.Get(name)
	if err != nil {
		return model.PickupPoint{}, err
	}
	return pickupPoint, nil
}

func (p PickupPointService) Read() []model.PickupPoint {
	return p.p.List()
}
