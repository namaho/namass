package main

import (
	"context"
	"database/sql"
)

func GetLastPort(ctx context.Context) (int, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT `last_port` FROM `helper` LIMIT 1")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var lastPort int
	err = stmt.QueryRow().Scan(&lastPort)
	if err != nil {
		return 0, err
	}

	return lastPort, nil
}

func UpdateLastPort(ctx context.Context, port int) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("UPDATE helper SET last_port=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(port)
	if err != nil {
		return err
	}

	return nil
}
