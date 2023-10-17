package middleware

import (
	"net/url"

	oidcauth "github.com/TJM/gin-gonic-oidcauth"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/config"
)

var globalAuth *oidcauth.OidcAuth
var redirectPath string

// SetupAuthentication will setup the gin.Engine (router) with the necessary routes for authentication
func SetupAuthentication(router *gin.Engine) {
	auth := getAuth()

	router.GET("/login", auth.Login) // Unnecessary, as requesting a "AuthRequired" resource will initiate login, but potentially convenient
	router.GET(redirectPath, auth.AuthCallback)
	router.GET("/logout", auth.Logout)
}

// Authenticate will reutrn a gin.HandlerFunc to authenticate users
func Authenticate() gin.HandlerFunc {
	return getAuth().AuthRequired()
}

// createAuth will configure a new authentication object and configure it
func createAuth() *oidcauth.OidcAuth {
	log.Info("Create Authentication Configuration")
	args := config.GetArgs()
	// NOTE: DefaultConfig uses Google Accounts
	// - See https://github.com/coreos/go-oidc/blob/v3/example/README.md
	authConfig := oidcauth.DefaultConfig() // Supply OIDC Params via env
	auth, err := authConfig.GetOidcAuth()
	if err != nil {
		panic("AUTH setup failed: " + err.Error())
	}
	if args.DebugAuth {
		auth.Debug = true
	}
	redirectURL, err := url.Parse(authConfig.RedirectURL)
	if err != nil {
		panic("RedirectURL is INVALID: " + err.Error())
	}
	redirectPath = redirectURL.Path
	return auth
}

// getAuth will return the active oidcAuth object (or create one then return it)
func getAuth() *oidcauth.OidcAuth {
	if globalAuth == nil {
		globalAuth = createAuth()
	}
	return globalAuth
}
