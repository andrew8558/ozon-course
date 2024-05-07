package service

import (
	mocks_pickup_point_service "Homework/internal/service/mocks"
	"github.com/golang/mock/gomock"
	"testing"
)

type pickupPointStorageFixtures struct {
	ctrl                   *gomock.Controller
	pickupPointService     PickupPointService
	mockPickupPointStorage *mocks_pickup_point_service.MockpickupPointStorage
}

func setUp(t *testing.T) pickupPointStorageFixtures {
	ctrl := gomock.NewController(t)
	mockPickupPointStorage := mocks_pickup_point_service.NewMockpickupPointStorage(ctrl)
	pickupPointService := PickupPointService{mockPickupPointStorage}
	return pickupPointStorageFixtures{
		ctrl:                   ctrl,
		pickupPointService:     pickupPointService,
		mockPickupPointStorage: mockPickupPointStorage,
	}
}

func (p *pickupPointStorageFixtures) tearDown() {
	p.ctrl.Finish()
}
