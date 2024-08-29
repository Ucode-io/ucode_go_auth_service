package helper

import (
	"encoding/json"
	"fmt"
	"io"
	net_http "net/http"
	"net/smtp"
	"ucode/ucode_go_auth_service/config"

	"github.com/pkg/errors"
)

const (
	// server we are authorized to send email through
	host     = "smtp.gmail.com"
	hostPort = ":587"

	// user we are authorizing as  old="gehwhgelispgqoql"  new="xkiaqodjfuielsug"
	from            string = "ucode.udevs.io@gmail.com"
	defaultPassword string = "xkiaqodjfuielsug"
)

type SendMessageToEmailRequest struct {
	Subject       string
	To            string
	UserId        string
	Email         string
	Password      string
	Username      string
	TempPassword  string
	EnvironmentId string
	ProjectId     string
	ClientTypeId  string
}

func GetGoogleUserInfo(accessToken string) (map[string]interface{}, error) {
	resp, err := net_http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + accessToken)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	userInfo := make(map[string]interface{})

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

func SendEmail(subject, to, link, token string) error {
	message := `
		You can update your password using the following url
   
	   ` + link + "?token=" + token

	auth := smtp.PlainAuth("", from, defaultPassword, host)
	msg := "To: \"" + to + "\" <" + to + ">\n" +
		"From: \"" + from + "\" <" + from + ">\n" +
		"Subject: " + subject + "\n" +
		message + "\n"

	if err := smtp.SendMail(host+hostPort, auth, from, []string{to}, []byte(msg)); err != nil {
		return errors.Wrap(err, "error while sending message to email")
	}

	return nil
}

func SendCodeToEmail(subject, to, code string, email string, password string) error {

	message := `
		Your verification code is: ` + code

	// if email == "" {
	// 	email = from
	// }
	// if password == "" {
	// 	password = defaultPassword
	// }

	auth := smtp.PlainAuth("", email, password, host)

	msg := "To: \"" + to + "\" <" + to + ">\n" +
		"From: \"" + email + "\" <" + email + ">\n" +
		"Subject: " + subject + "\n" +
		message + "\n"

	if err := smtp.SendMail(host+hostPort, auth, from, []string{to}, []byte(msg)); err != nil {
		return errors.Wrap(err, "error while sending message to email")
	}

	return nil
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}

func SendCodeToEnvironmentEmail(subject, to, code string, email string, password string) error {

	smtpServer := "outlook.office365.com"
	smtpPort := 587

	message := `
		Your verification code is ` + code

	msg := "To: \"" + to + "\" <" + to + ">\n" +
		"From: \"" + email + "\" <" + email + ">\n" +
		"Subject: " + subject + "\n" +
		message + "\n"

	auth := LoginAuth(email, password)

	err := smtp.SendMail(fmt.Sprintf("%s:%d", smtpServer, smtpPort), auth, email, []string{to}, []byte(msg))
	if err != nil {
		return errors.Wrap(err, "error while sending message to environment email")
	}

	return nil
}

func SendInviteMessageToEmail(input SendMessageToEmailRequest) error {

	cfg := config.Load()
	url := fmt.Sprintf(
		cfg.UcodeAppBaseUrl+"/invite-user?user_id=%s&environment_id=%s&project_id=%s&client_type_id=%s",
		input.UserId,
		input.EnvironmentId,
		input.ProjectId,
		input.ClientTypeId,
	)
	message := fmt.Sprintf(`
			Dear %s,

			I hope this message finds you well. I am pleased to invite you to access our admin panel, where you will have the ability to manage and oversee various aspects of our system.
			
			Below, you will find the details needed to access the admin panel:
			
			Admin Panel Link: %s
			Login: %s
			Temporary Password: %s`, input.To, url, input.Username, input.TempPassword)

	if input.Email == "" || input.Password == "" {
		input.Email = from
		input.Password = defaultPassword
	}

	auth := smtp.PlainAuth("", input.Email, input.Password, host)

	msg := "To: \"" + input.To + "\" <" + input.To + ">\n" +
		"From: \"" + input.Email + "\" <" + input.Email + ">\n" +
		"Subject: " + input.Subject + "\n" +
		message + "\n"

	if err := smtp.SendMail(host+hostPort, auth, from, []string{input.To}, []byte(msg)); err != nil {
		return errors.Wrap(err, "error while sending message to email")
	}

	return nil
}
