package helper

import (
	"encoding/json"
	_ "fmt"
	"io/ioutil"
	"log"
	net_http "net/http"
	"net/smtp"

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

func GetGoogleUserInfo(accessToken string) (map[string]interface{}, error) {
	resp, err := net_http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + accessToken)
	// fmt.Println("Request to https://www.googleapis.com/oauth2/v3/userinfo?access_token= " + accessToken)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	userInfo := make(map[string]interface{})

	body, err := ioutil.ReadAll(resp.Body)
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

	log.Printf("---SendCodeEmail---> email: %s, code: %s", to, code)

	message := `
		Код для подтверждения: ` + code

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
