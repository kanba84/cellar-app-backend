package database

import (
	"net"
	"net/url"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "cellar_app"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "cellar_app"
	}

	connectionURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, os.Getenv("DB_PASSWORD")),
		Host:   net.JoinHostPort(host, port),
		Path:   dbName,
	}
	query := connectionURL.Query()
	query.Set("sslmode", "disable")
	connectionURL.RawQuery = query.Encode()

	db, err := gorm.Open(postgres.Open(connectionURL.String()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
