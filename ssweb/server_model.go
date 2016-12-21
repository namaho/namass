package main

import (
	"context"
	"database/sql"
)

func GetServerIPList(ctx context.Context) ([]string, error) {
	db := ctx.Value("db").(*sql.DB)

	rows, err := db.Query("SELECT ip FROM `server` where is_up=1")
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)

	for rows.Next() {
		var ip string
		err = rows.Scan(&ip)
		ips = append(ips, ip)
	}

	return ips, nil
}

func SaveServer(ctx context.Context, ip string, area int) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("INSERT INTO `server` (`ip`, `area`, `is_up`) VALUES(?, ?, 1)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(ip, area)
	if err != nil {
		return err
	}

	return nil
}

func GetServerIdByIP(ctx context.Context, ip string) (int, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT id FROM `server` where ip=?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(ip).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
