package controllers

import (
	"net/http"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/events"
	"github.com/tjm/puppet-patching-automation/controllers/trelloapi"
	"github.com/tjm/puppet-patching-automation/models"
)

// GetTrelloBoards endpoint (GET)
// - PathParams: id (PatchRunID)
func GetTrelloBoards(c *gin.Context) {
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	boards := models.GetTrelloBoards(id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "boards": boards})
}

// GetTrelloBoard by ID endpoint (GET)
// - PathParams: id (trello_board_id)
// @Router /trelloboard/:id [GET]
func GetTrelloBoard(c *gin.Context) {
	board, err := getTrelloBoard(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "board": board})
}

// DeleteTrelloBoard by ID endpoint (POST)
// - PathParams: id (trello_board_id)
// @Router /trelloboard/:id [DELETE]
func DeleteTrelloBoard(c *gin.Context) {
	board, err := getTrelloBoard(c)
	if err != nil {
		return
	}
	patchRunID := board.PatchRunID
	err = trelloapi.DeleteTrelloBoard(board)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	patchRun, err := getPatchRunByID(c, patchRunID)
	if err != nil {
		return // error already output
	}
	event := models.NewEvent(models.ActionTrelloBoardDeleted)
	event.PatchRun = patchRun
	event.Target = board
	events.PatchRunEvent(c, patchRun, event)

	data := gin.H{"status": "success", "patch_run_id": patchRunID, "message": "DELETED"}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "patchRun-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// CreateTrelloBoard endpoint (POST)
// - PathParams: id
// - Post/Query Params: name, color, description
func CreateTrelloBoard(c *gin.Context) {
	patchRun, err := getPatchRun(c)
	if err != nil {
		return // error already output
	}
	board := new(models.TrelloBoard)
	err = c.Bind(board)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err})
		return
	}
	board.PatchRunID = patchRun.ID
	err = trelloapi.CreateTrelloBoard(board, false, location.Get(c).String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	event := models.NewEvent(models.ActionTrelloBoardCreated)
	event.PatchRun = patchRun
	event.Target = board
	events.PatchRunEvent(c, patchRun, event)

	data := gin.H{"status": "success", "message": "Populating board in background.", "patch_run_id": patchRun.ID, "board": board}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "patchRun-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

func getTrelloBoard(c *gin.Context) (trelloboard *models.TrelloBoard, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	return getTrelloBoardByID(c, id)
}

func getTrelloBoardByID(c *gin.Context, id uint) (trelloboard *models.TrelloBoard, err error) {
	// Get TrelloBoard from DB
	trelloboard, err = models.GetTrelloBoardByID(id)
	if err != nil {
		log.Error("Error retrieving TrelloBoard from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving TrelloBoard from DB: " + err.Error()})
		return
	}
	// Another check to verify the TrelloBoard was retrieved, id should not be 0
	if trelloboard.ID == 0 {
		err = errNotExist
		log.Error("Error trelloboard id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	return // success
}
