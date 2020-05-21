package services

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"

	tk "github.com/eaciit/toolkit"
	"github.com/spf13/viper"
	"github.com/syahriarreza/go-gmail-oauth2/internal/helper"
	"github.com/syahriarreza/go-gmail-oauth2/internal/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var (
	//ConfigScopes specifies optional requested permissions.
	ConfigScopes = []string{
		"https://www.googleapis.com/auth/userinfo.email",
		gmail.GmailReadonlyScope,
		gmail.GmailComposeScope,
	}

	oauthConfig = &oauth2.Config{}
	authCode    = ""
)

//HandleGoogleAPILogin handles google login
func HandleGoogleAPILogin(w http.ResponseWriter, r *http.Request) {
	credByte, err := ioutil.ReadFile(viper.GetString("credentials-output-path"))
	if err != nil {
		logger.Log.Error("unable to read credentials.json " +
			"(" + viper.GetString("credentials-output-path") + "), " +
			"then proceed with user manual authentication")
		HandleLogin(w, r)
		return
	}

	oauthConfig, err = google.ConfigFromJSON(credByte, ConfigScopes...) //--handles google login using credentials.json
	if err != nil {
		w.Write([]byte("unable to parse credentials file to config: " + err.Error()))
	}

	http.Redirect(w, r, "/callback-gl", http.StatusTemporaryRedirect)
}

//CallBackFromGoogleAPI CallBackFromGoogleAPI
func CallBackFromGoogleAPI(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("/callback-gl...")

	authCode = r.FormValue("code")
	logger.Log.Info("code: " + authCode)

	token, _ := helper.GetToken(oauthConfig, authCode)
	logger.Log.Info("GET: https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	if err != nil {
		w.Write([]byte("Get User Info: " + err.Error()))
		return
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.Write([]byte("ReadAll: " + err.Error()))
		return
	}
	logger.Log.Info("response: " + string(response))

	respM := tk.M{}
	if err := tk.Unjson(response, &respM); err != nil {
		w.Write([]byte("Unjson error"))
		return
	}

	t, err := template.ParseFiles("views/dashboard.html")
	if err != nil {
		logger.Log.Error("Unable to load template: dashboard.html")
	}
	t.Execute(w, respM)
}

//GetGmailServiceAPI get gmail service based on token
func GetGmailServiceAPI() (*gmail.Service, error) {
	token, _ := helper.GetToken(oauthConfig, authCode)

	client := oauthConfig.Client(context.Background(), token)
	gmailService, err := gmail.New(client)
	if err != nil {
		logger.Log.Error("error get gmail service: " + err.Error())
		return nil, err
	}

	return gmailService, nil
}

//SendEmailAPI send email
func SendEmailAPI(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	msgRFC := helper.MessageRFC2822{
		Headers: tk.M{
			"From":     "'" + r.Form.Get("from") + "'", //TODO: try change to gmail
			"reply-to": r.Form.Get("reply-to"),
			"To":       r.Form.Get("to"),
			"Subject":  r.Form.Get("subject"),
		},
		Body: r.Form.Get("content"),
	}

	msg := gmail.Message{
		Raw: msgRFC.EncodeRFC2822(),
	}

	gmailService, err := GetGmailServiceAPI()
	if err != nil {
		w.Write([]byte("unable to retrieve Gmail service: " + err.Error()))
		return
	}
	if _, err := gmailService.Users.Messages.Send("me", &msg).Do(); err != nil {
		w.Write([]byte("error sending email: " + err.Error()))
		return
	}

	w.Write([]byte("Email Sent"))
	return
}
