package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// HandleCORS will handle the CORS related functions and configuration
func HandleCORS() gin.HandlerFunc {
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://google.com"}
	// corsConfig.AllowAllOrigins = true
	log.Debugf("corsConfig: %+v", corsConfig)
	return cors.New(corsConfig)
}
