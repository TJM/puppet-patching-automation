package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

// GetAllApplications for a patch run ID
// PathParams: patchid
func GetAllApplications(c *gin.Context) {
	patchRun, err := getPatchRun(c)
	if err != nil {
		return
	}
	apps := models.GetApplications(patchRun.ID)
	if len(apps) == 1 {
		loc := fmt.Sprintf("/application/%v/environments", apps[0].ID)
		c.Redirect(http.StatusTemporaryRedirect, loc)
		return
	}
	data := gin.H{
		"status":       "success",
		"applications": apps,
		"patchRun":     patchRun,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "application-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, apps.GetBreadCrumbs(patchRun), data),
		Offered:  formatAllSupported,
	})
}

// GetApplication endpoint (GET)
// PathParams: id
func GetApplication(c *gin.Context) {
	app, err := getApplication(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "application": app})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getApllication will get the id from context and return job
func getApplication(c *gin.Context) (app *models.Application, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		if errors.Is(err, errIDLatest) {
			id, err = models.GetLatestPatchRunID()
			if err != nil {
				return
			}
			loc := fmt.Sprintf("/patchRun/%v", id)
			c.Redirect(http.StatusFound, loc)
			return nil, errIDLatest
		}
		return
	}
	return getApplicationByID(c, id)
}

// getApllicationByID retrives the patchRun from the DB
func getApplicationByID(c *gin.Context, id uint) (app *models.Application, err error) {
	// Get PatchRun from DB
	app, err = models.GetApplicationByID(id)
	if err != nil {
		log.Error("Error retrieving patchRun from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving patchRun from DB: " + err.Error()})
		return
	}
	// Another check to verify the jenkinsServer was retrieved, id should not be 0
	if app.ID == 0 {
		err = errNotExist
		log.Error("Error patchRun ID should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error patchRun ID should not be 0 (not found)"})
		return
	}
	return // success
}
