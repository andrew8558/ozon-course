package repository

import "errors"

var ErrObjectNotFound = errors.New("not found")

type PickupPoint struct {
	ID             int64  `db:"id"`
	Name           string `db:"name"`
	Address        string `db:"address"`
	ContactDetails string `db:"contact_details"`
}
