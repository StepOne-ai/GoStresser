package database

import (
	"testing"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestDBConnection(t *testing.T) {
	dsn := "host=localhost user=stresser password=stepan2005 dbname=go_stresser port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	DB = db
}

func TestTestDBConnection(t *testing.T) {
	dsn := "host=localhost user=stresser password=stepan2005 dbname=go_stresser_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	DB = db
}