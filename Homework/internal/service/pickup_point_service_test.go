package service

import (
	"Homework/internal/model"
	"errors"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Write(t *testing.T) {
	t.Parallel()
	t.Run("add pickup point", func(t *testing.T) {
		t.Parallel()
		//arrange
		s := setUp(t)
		defer s.tearDown()
		s.mockPickupPointStorage.EXPECT().List().Return([]model.PickupPoint{})
		s.mockPickupPointStorage.EXPECT().Write("pvz1", "spb", "mail@mail.ru")
		// act
		err := s.pickupPointService.Write("pvz1", "spb", "mail@mail.ru")
		// assert
		require.Equal(t, nil, err)
	})
	t.Run("existing pickup point error", func(t *testing.T) {
		t.Parallel()
		//arrange
		s := setUp(t)
		defer s.tearDown()
		s.mockPickupPointStorage.EXPECT().List().Return([]model.PickupPoint{model.PickupPoint{
			Name:           "pvz1",
			Address:        "msk",
			ContactDetails: "mail",
		}})
		//act
		err := s.pickupPointService.Write("pvz1", "spb", "mail@mail.ru")
		//assert
		require.Equal(t, errors.New("пвз существует"), err)
	})
}

func Test_GetPickupPoint(t *testing.T) {
	t.Parallel()
	testPickupPoint := model.PickupPoint{
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "mail@mail.ru",
	}
	gotError := errors.New("пвз не существует")

	t.Run("get pickup point", func(t *testing.T) {
		t.Parallel()
		//arrange
		s := setUp(t)
		defer s.tearDown()
		s.mockPickupPointStorage.EXPECT().Get("pvz1").Return(testPickupPoint, nil)
		//act
		pickupPoint, err := s.pickupPointService.GetPickupPoint("pvz1")
		//assert
		require.Equal(t, nil, err)
		assert.Equal(t, testPickupPoint, pickupPoint)
	})

	t.Run("no pickup point error", func(t *testing.T) {
		t.Parallel()
		//arrange
		s := setUp(t)
		defer s.tearDown()
		s.mockPickupPointStorage.EXPECT().Get("pvz1").Return(model.PickupPoint{}, gotError)
		//act
		_, err := s.pickupPointService.GetPickupPoint("pvz1")
		//assert
		require.Equal(t, gotError, err)
	})
}

func Test_Read(t *testing.T) {
	t.Parallel()
	t.Run("smoke test", func(t *testing.T) {
		t.Parallel()
		//arrange
		s := setUp(t)
		defer s.tearDown()
		s.mockPickupPointStorage.EXPECT().List().Return([]model.PickupPoint{})
		//act
		pickupPointsList := s.pickupPointService.Read()
		//assert
		require.Equal(t, []model.PickupPoint{}, pickupPointsList)
	})
}
