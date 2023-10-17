package events

import (
	"fmt"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"

	"github.com/tjm/puppet-patching-automation/controllers/chat"
	"github.com/tjm/puppet-patching-automation/models"
)

// PatchRunEvent will handle the event for the list of rooms
func PatchRunEvent(c *gin.Context, patchRun *models.PatchRun, event *models.Event) {
	event.PatchRun = patchRun
	event.ThreadKey = fmt.Sprint(patchRun.ID)
	event.URL = location.Get(c)
	for _, room := range patchRun.ChatRooms {
		if room.Enabled {
			chat := chat.NewChat(room)
			chat.HandleEvent(event)
		}
	}
}
