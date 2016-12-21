package main

import (
	"context"
	"github.com/dchest/captcha"
	"html/template"
	"net/http"
)

type ErrorMsg struct {
	Text string
}

func ErrorPage(w http.ResponseWriter) {
	t, _ := template.ParseFiles("templates/error.html")
	t.Execute(w, nil)
}

func ErrorPageWithMsg(w http.ResponseWriter, msg string) {
	t, _ := template.ParseFiles("templates/error_msg.html")
	t.Execute(w, ErrorMsg{msg})
}

func ErrorPageEmailResend(w http.ResponseWriter, msg string) {
	t, _ := template.ParseFiles("templates/error_email_resend.html")
	t.Execute(w, ErrorMsg{msg})
}

func Signup(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/signup.html")

	d := struct {
		CaptchaId string
	}{
		captcha.New(),
	}

	t.Execute(w, &d)
}

func Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/login.html")
	t.Execute(w, nil)
}

func UserForgotPassword(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/forgot_password.html")
	t.Execute(w, nil)
}
