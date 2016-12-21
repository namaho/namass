package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/dchest/captcha"
	mysql "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"math/rand"
	"net"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

const (
	VERIFY_EMAIL_MAX_RESENDS       = 10
	VERIFY_EMAIL_SUBJECT           = "Verify your email"
	VERIFY_EMAIL_BODY_PART         = "Access the following link to verify your email address, and enjoy!  \r\n"
	RESET_PASSWORD_EMAIL_SUBJECT   = "Reset password"
	RESET_PASSWORD_EMAIL_BODY_PART = "Access the following link to reset your password:   \r\n"
	SIGNUP_LIMIT_FOR_SINGLE_IP     = 10
)

const (
	PORT_STATE_DISABLE = 0
	PORT_STATE_ENABLE  = 1
)

type PageInfo struct {
	Email     string
	Port      int
	Password  string
	PortState template.HTML
	Transfer  string
	QRCode    string
	USIP4     string
	JPIP4     string
}

type VerifyCode struct {
	Value string
}

type ByteSize float64

// const for bytesize. B is also specified.
const (
	B ByteSize = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

// Print readable values of byte size
func (b ByteSize) String() string {
	switch {
	case b >= YB:
		return fmt.Sprintf("%7.2f YB", b/YB)
	case b >= ZB:
		return fmt.Sprintf("%7.2f ZB", b/ZB)
	case b >= EB:
		return fmt.Sprintf("%7.2f EB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%7.2f PB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%7.2f TB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%7.2f GB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%7.2f MB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%7.2f KB", b/KB)
	}
	return fmt.Sprintf("%7.2f  B", b)
}

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func SendEmail(ctx context.Context, to, subject, content string) error {
	config := ctx.Value("config").(Config)

	from := config.Smtp.User
	password := config.Smtp.Password

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		content

	err := smtp.SendMail(
		config.Smtp.Address+":"+config.Smtp.Port,
		smtp.PlainAuth("", from, password, config.Smtp.Address),
		from, []string{to}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}

// Handle /ssweb/user/signup request, send a verify email.
func UserSignup(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	captchaId := r.FormValue("captchaId")
	captchaSolution := r.FormValue("captchaSolution")
	ip := r.Header.Get("x-forwarded-for")

	if !captcha.VerifyString(captchaId, captchaSolution) {
		ErrorPageWithMsg(w, "你输入的验证码有误")
		return
	}

	verifyCode, err := GenerateRandomString(64)
	if err != nil {
		log.Error(err)
		ErrorPage(w)
		return
	}

	count, err := CountSignupIP(ctx, ip)
	if err != nil {
		log.Error(err)
		ErrorPageWithMsg(w, "奇怪的错误出现了，快联系下管理员。")
		return
	}
	if count > SIGNUP_LIMIT_FOR_SINGLE_IP {
		log.Error("email " + email + " with ip " + ip + " has registered too much times, block")
		ErrorPageWithMsg(w, "出了奇怪的错误，快去联系管理员。")
		return
	}

	if password == "" {
		log.Error("block email " + email + " register with empty password")
		ErrorPageWithMsg(w, "不能空密码啊。")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("an error occur while hashing password")
		ErrorPageWithMsg(w, "出错了。")
		return
	}

	err = NewUser(ctx, email, string(hashedPassword), verifyCode, ip)
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == ER_DUP_ENTRY {
				ErrorPageWithMsg(w, "此邮箱已注册过。")
				log.Warn("email " + email + " has been registered")
			} else {
				ErrorPage(w)
				log.Error(err)
			}
		}
		return
	} else {
		log.Info("create user " + email + " (" + ip + ")")
	}

	config := ctx.Value("config").(Config)
	verifyLink := "http://" + config.SSWebHTTPAddress + api.userVerify + "?c=" + verifyCode

	err = SendEmail(ctx, email, VERIFY_EMAIL_SUBJECT, VERIFY_EMAIL_BODY_PART+verifyLink)
	if err != nil {
		log.Error(err)
		ErrorPage(w)
		return
	}

	http.Redirect(w, r, api.login, http.StatusSeeOther)
}

