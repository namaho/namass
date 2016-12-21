package main

import (
	"database/sql"
	"fmt"
)

func NewDB(config Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		config.DB.User,
		config.DB.Password,
		config.DB.Host,
		config.DB.Port,
		config.DB.Database)

	return sql.Open(config.DB.Driver, dsn)
}
