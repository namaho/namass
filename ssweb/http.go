package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/dchest/captcha"
	"net/http"
)

type API struct {
	signup             string
	login              string
	userSignup         string
	userLogin          string
	userInfo           string
	userVerify         string
	userEmailResend    string
	serverDiscover     string
	transferReport     string
	userForgotPassword string
	userSendResetEmail string
	userResetPassword  string
	captcha            string
	userEnablePort     string
}

const API_PREFIX = "/ssweb"

var api = API{
	API_PREFIX + "/signup",
	API_PREFIX + "/login",
	API_PREFIX + "/user/signup",
	API_PREFIX + "/user/login",
	API_PREFIX + "/user/info",
	API_PREFIX + "/user/verify",
	API_PREFIX + "/user/email/resend",
	API_PREFIX + "/server/discover",
	API_PREFIX + "/transfer/report",
	API_PREFIX + "/user/forgot_password",
	API_PREFIX + "/user/send_reset_email",
	API_PREFIX + "/user/reset_password",
	API_PREFIX + "/captcha/",
	API_PREFIX + "/user/enable_port/",
}

func HandleRequest(request string, ctx context.Context, handler func(context.Context, http.ResponseWriter, *http.Request)) {
	http.Handle(request, &ContextAdapter{ctx, ContextHandlerFunc(handler)})
}

func StartHTTPServer(ctx context.Context) {
	config := ctx.Value("config").(Config)

	HandleRequest(api.signup, ctx, Signup)
	HandleRequest(api.login, ctx, Login)
	HandleRequest(api.userSignup, ctx, UserSignup)
	HandleRequest(api.userLogin, ctx, UserLogin)
	HandleRequest(api.userInfo, ctx, InfoPage)
	HandleRequest(api.userVerify, ctx, UserVerify)
	HandleRequest(api.userEmailResend, ctx, UserEmailResend)
	http.Handle(api.serverDiscover, ValidateServerToken(ctx, &ContextAdapter{ctx, ContextHandlerFunc(ServerDiscover)}))
	http.Handle(api.transferReport, ValidateServerToken(ctx, &ContextAdapter{ctx, ContextHandlerFunc(ReportTransfer)}))
	HandleRequest(api.userForgotPassword, ctx, UserForgotPassword)
	HandleRequest(api.userSendResetEmail, ctx, UserSendResetEmail)
	HandleRequest(api.userResetPassword, ctx, UserResetPassword)
	http.Handle(api.captcha, captcha.Server(captcha.StdWidth, captcha.StdHeight))
	HandleRequest(api.userEnablePort, ctx, UserEnablePort)

	log.Info("listening on " + config.Listen)
	http.ListenAndServe(config.Listen, nil)
}
