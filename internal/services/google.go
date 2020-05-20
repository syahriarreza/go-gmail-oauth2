package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/syahriarreza/go-gmail-oauth2/internal/helper"
	"github.com/syahriarreza/go-gmail-oauth2/internal/logger"

	tk "github.com/eaciit/toolkit"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var (
	oauthConfGl = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "http://localhost:9090/callback-gl",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			gmail.GmailReadonlyScope,
			gmail.GmailComposeScope,
		}, //--The list of user's data that we require from google
		Endpoint: google.Endpoint, //--The end-points which are provided by google for oauth2
	}
	oauthStateStringGl = ""
	authenticationCode = ""
)

//InitializeOAuthGoogle Function
func InitializeOAuthGoogle() {
	oauthConfGl.ClientID = viper.GetString("google.clientID")
	oauthConfGl.ClientSecret = viper.GetString("google.clientSecret")
	oauthStateStringGl = viper.GetString("oauthStateString")
}

//HandleGoogleLogin Function
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	HandleLogin(w, r, oauthConfGl, oauthStateStringGl)
}

//CallBackFromGoogle Function
func CallBackFromGoogle(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Callback-gl..")

	state := r.FormValue("state")
	logger.Log.Info("state: " + state)
	if state != oauthStateStringGl {
		logger.Log.Info("invalid oauth state, expected " + oauthStateStringGl + ", got " + state + "\n")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	logger.Log.Info("code: " + code)

	if code == "" {
		errMsg := "Code Not Found to provide AccessToken.."
		if r.FormValue("error_reason") == "user_denied" {
			errMsg = "User has denied Permission.."
		}
		w.Write([]byte(errMsg))
		return
	}
	authenticationCode = code

	//--GMAIL Service
	gmailService, err := GetGmailService()
	if err != nil {
		logger.Log.Fatal("Unable to retrieve Gmail service: " + err.Error())
		return
	}

	//--GMAIL Get User Labels
	labelsStr := "No labels found."
	user := "me"
	listLabelResp, err := gmailService.Users.Labels.List(user).Do()
	if err != nil {
		logger.Log.Error("unable to retrieve labels: " + err.Error())
	} else {
		labelsStr = ""
		for _, l := range listLabelResp.Labels {
			labelsStr += l.Name + ", "
		}
	}

	//--Get User Info
	token, _ := getToken(authenticationCode)
	logger.Log.Info("GET: https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	if err != nil {
		logger.Log.Error("Get User Info: " + err.Error() + "\n")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("ReadAll: " + err.Error() + "\n")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	respM := tk.M{}
	if err := tk.Unjson(response, &respM); err != nil {
		logger.Log.Error("Unjson error")
		return
	}
	respM.Set("labels", labelsStr)

	t, err := template.ParseFiles("views/dashboard.html")
	if err != nil {
		logger.Log.Error("Unable to load template: dashboard.html")
	}
	t.Execute(w, respM)
}

//GetGmailService get gmail service based on token
func GetGmailService() (*gmail.Service, error) {
	//--Get Token
	token, err := getToken(authenticationCode)
	if err != nil {
		logger.Log.Error("error get token: " + err.Error())
		return nil, err
	}

	//--GMAIL Get Client with Token
	client := oauthConfGl.Client(context.Background(), token)
	gmailService, err := gmail.New(client)
	if err != nil {
		logger.Log.Error("error get gmail service: " + err.Error())
		return nil, err
	}

	return gmailService, nil
}

//SendEmail send email
func SendEmail(w http.ResponseWriter, r *http.Request) {
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

	gmailService, err := GetGmailService()
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

// Retrieves a token
func getToken(code string) (*oauth2.Token, error) {
	tok := &oauth2.Token{}
	path := viper.GetString("token-output-path")
	exchangeNeeded := false

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		exchangeNeeded = true
	} else {
		if err := json.NewDecoder(f).Decode(tok); err != nil {
			return nil, fmt.Errorf("error decode token: %s", err.Error())
		}

		if time.Now().After(tok.Expiry) { //--check Token expiry
			exchangeNeeded = true
		}
	}

	if exchangeNeeded {
		tok, err = oauthConfGl.Exchange(oauth2.NoContext, code)
		if err != nil {
			return nil, fmt.Errorf("oauthConfGl.Exchange() failed with " + err.Error())
		}
		saveToken(tok)
	}

	logger.Log.Info("TOKEN>> AccessToken>> " + tok.AccessToken)
	logger.Log.Info("TOKEN>> Expiration Time>> " + tok.Expiry.String())
	logger.Log.Info("TOKEN>> RefreshToken>> " + tok.RefreshToken)
	return tok, err
}

// Saves a token to a file
func saveToken(token *oauth2.Token) {
	path := viper.GetString("token-output-path")
	logger.Log.Info("saving token file to: " + path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logger.Log.Error("Unable to cache oauth token: +" + err.Error())
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
