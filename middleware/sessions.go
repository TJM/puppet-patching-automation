package middleware

import (
	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"

	"github.com/tjm/puppet-patching-automation/config"
	"github.com/tjm/puppet-patching-automation/models"
)

// HandleSession will setup and handle sessions
func HandleSession() gin.HandlerFunc {
	args := config.GetArgs()

	// Session Config (store sessions in GoRM DB)
	store := gormsessions.NewStore(models.GetDB(), true, []byte(args.SessionAuthKey), []byte(args.SessionEncKey))
	return sessions.Sessions(args.SessionName, store)
}
