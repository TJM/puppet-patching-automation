package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/chat"
	"github.com/tjm/puppet-patching-automation/models"
)

// ListChatRooms endpoint (GET)
func ListChatRooms(c *gin.Context) {
	rooms := models.GetChatRooms()
	data := gin.H{"status": "success", "rooms": rooms}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "chatRoom-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, rooms.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// GetChatRoom endpoint (GET)
// PathParams: id
func GetChatRoom(c *gin.Context) {
	room, err := getChatRoom(c)
	if err != nil {
		return // error has already been logged
	}

	data := gin.H{"status": "success", "room": room}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "chatRoom-show.gohtml",
		HTMLData: getHTMLData(c, room.GetBreadCrumbs(), data),
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// UpdateChatRoom endpoint (PUT)
// - PathParams: id
func UpdateChatRoom(c *gin.Context) {
	room, err := getChatRoom(c)
	if err != nil {
		return // error has already been logged
	}

	err = c.Bind(room)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Handle DISABLE (Enable not checked)
	if c.PostForm("Enabled") == "" {
		room.Enabled = false
	}

	room.Save()
	data := gin.H{"status": "success", "room": room}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "chatRoom-success-redirect.gohtml",
		Data:     data,
		HTMLData: gin.H{},
		Offered:  formatAllSupported,
	})
}

// DeleteChatRoom endpoint (DELETE)
// - PathParams: id
func DeleteChatRoom(c *gin.Context) {
	room, err := getChatRoom(c)
	if err != nil {
		return // error has already been logged
	}
	err = room.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "message": "Deleted"} // no room_id, redirects to /config/ChatRoom/
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "chatRoom-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// TestChatRoom endpoint (GET)
// PathParams: id
func TestChatRoom(c *gin.Context) {
	room, err := getChatRoom(c)
	if err != nil {
		return // error has already been logged
	}

	chat := chat.NewChat(room)
	event := models.NewEvent(models.ActionTest)
	event.ThreadKey = "PATestEvent"
	chat.HandleEvent(event)

	data := gin.H{"status": "success"}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "chatRoom-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// LinkChatRoomToPatchRun endpoint (POST)
// Currently not used from the WebUI, keeping for future API use
// PathParams: id
func LinkChatRoomToPatchRun(c *gin.Context) {
	// PatchRun
	patchRun, err := getPatchRun(c)
	if err != nil {
		return
	}

	err = linkChatRoomsToPatchRun(c, patchRun)
	if err != nil {
		// Error should already be output
		return
	}

	data := gin.H{"status": "success", "patch_run_id": patchRun.ID, "ChatRooms": patchRun.ChatRooms}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "patchRun-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// linkChatRoomsToPatchRun will link the ChatRooms to the Patch Run
func linkChatRoomsToPatchRun(c *gin.Context, patchRun *models.PatchRun) (err error) {
	// ChatRooms
	roomIDs, err := convertSliceStringToUint(c.PostFormArray("rooms"))
	if err != nil {
		log.Error("Error getting []uint from rooms param: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	log.WithField("rooms", roomIDs).Debug("Link Chat Rooms")
	rooms, err := models.GetChatRoomsByIDs(roomIDs)
	if err != nil {
		log.Error("Error retrieving rooms from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	err = patchRun.LinkChatRooms(rooms)
	if err != nil {
		log.Error("Error saving patchRun: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
	}
	return
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getChatRoom will get the id from context and return job
func getChatRoom(c *gin.Context) (room *models.ChatRoom, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		if errors.Is(err, errIDNew) {
			room = models.NewChatRoom()
			err = nil
		}
		return
	}
	return getChatRoomByID(c, id)
}

// getChatRoomByID retrives the room from the DB
func getChatRoomByID(c *gin.Context, id uint) (room *models.ChatRoom, err error) {
	// Get room from DB
	room, err = models.GetChatRoomByID(id)
	if err != nil {
		log.Error("Error retrieving room from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving room from DB: " + err.Error()})
		return
	}
	// Another check to verify the room was retrieved, id should not be 0
	if room.ID == 0 {
		err = errNotExist
		log.Error("Error room id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error room id should not be 0 (not found)"})
		return
	}
	return // success
}
