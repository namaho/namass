package main

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	Id         int
	Email      string
	Password   string
	IsVerified bool
	VerifyCode string
}

func NewUser(ctx context.Context, email, password, verifyCode, ip string) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("INSERT INTO `user` (`email`, `password`, `verify_code`, `reg_time`, `reg_ip`) VALUES(?, ?, ?, ?, INET_ATON(?))")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(email, password, verifyCode, time.Now().Unix(), ip)
	if err != nil {
		return err
	}

	return nil
}

func GetUserByEmail(ctx context.Context, email string) (User, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT id, email, password, is_verified, verify_code FROM `user` where email=?")
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()

	var id int
	var password string
	var isVerified bool
	var verifyCode string
	err = stmt.QueryRow(email).Scan(&id, &email, &password, &isVerified, &verifyCode)
	if err != nil {
		return User{}, err
	}

	return User{id, email, password, isVerified, verifyCode}, nil
}

func GetEmailByVerifyCode(ctx context.Context, verifyCode string) (string, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT email FROM `user` where verify_code=?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var email string
	err = stmt.QueryRow(verifyCode).Scan(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

func GetVerifyCodeByEmail(ctx context.Context, email string) (string, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT verify_code FROM `user` where email=?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var verifyCode string
	err = stmt.QueryRow(email).Scan(&verifyCode)
	if err != nil {
		return "", err
	}

	return verifyCode, nil
}

func GetVerifyEmailResendsByEmail(ctx context.Context, email string) (int, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT verify_email_resends FROM `user` where email=?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var resends int
	err = stmt.QueryRow(email).Scan(&resends)
	if err != nil {
		return 0, err
	}

	return resends, nil
}

func IncreaseVerifyEmailResends(ctx context.Context, email string) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("UPDATE `user` SET verify_email_resends=verify_email_resends+1 WHERE email=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(email)
	if err != nil {
		return err
	}

	return nil
}

func GetUserIdByEmail(ctx context.Context, email string) (int, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT id FROM `user` where email=?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(email).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetUserById(ctx context.Context, id int) (User, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT id, email, password, is_verified, verify_code FROM `user` where id=?")
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()

	var email string
	var password string
	var isVerified bool
	var verifyCode string
	err = stmt.QueryRow(id).Scan(&id, &email, &password, &isVerified, &verifyCode)
	if err != nil {
		return User{}, err
	}

	return User{id, email, password, isVerified, verifyCode}, nil
}

func GetUserByVerifyCode(ctx context.Context, verifyCode string) (User, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT id, email, password, is_verified, verify_code FROM `user` where verify_code=?")
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()

	var id int
	var email string
	var password string
	var isVerified bool
	err = stmt.QueryRow(verifyCode).Scan(&id, &email, &password, &isVerified, &verifyCode)
	if err != nil {
		return User{}, err
	}

	return User{id, email, password, isVerified, verifyCode}, nil
}

func VerifyUser(ctx context.Context, verifyCode string) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("UPDATE `user` SET is_verified=1 WHERE verify_code=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(verifyCode)
	if err != nil {
		return err
	}

	return nil
}

func IsVerified(ctx context.Context, verifyCode string) (bool, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT is_verified FROM `user` where verify_code=?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var isVerified bool
	err = stmt.QueryRow(verifyCode).Scan(&isVerified)
	if err != nil {
		return false, err
	}

	return isVerified, nil
}

func CountSignupIP(ctx context.Context, ip string) (int, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT count(1) cnt FROM `user` where reg_ip=INET_ATON(?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var cnt int
	err = stmt.QueryRow(ip).Scan(&cnt)
	if err != nil {
		return 0, err
	}

	return cnt, nil
}

func GetIPByEmail(ctx context.Context, email string) (string, error) {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("SELECT INET_NTOA(reg_ip) cnt FROM `user` where email=?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var ip string
	err = stmt.QueryRow(email).Scan(&ip)
	if err != nil {
		return "", err
	}

	return ip, nil
}

func UpdateIP(ctx context.Context, email, ip string) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("UPDATE `user` SET reg_ip=INET_ATON(?) WHERE email=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(ip, email)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUserPassword(ctx context.Context, id int, password string) error {
	db := ctx.Value("db").(*sql.DB)

	stmt, err := db.Prepare("UPDATE `user` SET password=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(password, id)
	if err != nil {
		return err
	}

	return nil
}
