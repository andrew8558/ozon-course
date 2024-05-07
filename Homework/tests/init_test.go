//go:build integration
// +build integration

package tests

import (
	"Homework/tests/postgresql"
	"github.com/joho/godotenv"
	"log"
)

var db *postgresql.TDB

func init() {
	if err := godotenv.Load("test.env"); err != nil {
		log.Print("No .env file found")
	}
	db = postgresql.NewFromEnv()
}
