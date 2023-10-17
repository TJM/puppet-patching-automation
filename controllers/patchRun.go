package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/tjm/puppet-patching-automation/controllers/events"
	"github.com/tjm/puppet-patching-automation/controllers/puppet"
	"github.com/tjm/puppet-patching-automation/controllers/trelloapi"
	"github.com/tjm/puppet-patching-automation/functions"
	"github.com/tjm/puppet-patching-automation/models"
	"github.com/tjm/puppet-patching-automation/views"
)

var (
	errPatchRunRequired = errors.New("ERROR: Patch Window (patch_window) is required")
)

// GetPatchRunList endpoint (GET)
func GetPatchRunList(c *gin.Context) {
	runs := models.GetPatchRuns()
	data := gin.H{"status": "success", "patch_runs": runs}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "patchRun-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, runs.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// UpdatePatchRun (or create if id == new) endpoint (PUT)
// - PathParams: id
func UpdatePatchRun(c *gin.Context) {
	update := true
	run, err := getPatchRun(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	oldPatchWindow := run.PatchWindow
	if run.ID == 0 {
		update = false
	}
	err = c.Bind(run)
	if err != nil {
		if run.PatchWindow == "" { // Lets get a nicer error for patch_run missing
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": errPatchRunRequired.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}
	err = run.Save()
	if err != nil {
		log.Error("Error saving patchRun: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
	}
	if run.PatchWindow != oldPatchWindow {
		errors := puppet.GetInventoryForPatchRun(run)
		if len(errors) > 0 {
			errorStrings := getErrorStrings(errors)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "messages": errorStrings})
			return
		}
	}

	// Handle Linking Chat Rooms
	err = linkChatRoomsToPatchRun(c, run)
	if err != nil {
		// Error should already be output
		return
	}

	eventType := models.ActionPatchRunUpdated
	if !update {
		eventType = models.ActionPatchRunCreated
	}
	events.PatchRunEvent(c, run, models.NewEvent(eventType))

	data := gin.H{"status": "success", "patch_run_id": run.ID, "patch_run": run}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "patchRun-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// DeletePatchRun endpoint (DELETE)
// - PathParams: id
func DeletePatchRun(c *gin.Context) {
	run, err := getPatchRun(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	boards := run.GetTrelloBoards()
	for _, board := range boards {
		err = trelloapi.DeleteTrelloBoard(board)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
			return
		}
	}
	err = run.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	events.PatchRunEvent(c, run, models.NewEvent(models.ActionPatchRunDeleted))

	data := gin.H{"status": "success", "message": "Deleted"} // no patch_run_id, redirects to /patchRun/
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "patchRun-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// GetPatchRun endpoint (GET)
// - PathParams: id
func GetPatchRun(c *gin.Context) {
	run, err := getPatchRun(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	data := gin.H{"status": "success", "patch_run": run}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "patchRun-show.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, run.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// RunPuppetDBQuery endpoint - Re-Run Puppet Query (POST)
// - PathParams: id
func RunPuppetDBQuery(c *gin.Context) {
	run, err := getPatchRun(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	errors := puppet.GetInventoryForPatchRun(run)
	if len(errors) > 0 {
		errorStrings := getErrorStrings(errors)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "messages": errorStrings})
		return
	}

	events.PatchRunEvent(c, run, models.NewEvent(models.ActionPatchRunUpdated))

	data := gin.H{"status": "success", "patch_run_id": run.ID, "patch_run": run}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "patchRun-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// GetServerList endpoint
// PathParams: patchid
func GetServerList(c *gin.Context) {
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	apps := models.GetApplications(id)
	output := views.OutputServerList(apps)
	c.String(http.StatusOK, output)
}

// GetDetailedServerList endpoint
// PathParams: patchid
func GetDetailedServerList(c *gin.Context) {
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	run, err := getPatchRun(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	apps := models.GetApplications(id)
	output, _ := views.OutputServerCSV(apps)
	contentLength := int64(output.Len())
	fileName := functions.SanitizeFilename(run.Name+"-"+functions.FormatAsISO8601NoSpace(time.Now())+".csv", false)
	headers := map[string]string{
		"Content-Disposition": `attachment; filename="` + fileName + `"`,
	}

	//c.Data(http.StatusOK, "text/csv", output.Bytes())
	c.DataFromReader(http.StatusOK, contentLength, "text/csv", output, headers)
}

// GetAppsEnvs endpoint
// PathParams: patchid
func GetAppsEnvs(c *gin.Context) {
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	apps := models.GetApplications(id)
	output := views.OutputAppsEnvs(apps)
	c.String(http.StatusOK, output)
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getPatchRun will get the id from context and return job
func getPatchRun(c *gin.Context) (patchRun *models.PatchRun, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		if errors.Is(err, errIDLatest) {
			id, err = models.GetLatestPatchRunID()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) { // If there are no patch run, create a new one
					c.Redirect(http.StatusTemporaryRedirect, "/patchRun/new")
				}
				return
			}
			loc := fmt.Sprintf("/patchRun/%v", id)
			c.Redirect(http.StatusFound, loc)
			return nil, errIDLatest
		} else if errors.Is(err, errIDNew) {
			patchRun = models.NewPatchRun()
			err = nil
		}
		return
	}
	return getPatchRunByID(c, id)
}

// getPatchRunByID retrives the patchRun from the DB
func getPatchRunByID(c *gin.Context, id uint) (patchRun *models.PatchRun, err error) {
	// Get PatchRun from DB
	patchRun, err = models.GetPatchRunByID(id)
	if err != nil {
		log.Error("Error retrieving patchRun from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving patchRun from DB: " + err.Error()})
		return
	}
	// Another check to verify the patchRun was retrieved, id should not be 0
	if patchRun.ID == 0 {
		err = errNotExist
		log.Error("Error patchRun ID should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error patchRun ID should not be 0 (not found)"})
		return
	}
	return // success
}
