package postgresql

import (
	"Homework/internal/db"
	"Homework/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
)

type PickupPointRepo struct {
	db db.DBops
}

func NewPickupPoints(database db.DBops) *PickupPointRepo {
	return &PickupPointRepo{db: database}
}

func (r *PickupPointRepo) Add(ctx context.Context, pickupPoint repository.PickupPoint) (int64, error) {
	var id int64
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			fmt.Println(err)
		}
	}()

	err = tx.QueryRow(ctx, "INSERT INTO pickup_points(name, address, contact_details) VALUES ($1, $2, $3) RETURNING id;", pickupPoint.Name, pickupPoint.Address, pickupPoint.ContactDetails).Scan(&id)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PickupPointRepo) GetByID(ctx context.Context, id int64) (repository.PickupPoint, error) {
	var p repository.PickupPoint
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.PickupPoint{}, err
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			fmt.Println(err)
		}
	}()

	err = tx.QueryRow(ctx, "SELECT id, name, address, contact_details FROM pickup_points WHERE id=$1", id).Scan(p)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.PickupPoint{}, repository.ErrObjectNotFound
		}
		return repository.PickupPoint{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return repository.PickupPoint{}, err
	}

	return p, nil
}

func (r *PickupPointRepo) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			fmt.Println(err)
		}
	}()

	commandTag, err := tx.Exec(ctx, "DELETE FROM pickup_points WHERE id=$1", id)
	if commandTag.RowsAffected() == 0 {
		return repository.ErrObjectNotFound
	}
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *PickupPointRepo) List(ctx context.Context) ([]repository.PickupPoint, error) {
	var pickupPoints []repository.PickupPoint
	err := r.db.Select(ctx, &pickupPoints, "SELECT id, name, address, contact_details FROM pickup_points")
	if err != nil {
		return nil, err
	}
	return pickupPoints, nil
}

func (r *PickupPointRepo) Update(ctx context.Context, pickupPoint repository.PickupPoint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			fmt.Println(err)
		}
	}()

	commandTag, err := r.db.Exec(ctx, "UPDATE pickup_points SET name=$1, address=$2, contact_details=$3 WHERE id=$4", pickupPoint.Name, pickupPoint.Address, pickupPoint.ContactDetails, pickupPoint.ID)
	if commandTag.RowsAffected() == 0 {
		return repository.ErrObjectNotFound
	}
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
