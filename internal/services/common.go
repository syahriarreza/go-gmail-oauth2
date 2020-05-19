package services

import (
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/syahriarreza/go-gmail-oauth2/internal/logger"
	"golang.org/x/oauth2"
)

//HandleMain Function renders the index page when the application index route is called
func HandleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	t, err := template.ParseFiles("views/index.html")
	if err != nil {
		logger.Log.Error("Unable to load template: index.html")
	}
	t.Execute(w, nil)
}

//HandleLogin Function
func HandleLogin(w http.ResponseWriter, r *http.Request, oauthConf *oauth2.Config, oauthStateString string) {
	URL, err := url.Parse(oauthConf.Endpoint.AuthURL)
	if err != nil {
		logger.Log.Error("Parse: " + err.Error())
	}
	logger.Log.Info(URL.String())
	parameters := url.Values{}
	parameters.Add("access_type", "offline")
	parameters.Add("client_id", oauthConf.ClientID)
	parameters.Add("scope", strings.Join(oauthConf.Scopes, " "))
	parameters.Add("redirect_uri", oauthConf.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", oauthStateString)
	URL.RawQuery = parameters.Encode()
	url := URL.String()
	logger.Log.Info("Auth URL: " + url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
