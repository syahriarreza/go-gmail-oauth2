package services

import (
	"net/http"
	"net/url"
	"strings"
	"text/template"

	tk "github.com/eaciit/toolkit"
	"github.com/spf13/viper"
	"github.com/syahriarreza/go-gmail-oauth2/internal/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

//HandleMain Function renders the index page when the application index route is called
func HandleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	data := []tk.M{}
	data = append(data, tk.M{"title": "Google", "url": "/login-gl", "disabled": false})
	data = append(data, tk.M{"title": "Facebook", "url": "/login-fb", "disabled": true})
	data = append(data, tk.M{"title": "Google API", "url": "/login-gl-api", "disabled": false})
	data = append(data, tk.M{"title": "Send Mail using Credentials & Token", "url": "/callback-gl", "disabled": false})

	t, err := template.ParseFiles("views/index.html")
	if err != nil {
		logger.Log.Error("Unable to load template: index.html")
	}
	t.Execute(w, data)
}

//HandleLogin Function
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if oauthConfig.ClientID == "" {
		oauthConfig = &oauth2.Config{
			ClientID:     viper.GetString("google.clientID"),
			ClientSecret: viper.GetString("google.clientSecret"),
			RedirectURL:  viper.GetString("google.redirectURL"),
			Scopes:       ConfigScopes,
			Endpoint:     google.Endpoint, //--The end-points which are provided by google for oauth2
		}
	}

	authURL, err := url.Parse(oauthConfig.Endpoint.AuthURL)
	if err != nil {
		w.Write([]byte("Parse: " + err.Error()))
	}
	parameters := url.Values{}
	parameters.Add("access_type", "offline")
	parameters.Add("prompt", "consent")
	parameters.Add("client_id", oauthConfig.ClientID)
	parameters.Add("scope", strings.Join(oauthConfig.Scopes, " "))
	parameters.Add("redirect_uri", oauthConfig.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", viper.GetString("oauthStateString"))
	authURL.RawQuery = parameters.Encode()
	logger.Log.Info("Auth URL: " + authURL.String())
	http.Redirect(w, r, authURL.String(), http.StatusTemporaryRedirect)
}
