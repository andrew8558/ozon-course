package postgresql

import (
	"Homework/internal/repository"
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetByID(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()
		id  = int64(1)
	)
	t.Run("smoke test", func(t *testing.T) {
		t.Parallel()
		// arrange
		s := setUp(t)
		defer s.tearDown()
		s.mockDB.EXPECT().Get(gomock.Any(), gomock.Any(), "SELECT id, name, address, contact_details FROM pickup_points WHERE id=$1", id).Return(nil)
		// act
		pickupPoint, err := s.repo.GetByID(ctx, id)
		// assert
		require.NoError(t, err)
		assert.Equal(t, int64(0), pickupPoint.ID)
	})

	t.Run("fail", func(t *testing.T) {
		t.Parallel()
		t.Run("not found", func(t *testing.T) {
			t.Parallel()
			// arrange
			s := setUp(t)
			defer s.tearDown()
			s.mockDB.EXPECT().Get(gomock.Any(), gomock.Any(), "SELECT id, name, address, contact_details FROM pickup_points WHERE id=$1", id).
				Return(pgx.ErrNoRows)
			// act
			pickupPoint, err := s.repo.GetByID(ctx, id)
			// assert
			require.EqualError(t, err, "not found")
			require.True(t, errors.Is(err, repository.ErrObjectNotFound))
			assert.Equal(t, repository.PickupPoint{}, pickupPoint)
		})
		t.Run("internal error", func(t *testing.T) {
			t.Parallel()
			// arrange
			s := setUp(t)
			defer s.tearDown()
			s.mockDB.EXPECT().Get(gomock.Any(), gomock.Any(), "SELECT id, name, address, contact_details FROM pickup_points WHERE id=$1", id).
				Return(assert.AnError)
			// act
			pickupPoint, err := s.repo.GetByID(ctx, id)
			// assert
			require.EqualError(t, err, "assert.AnError general error for testing")
			assert.Equal(t, repository.PickupPoint{}, pickupPoint)
		})
	})
}

func Test_List(t *testing.T) {
	t.Parallel()
	var (
		ctx             = context.Background()
		pickupPointList = []repository.PickupPoint(nil)
	)

	t.Run("smoke test", func(t *testing.T) {
		t.Parallel()
		// arrange
		s := setUp(t)
		defer s.tearDown()
		s.mockDB.EXPECT().Select(gomock.Any(), gomock.Any(), "SELECT id, name, address, contact_details FROM pickup_points", gomock.Any()).Return(nil)
		// act
		pickupPoints, err := s.repo.List(ctx)
		// assert
		require.NoError(t, err)
		assert.Equal(t, pickupPointList, pickupPoints)
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()
		// arrange
		s := setUp(t)
		defer s.tearDown()
		s.mockDB.EXPECT().Select(gomock.Any(), gomock.Any(), "SELECT id, name, address, contact_details FROM pickup_points", gomock.Any()).
			Return(assert.AnError)
		// act
		_, err := s.repo.List(ctx)
		// assert
		require.EqualError(t, err, "assert.AnError general error for testing")
	})
}
