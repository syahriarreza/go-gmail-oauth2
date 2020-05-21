package helper

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/syahriarreza/go-gmail-oauth2/internal/logger"
	"golang.org/x/oauth2"
)

//GetToken retrieves a token
func GetToken(oauthConfGl *oauth2.Config, code string) (*oauth2.Token, error) {
	tok := &oauth2.Token{}

	f, err := os.Open(viper.GetString("token-output-path"))
	if err != nil {
		if code == "" {
			return nil, fmt.Errorf("code is not provided and could not open token file (%s): %s", viper.GetString("token-output-path"), err.Error())
		}

		//--token not found, then get new one
		tok, err = oauthConfGl.Exchange(oauth2.NoContext, code)
		if err != nil {
			return nil, fmt.Errorf("oauthConfGl.Exchange() failed with " + err.Error())
		}
		SaveToken(tok)
	} else {
		defer f.Close()
		if err := json.NewDecoder(f).Decode(tok); err != nil {
			return nil, fmt.Errorf("error decode token: %s", err.Error())
		}
	}

	if time.Now().After(tok.Expiry) { //--check Token expiry
		logger.Log.Info("token has been expired: " + tok.Expiry.Format(time.RFC1123))
	}

	logger.Log.Info("TOKEN>> AccessToken: " + tok.AccessToken)
	logger.Log.Info("TOKEN>> Expiration Time: " + tok.Expiry.String())
	logger.Log.Info("TOKEN>> RefreshToken: " + tok.RefreshToken)
	return tok, err
}

//SaveToken saves a token to a file
func SaveToken(token *oauth2.Token) {
	path := viper.GetString("token-output-path")
	logger.Log.Info("saving token file to: " + path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logger.Log.Error("unable to cache oauth token: +" + err.Error())
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