// Handle /ssweb/signup/verify request, verify register email, add a new port for user.
func UserVerify(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	verifyCode := r.URL.Query().Get("c")
	log.Info("verifing signup code " + verifyCode)

	isVerified, err := IsVerified(ctx, verifyCode)
	if err != nil {
		if err == sql.ErrNoRows {
			ErrorPageWithMsg(w, "验证码无效。请确保复制完整的验证链接粘贴到浏览器地址栏。")
			log.Error("invalid verify code: " + verifyCode)
			return
		} else {
			ErrorPageWithMsg(w, "未知错误，请联系管理员。")
			log.Error(err)
			return
		}
	} else if isVerified {
		log.Warn("email has already been verified")
		ErrorPageWithMsg(w, "邮箱已验证，请登陆。")
		return
	}

	err = VerifyUser(ctx, verifyCode)
	if err != nil {
		ErrorPage(w)
		log.Error(err)
		return
	}

	email, err := GetEmailByVerifyCode(ctx, verifyCode)
	if err != nil {
		if err == sql.ErrNoRows {
			ErrorPageWithMsg(w, "验证无效，此邮箱还未注册。")
			log.Error("invalid verify code: " + verifyCode)
		} else {
			ErrorPage(w)
			log.Error(err)
		}
		return
	}

	userId, err := GetUserIdByEmail(ctx, email)
	if err != nil {
		log.Error(err)
		ErrorPage(w)
		return
	}

	lastPort, err := GetLastPort(ctx)
	if err != nil {
		log.Error(err)
		ErrorPage(w)
		return
	}

	newPort := lastPort + 1
	config := ctx.Value("config").(Config)
	portPassword := config.SSPasswordPrefix + strings.ToUpper(RandomString(5))

	err = NewPort(ctx, userId, newPort, portPassword)
	if err != nil {
		log.Error(err)
		ErrorPage(w)
		return
	} else {
		log.Info("add port " + strconv.Itoa(newPort) + " for user " + email + " (" + strconv.Itoa(userId) + ")")
	}

	err = UpdateLastPort(ctx, newPort)
	if err != nil {
		log.Error(err)
		ErrorPage(w)
		return
	}

	err = SSCmdAddPort(ctx, newPort, portPassword)
	if err != nil {
		log.Error(err)
		ErrorPageWithMsg(w, "出现了某些错误，但可能不影响使用，如果真用不了请联系下管理员。")
		return
	}

	http.Redirect(w, r, api.login, http.StatusSeeOther)
}

func UserEnablePort(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("x-forwarded-for")
	tokenCookie, err := r.Cookie("ntk")
	if err != nil {
		ErrorPageWithMsg(w, "无权访问，请注册或重新登陆。")
		log.Warn("token not found")
		return
	}

	userId, err := GetUserIdByToken(ctx, tokenCookie.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("invalid token " + tokenCookie.Value)
			ErrorPageWithMsg(w, "Token失效，请重新登陆。请注意使用另一浏览器登陆时，原来的Token就会失效。")
			return
		}
		log.Error(err)
		ErrorPageWithMsg(w, "未知错误，请联系管理员。")
		return
	}

	ssport, err := GetSSPortByUserId(ctx, userId)
	if err != nil {
		log.Error(err)
		log.Debug("ssport")
		ErrorPage(w)
		return
	}

	err = SSCmdAddPort(ctx, ssport.Port, ssport.Password)
	if err != nil {
		log.Error(err)
		ErrorPageWithMsg(w, "出现了某些错误，但可能不影响使用，如果真用不了请联系下管理员。")
		return
	}

	err = UpdatePortState(ctx, ssport.Port, PORT_STATE_ENABLE)
	if err != nil {
		log.Error("an error occur while updating port state")
		ErrorPageWithMsg(w, "未知错误，请联系管理员。")
		log.Error(err)
		return
	}

	user, err := GetUserById(ctx, userId)
	if err != nil {
		log.Error(err)
	}

	log.Info("user " + user.Email + " (" + strconv.Itoa(userId) + ") " + "enable port " + strconv.Itoa(ssport.Port) + " (" + ip + ")")

	ErrorPageWithMsg(w, "你的端口已启用。")
}

