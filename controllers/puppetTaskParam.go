package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

// // ListPuppetTaskParams endpoint (GET)
// func ListPuppetTaskParams(c *gin.Context) {
// 	data := gin.H{"status": "success", "tasks": models.GetPuppetTaskParams()}
// 	// Add empty task for html (for default values)
// 	htmlData := gin.H{"status": "success", "task_params": models.GetPuppetTaskParams(), "task": models.NewPuppetTaskParam()}
// 	c.Negotiate(http.StatusOK, gin.Negotiate{
// 		HTMLName: "puppetTaskParam-list.gohtml",
// 		Data:     data,
// 		HTMLData: htmlData,
// 		Offered:  formatAllSupported,
// 	})
// }

// // GetPuppetTaskParam endpoint (GET)
// // PathParams: id
// func GetPuppetTaskParam(c *gin.Context) {
// 	puppetTaskParam, err := getPuppetTaskParam(c)
// 	if err != nil {
// 		return // error has already been logged
// 	}
// 	data := gin.H{"status": "success", "task_param": puppetTaskParam}
// 	c.Negotiate(http.StatusOK, gin.Negotiate{
// 		HTMLName: "puppetTaskParam-show.gohtml",
// 		Data:     data,
// 		Offered:  formatAllSupported,
// 	})
// }

// UpdatePuppetTaskParam endpoint (PUT)
// - PathParams: id
func UpdatePuppetTaskParam(c *gin.Context) {
	puppetTaskParam, err := getPuppetTaskParam(c)
	if err != nil {
		return // error has already been logged
	}
	// Bind fields submitted
	err = c.Bind(puppetTaskParam)
	if err != nil {
		log.Error("ERROR binding puppetTaskParam: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ERROR binding puppetTaskParam: " + err.Error()})
		return
	}
	err = puppetTaskParam.Save()
	if err != nil {
		log.Error("ERROR SAVING puppetTaskParam: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ERROR SAVING puppetTaskParam: " + err.Error()})
		return
	}
	data := gin.H{"status": "success", "task_param": puppetTaskParam}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTaskParam-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// DeletePuppetTaskParam endpoint (DELETE)
// - PathParams: id
func DeletePuppetTaskParam(c *gin.Context) {
	puppetTaskParam, err := getPuppetTaskParam(c)
	if err != nil {
		return // error has already been logged
	}
	err = puppetTaskParam.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "message": "Deleted"}
	htmlData := gin.H{
		"status":     "success",
		"message":    "Deleted",
		"task_param": puppetTaskParam,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTaskParam-success-redirect.gohtml",
		Data:     data,
		HTMLData: htmlData,
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getPuppetTaskParam will get the id from context and return job
func getPuppetTaskParam(c *gin.Context) (puppetTaskParam *models.PuppetTaskParam, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	// Get puppetTaskParam from DB
	puppetTaskParam, err = models.GetPuppetTaskParamByID(id)
	if err != nil {
		log.Error("Error retrieving PuppetTaskParam from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving PuppetTaskParam from DB: " + err.Error()})
		return
	}
	// Another check to verify the PuppetTaskParam was retrieved, id should not be 0
	if puppetTaskParam.ID == 0 {
		err = errNotExist
		log.Error("Error puppetTaskParam id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error puppetTaskParam id should not be 0 (not found)"})
		return
	}
	return // success
}
