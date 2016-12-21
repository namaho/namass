package main

import (
	"context"
	"database/sql"
	log "github.com/Sirupsen/logrus"
	"strconv"
)

type SSPort struct {
	Port     int    `json:"port"`
	Password string `json:"password"`
	State    int    `json:"state"`
}

func NewPort(ctx context.Context, userId, port int, password string) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("INSERT INTO `ssport` (`user_id`, `port`, `password`) VALUES(?, ?, ?)")
	if err != nil {
		log.Error(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(userId, port, password)
	if err != nil {
		log.Error(err)
	}

	return nil
}

func GetSSPortByUserId(ctx context.Context, userId int) (SSPort, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT port, password, enable FROM `ssport` where user_id=?")
	if err != nil {
		return SSPort{}, err
	}
	defer stmt.Close()

	var port int
	var password string
	var enable int
	err = stmt.QueryRow(userId).Scan(&port, &password, &enable)
	if err != nil {
		return SSPort{}, err
	}

	return SSPort{port, password, enable}, nil
}

func GetSSPortList(ctx context.Context) (map[string]string, error) {
	db := ctx.Value("db").(*sql.DB)

	rows, err := db.Query("SELECT port, password FROM `ssport` WHERE enable=1")
	if err != nil {
		return nil, err
	}

	portList := make(map[string]string)

	for rows.Next() {
		var port int
		var password string
		err = rows.Scan(&port, &password)
		portList[strconv.Itoa(port)] = password
	}

	return portList, nil
}

func UpdatePortState(ctx context.Context, port, state int) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("UPDATE `ssport` SET enable=? WHERE port=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(state, port)
	if err != nil {
		return err
	}

	return nil
}