// Handle /ssweb/user/email/resend request, resend email verify link.
func UserEmailResend(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	ip := r.Header.Get("x-forwarded-for")

	user, err := GetUserByEmail(ctx, email)
	if err != nil {
		log.Error(err)
		ErrorPageWithMsg(w, "此邮箱还未注册。")
		return
	}

	if user.IsVerified != true {
		resends, err := GetVerifyEmailResendsByEmail(ctx, email)
		if err != nil {
			log.Error(err)
			return
		}
		if resends > VERIFY_EMAIL_MAX_RESENDS {
			log.Warn("verify email resends more than " + strconv.Itoa(VERIFY_EMAIL_MAX_RESENDS) + " times, lock " + email)
			ErrorPageWithMsg(w, "你的重发次数太多了，邮箱已被锁定，请联系管理员。")
			return
		}

		verifyCode, err := GetVerifyCodeByEmail(ctx, email)
		if err != nil {
			log.Error(err)
			ErrorPageWithMsg(w, "你不应该看到这条信息，绝对出了什么奇怪错误了啊，请联系下管理员。")
			return
		}

		config := ctx.Value("config").(Config)
		verifyLink := "http://" + config.SSWebHTTPAddress + api.userVerify + "?c=" + verifyCode

		err = SendEmail(ctx, email, VERIFY_EMAIL_SUBJECT, VERIFY_EMAIL_BODY_PART+verifyLink)
		if err != nil {
			log.Error(err)
			ErrorPage(w)
			return
		}

		err = IncreaseVerifyEmailResends(ctx, email)
		if err != nil {
			log.Error(err)
			ErrorPageWithMsg(w, "出了什么奇怪错误了。")
			return
		}

		log.Info("resend verify link to email " + email + " (" + ip + ")")
		ErrorPageWithMsg(w, "已经向你的邮箱 "+email+" 重新发了一封验证邮件，请查收。")
		return
	} else {
		ErrorPageWithMsg(w, "你输入的邮箱已通过验证，请登陆。")
	}
}

func UserLogin(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	config := ctx.Value("config").(Config)

	user, err := GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("invalid login email " + email)
			ErrorPageWithMsg(w, "无效的登陆邮箱，可能是未注册。")
			return
		} else {
			ErrorPageWithMsg(w, "未知错误，请联系管理员。")
			log.Error(err)
		}
		return
	}

	if user.IsVerified != true {
		log.Warn("account not verified: " + email)
		ErrorPageEmailResend(w, "你的邮箱还没通过验证，一封验证邮件已经发送到你的注册邮箱，请到邮箱查收后完成注册。(Please check your email.)")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Warn("invalid login password " + email)
		ErrorPageWithMsg(w, "密码错误。")
		return
	}

	log.Info("user " + user.Email + " login")

	token, err := GenerateRandomString(32)
	if err != nil {
		log.Error(err)
		ErrorPage(w)
		return
	}

	err = SaveToken(ctx, token, user.Id)
	if err != nil {
		log.Error(err)
		ErrorPage(w)
		return
	}

	expiration := time.Now().Add(15 * 24 * time.Hour)
	cookie := http.Cookie{
		Name:     "ntk",
		Value:    token,
		Path:     "/",
		Domain:   config.SSWebHTTPAddress,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   false,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, api.userInfo, http.StatusFound)
}

