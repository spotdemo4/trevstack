package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresConnection(user, pass, host, port, name string) (*gorm.DB, error) {
	dsn := "host=" + host + " user=" + user + " password=" + pass + " dbname=" + name + " port=" + port + " sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
