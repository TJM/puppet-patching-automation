package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

// GetAllEnvironments endpoint (GET)
// PathParams: id
func GetAllEnvironments(c *gin.Context) {
	app, err := getApplication(c)
	if err != nil {
		return
	}
	envs := app.GetEnvironments()
	if len(envs) == 1 { // If there is only one environment, redirect there
		loc := fmt.Sprintf("/environment/%v/components", envs[0].ID)
		c.Redirect(http.StatusTemporaryRedirect, loc)
		return
	}
	data := gin.H{
		"status":       "success",
		"environments": envs,
		"application":  app,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "environment-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, envs.GetBreadCrumbs(app), data),
		Offered:  formatAllSupported,
	})
}

// GetEnvironment endpoint (GET)
// PathParams: id
func GetEnvironment(c *gin.Context) {
	env, err := getEnvironment(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "env_id": env.ID, "environment": env})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

func getEnvironment(c *gin.Context) (environment *models.Environment, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	return getEnvironmentByID(c, id)
}

func getEnvironmentByID(c *gin.Context, id uint) (environment *models.Environment, err error) {
	// Get jenkinsServer from DB
	environment, err = models.GetEnvironmentByID(id)
	if err != nil {
		log.Error("Error retrieving jenkinsServer from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving jenkinsServer from DB: " + err.Error()})
		return
	}
	// Another check to verify the jenkinsServer was retrieved, id should not be 0
	if environment.ID == 0 {
		err = errNotExist
		log.Error("Error server id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	return // success
}