// Handle /ssweb/user/info request, show useful information.
func InfoPage(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	config := ctx.Value("config").(Config)
	ip := r.Header.Get("x-forwarded-for")
	tokenCookie, err := r.Cookie("ntk")
	if err != nil {
		ErrorPageWithMsg(w, "无权访问，请注册或重新登陆。")
		log.Warn("token not found")
		return
	}

	userId, err := GetUserIdByToken(ctx, tokenCookie.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("invalid token " + tokenCookie.Value)
			ErrorPageWithMsg(w, "Token失效，请重新登陆。请注意使用另一浏览器登陆时，原来的Token就会失效。")
			return
		}
		log.Error(err)
		ErrorPageWithMsg(w, "未知错误，请联系管理员。")
		return
	}

	user, err := GetUserById(ctx, userId)
	if err != nil {
		log.Error(err)
		log.Debug("user")
		ErrorPage(w)
		return
	}

	ssport, err := GetSSPortByUserId(ctx, userId)
	if err != nil {
		log.Error(err)
		log.Debug("ssport")
		ErrorPage(w)
		return
	}

	transfer, err := GetTransferByPort(ctx, ssport.Port)
	if err != nil {
		transfer = 0
	}

	portState := map[int]string{0: "不可用(<a href=\"/ssweb/user/enable_port\">点击启用</a>)", 1: "可用"}
	serverAddresses := map[int]string{0: "jp." + config.SSServerDomain, 1: "us." + config.SSServerDomain}
	rand.Seed(time.Now().Unix())
	randomServerAddress := serverAddresses[rand.Intn(1001)%len(serverAddresses)]
	qrCode := "ss://" + base64.URLEncoding.EncodeToString([]byte("aes-256-cfb:"+ssport.Password+"@"+randomServerAddress+":"+strconv.Itoa(ssport.Port)))

	usIPs, err := net.LookupIP("us." + config.SSServerDomain)
	if err != nil {
		log.Error(err)
	}
	usIP4 := usIPs[0]

	jpIPs, err := net.LookupIP("jp." + config.SSServerDomain)
	if err != nil {
		log.Error(err)
	}
	jpIP4 := jpIPs[0]

	t, _ := template.ParseFiles("templates/info.html")
	t.Execute(w, PageInfo{
		Email:     user.Email,
		Port:      ssport.Port,
		Password:  ssport.Password,
		PortState: template.HTML(portState[ssport.State]),
		Transfer:  ByteSize(transfer).String(),
		QRCode:    qrCode,
		USIP4:     usIP4.String(),
		JPIP4:     jpIP4.String(),
	})
	log.Info("user " + user.Email + " request info page (" + ip + ")")

	ip2, err := GetIPByEmail(ctx, user.Email)
	if ip2 == "0.0.0.0" {
		UpdateIP(ctx, user.Email, ip)
		log.Info("update reg_ip for email " + user.Email)
	}
}

func UserSendResetEmail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	config := ctx.Value("config").(Config)
	ip := r.Header.Get("x-forwarded-for")

	user, err := GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("invalid email " + email + " requests for reset passwd (" + ip + ")")
			ErrorPageWithMsg(w, "无效的登陆邮箱，可能是未注册。")
			return
		} else {
			ErrorPageWithMsg(w, "未知错误，请联系管理员。")
			log.Error(err)
		}
		return
	}

	resetLink := "http://" + config.SSWebHTTPAddress + api.userResetPassword + "?c=" + user.VerifyCode

	err = SendEmail(ctx, email, RESET_PASSWORD_EMAIL_SUBJECT, RESET_PASSWORD_EMAIL_BODY_PART+resetLink)
	if err != nil {
		log.Error(err)
		ErrorPage(w)
		return
	}

	ErrorPageWithMsg(w, "一封密码重置邮件已发送到你的注册邮箱，请查收。")
}

func UserResetPassword(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("x-forwarded-for")

	switch r.Method {
	case "GET":
		verifyCode := r.URL.Query().Get("c")
		log.Info("verifing reset password code " + verifyCode)
		_, err := GetUserByVerifyCode(ctx, verifyCode)
		if err != nil {
			if err == sql.ErrNoRows {
				ErrorPageWithMsg(w, "验证码无效。请确定复制完整的链接粘贴到浏览器地址栏。")
				log.Error("invalid verify code: " + verifyCode)
				return
			} else {
				ErrorPageWithMsg(w, "未知错误，请联系管理员。")
				log.Error(err)
			}
			return
		}
		t, _ := template.ParseFiles("templates/reset_password.html")
		t.Execute(w, VerifyCode{verifyCode})
		return
	case "POST":
		password := r.FormValue("password")
		verifyCode := r.FormValue("verify_code")

		user, err := GetUserByVerifyCode(ctx, verifyCode)
		if err != nil {
			if err == sql.ErrNoRows {
				ErrorPageWithMsg(w, "验证码无效。请确定复制完整的链接粘贴到浏览器地址栏。")
				log.Error("invalid verify code: " + verifyCode)
				return
			} else {
				ErrorPageWithMsg(w, "未知错误，请联系管理员。")
				log.Error(err)
			}
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("an error occur while hashing password")
			ErrorPageWithMsg(w, "出错了。")
			return
		}

		err = UpdateUserPassword(ctx, user.Id, string(hashedPassword))
		if err != nil {
			log.Error("an error occur while updating user password")
			ErrorPageWithMsg(w, "未知错误，请联系管理员。")
			log.Error(err)
		}
		log.Info("password changed for user " + user.Email + " (" + ip + ")")
		ErrorPageWithMsg(w, "密码修改成功。")
	}

}
