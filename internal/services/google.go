package services

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

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
			"https://www.googleapis.com/auth/gmail.readonly",
		}, //--The list of user's data that we require from google
		Endpoint: google.Endpoint, //--The end-points which are provided by google for oauth2
	}
	oauthStateStringGl = ""
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
		logger.Log.Warn("Code not found..")
		w.Write([]byte("Code Not Found to provide AccessToken..\n"))
		reason := r.FormValue("error_reason")
		if reason == "user_denied" {
			w.Write([]byte("User has denied Permission.."))
		}
		return
	}

	token, err := oauthConfGl.Exchange(oauth2.NoContext, code)
	if err != nil {
		logger.Log.Error("oauthConfGl.Exchange() failed with " + err.Error() + "\n")
		return
	}
	logger.Log.Info("TOKEN>> AccessToken>> " + token.AccessToken)
	logger.Log.Info("TOKEN>> Expiration Time>> " + token.Expiry.String())
	logger.Log.Info("TOKEN>> RefreshToken>> " + token.RefreshToken)

	//--GMAIL Get Client with Token
	saveToken(token)
	client := oauthConfGl.Client(context.Background(), token)
	gmailService, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	//--GMAIL Get User Labels
	labelsStr := ""
	user := "me"
	listLabelResp, err := gmailService.Users.Labels.List(user).Do()
	if err != nil {
		logger.Log.Fatal("unable to retrieve labels: " + err.Error())
	}
	for _, l := range listLabelResp.Labels {
		labelsStr += l.Name + ", "
	}
	if len(listLabelResp.Labels) == 0 {
		labelsStr = "No labels found."
	}

	//--Get User Info
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

	m := tk.M{}
	if err := tk.Unjson(response, &m); err != nil {
		logger.Log.Error("Unjson error")
		return
	}
	m.Set("labels", labelsStr)

	w.Write([]byte("Hello, I'm protected\n"))
	w.Write([]byte(tk.JsonStringIndent(m, "    ")))
	return
}

// Saves a token to a file path.
func saveToken(token *oauth2.Token) {
	path := viper.GetString("token-output-path")
	logger.Log.Info("Saving credential file to: " + path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logger.Log.Error("Unable to cache oauth token: +" + err.Error())
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
