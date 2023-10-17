package middleware

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/config"
)

// SetupTrustedProxies will configure a gin.Engine (router) with trusted proxies from our config/env
func SetupTrustedProxies(router *gin.Engine) {
	args := config.GetArgs()
	// Handle TrustedProxies if they are specified
	// See: https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies
	if len(args.TrustedProxies) > 0 {
		log.Debugf("Trust Proxies: %+v", args.TrustedProxies)
		err := router.SetTrustedProxies(args.TrustedProxies)
		if err != nil {
			log.Error("Error in setTrustedProxies: " + err.Error())
		}
	} else {
		err := router.SetTrustedProxies(nil) // Trust NO Proxies (c.ClientIP may be wrong if using Load Balancer)
		if err != nil {
			log.Error("Error in setTrustedProxies: " + err.Error())
		}
	}
}
