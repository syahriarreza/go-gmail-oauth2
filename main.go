package main

import (
	"log"
	"net/http"

	"github.com/syahriarreza/go-gmail-oauth2/internal/configs"
	"github.com/syahriarreza/go-gmail-oauth2/internal/logger"
	"github.com/syahriarreza/go-gmail-oauth2/internal/services"

	"github.com/spf13/viper"
)

func main() {
	// Initialize Viper across the application
	configs.InitializeViper()

	// Initialize Logger across the application
	logger.InitializeZapCustomLogger()

	// Initialize Oauth2 Services [OLD]
	// services.InitializeOAuthFacebook()
	// services.InitializeOAuthGoogle()

	// Routes for the application
	http.HandleFunc("/", services.HandleMain)
	// http.HandleFunc("/login-fb", services.HandleFacebookLogin)
	// http.HandleFunc("/callback-fb", services.CallBackFromFacebook)
	// http.HandleFunc("/login-gl", services.HandleGoogleLogin)
	// http.HandleFunc("/callback-gl", services.CallBackFromGoogle)
	// http.HandleFunc("/send-email", services.SendEmail)
	http.HandleFunc("/login-gl-api", services.HandleGoogleAPILogin)
	http.HandleFunc("/callback-gl", services.CallBackFromGoogleAPI)
	http.HandleFunc("/send-email", services.SendEmailAPI)

	logger.Log.Info("Started running on http://localhost:" + viper.GetString("port"))
	log.Fatal(http.ListenAndServe(":"+viper.GetString("port"), nil))
}
