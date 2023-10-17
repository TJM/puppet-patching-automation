package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/puppet"
	"github.com/tjm/puppet-patching-automation/models"
)

// ListPuppetTasks endpoint (GET)
func ListPuppetTasks(c *gin.Context) {
	tasks := models.GetPuppetTasks()
	data := gin.H{"status": "success", "tasks": tasks}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTask-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, tasks.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// GetPuppetTask endpoint (GET)
// PathParams: id
func GetPuppetTask(c *gin.Context) {
	puppetTask, err := getPuppetTask(c)
	if err != nil {
		return // error has already been logged
	}
	data := gin.H{"status": "success", "task": puppetTask}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTask-show.gohtml",
		HTMLData: getHTMLData(c, puppetTask.GetBreadCrumbs(), data),
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// UpdatePuppetTask endpoint (PUT)
// - PathParams: id
func UpdatePuppetTask(c *gin.Context) {
	puppetTask, err := getPuppetTask(c)
	if err != nil {
		return // error has already been logged
	}
	// Bind fields submitted
	err = c.Bind(puppetTask)
	if err != nil {
		log.Error("ERROR binding puppetTask: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ERROR binding puppetTask: " + err.Error()})
		return
	}
	// Handle Checkboxes not checked
	if c.PostForm("Enabled") == "" {
		puppetTask.Enabled = false
	}
	if c.PostForm("IsForPatchRun") == "" {
		puppetTask.IsForPatchRun = false
	}
	if c.PostForm("IsForApplication") == "" {
		puppetTask.IsForApplication = false
	}
	if c.PostForm("IsForComponent") == "" {
		puppetTask.IsForComponent = false
	}
	if c.PostForm("IsForServer") == "" {
		puppetTask.IsForServer = false
	}
	err = puppetTask.Save()
	if err != nil {
		log.Error("Error saving puppetTask: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
	}
	data := gin.H{"status": "success", "task": puppetTask}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTask-success-redirect.gohtml",
		Data:     data,
		HTMLData: gin.H{},
		Offered:  formatAllSupported,
	})
}

// DeletePuppetTask endpoint (DELETE)
// - PathParams: id
func DeletePuppetTask(c *gin.Context) {
	puppetTask, err := getPuppetTask(c)
	if err != nil {
		return // error has already been logged
	}
	err = puppetTask.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "message": "Deleted"} // no task_id, redirects to /config/puppetTask/
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTask-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// UpdatePuppetTaskFromAPI endpoint (POST)
// - PathParams: id
func UpdatePuppetTaskFromAPI(c *gin.Context) {
	puppetTask, err := getPuppetTask(c)
	if err != nil {
		return // error has already been logged
	}
	puppetServers, err := puppetTask.GetPuppetServers()
	if err != nil {
		return // error has already been logged
	}
	puppetServer := puppetServers[0] // TODO: Decide whether to check multiple servers, which server or all? use first server for now
	err = puppet.UpdateTaskDetails(puppetServer, puppetTask)
	if err != nil {
		log.Error("Error updating DB PuppetTask from API: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error updating DB PuppetTask from API: " + err.Error()})
		return
	}
	data := gin.H{"status": "success", "task": puppetTask}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTask-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getPuppetTask will get the id from context and return job
func getPuppetTask(c *gin.Context) (puppetTask *models.PuppetTask, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	return getPuppetTaskByID(c, id)
}

// getPuppetTaskByID retrives the puppetTask from the DB
func getPuppetTaskByID(c *gin.Context, id uint) (puppetTask *models.PuppetTask, err error) {
	// Get puppetTask from DB
	puppetTask, err = models.GetPuppetTaskByID(id)
	if err != nil {
		log.Error("Error retrieving puppetServer from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving puppetServer from DB: " + err.Error()})
		return
	}
	// Another check to verify the puppetServer was retrieved, id should not be 0
	if puppetTask.ID == 0 {
		err = errNotExist
		log.Error("Error puppetTask id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error puppetTask id should not be 0 (not found)"})
		return
	}
	return // success
}
