//go:generate mockgen -source ./repository.go -destination=./mocks/repository.go -package=mock_repository
package repository

import "context"

type PickupPointRepo interface {
	Add(ctx context.Context, pickupPoint PickupPoint) (int64, error)
	GetByID(ctx context.Context, id int64) (PickupPoint, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]PickupPoint, error)
	Update(ctx context.Context, pickupPoint PickupPoint) error
}
