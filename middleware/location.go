package middleware

import (
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

// HandleLocation will setup the location handler
// - detects the scheme (http/https) and hostname of the server via the http.Request
func HandleLocation() gin.HandlerFunc {
	locConfig := location.DefaultConfig()
	locConfig.Headers.Host = "X-Forwarded-Host" // BUG in location middleware
	return location.New(locConfig)
}
