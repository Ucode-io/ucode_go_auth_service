package helper

import (
	"log"
	"net/smtp"
	"io/ioutil"
	net_http "net/http"
	"github.com/pkg/errors"
	"encoding/json"
)

const (
	// server we are authorized to send email through
	host     = "smtp.gmail.com"
	hostPort = ":587"

	// user we are authorizing as
	from     string = "ucode.udevs.io@gmail.com"
	password string = "gehwhgelispgqoql"
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

	auth := smtp.PlainAuth("", from, password, host)

	//  // // NOTE: Using the backtick here ` works like a heredoc, which is why all the
	//  // // rest of the lines are forced to the beginning of the line, otherwise the
	//  // // formatting is wrong for the RFC 822 style
	//  msg := `To: "` + to + `" <` + to + `>
	// From: "` + from + `" <` + from + `>
	// Subject: ` + subject + `
	// ` + message
	msg := "To: \"" + to + "\" <" + to + ">\n" +
		"From: \"" + from + "\" <" + from + ">\n" +
		"Subject: " + subject + "\n" +
		message + "\n"

	if err := smtp.SendMail(host+hostPort, auth, from, []string{to}, []byte(msg)); err != nil {
		return errors.Wrap(err, "error while sending message to email")
	}

	return nil
}

func SendCodeToEmail(subject, to, code string) error {

	log.Printf("---SendCodeEmail---> email: %s, code: %s", to, code)

	message := `
		Ваше код подверждение: ` + code

	auth := smtp.PlainAuth("", from, password, host)

	//  // // NOTE: Using the backtick here ` works like a heredoc, which is why all the
	//  // // rest of the lines are forced to the beginning of the line, otherwise the
	//  // // formatting is wrong for the RFC 822 style
	//  msg := `To: "` + to + `" <` + to + `>
	// From: "` + from + `" <` + from + `>
	// Subject: ` + subject + `
	// ` + message
	msg := "To: \"" + to + "\" <" + to + ">\n" +
		"From: \"" + from + "\" <" + from + ">\n" +
		"Subject: " + subject + "\n" +
		message + "\n"

	if err := smtp.SendMail(host+hostPort, auth, from, []string{to}, []byte(msg)); err != nil {
		return errors.Wrap(err, "error while sending message to email")
	}

	return nil
}
