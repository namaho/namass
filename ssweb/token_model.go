package main

import (
	"context"
	"database/sql"
)

func SaveToken(ctx context.Context, token string, userId int) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("INSERT INTO `token` (`user_id`, `token`) VALUES(?, ?) ON DUPLICATE KEY UPDATE token=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userId, token, token)
	if err != nil {
		return err
	}

	return nil
}

func GetUserIdByToken(ctx context.Context, token string) (int, error) {
	db := ctx.Value("db").(*sql.DB)
	stmt, err := db.Prepare("SELECT user_id FROM `token` where token=?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var userId int
	err = stmt.QueryRow(token).Scan(&userId)
	if err != nil {
		return 0, err
	}

	return userId, nil
}
