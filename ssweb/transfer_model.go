package main

import (
	"context"
	"database/sql"
)

func UpdateTransfer(ctx context.Context, port, transfer int) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("INSERT INTO `transfer` (`port`, `transfer`) VALUES(?, ?) ON DUPLICATE KEY UPDATE transfer=transfer+?, last_update=UNIX_TIMESTAMP()")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(port, transfer, transfer)
	if err != nil {
		return err
	}

	return nil
}

func GetTransferByPort(ctx context.Context, port int) (int, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT transfer FROM `transfer` where port=?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var transfer int
	err = stmt.QueryRow(port).Scan(&transfer)
	if err != nil {
		return 0, err
	}

	return transfer, nil
}
