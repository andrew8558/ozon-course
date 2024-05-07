package postgresql

import (
	mock_database "Homework/internal/db/mocks"
	"Homework/internal/repository"
	"github.com/golang/mock/gomock"
	"testing"
)

type pickupPointsRepoFixtures struct {
	ctrl   *gomock.Controller
	repo   repository.PickupPointRepo
	mockDB *mock_database.MockDBops
}

func setUp(t *testing.T) pickupPointsRepoFixtures {
	ctrl := gomock.NewController(t)
	mockDB := mock_database.NewMockDBops(ctrl)
	repo := NewPickupPoints(mockDB)
	return pickupPointsRepoFixtures{
		ctrl:   ctrl,
		repo:   repo,
		mockDB: mockDB,
	}
}

func (p *pickupPointsRepoFixtures) tearDown() {
	p.ctrl.Finish()
}
