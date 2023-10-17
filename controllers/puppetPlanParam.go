package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

// // ListPuppetPlanParams endpoint (GET)
// func ListPuppetPlanParams(c *gin.Context) {
// 	data := gin.H{"status": "success", "plans": models.GetPuppetPlanParams()}
// 	// Add empty plan for html (for default values)
// 	htmlData := gin.H{"status": "success", "plan_params": models.GetPuppetPlanParams(), "plan": models.NewPuppetPlanParam()}
// 	c.Negotiate(http.StatusOK, gin.Negotiate{
// 		HTMLName: "puppetPlanParam-list.gohtml",
// 		Data:     data,
// 		HTMLData: htmlData,
// 		Offered:  formatAllSupported,
// 	})
// }

// // GetPuppetPlanParam endpoint (GET)
// // PathParams: id
// func GetPuppetPlanParam(c *gin.Context) {
// 	puppetPlanParam, err := getPuppetPlanParam(c)
// 	if err != nil {
// 		return // error has already been logged
// 	}
// 	data := gin.H{"status": "success", "plan_param": puppetPlanParam}
// 	c.Negotiate(http.StatusOK, gin.Negotiate{
// 		HTMLName: "puppetPlanParam-show.gohtml",
// 		Data:     data,
// 		Offered:  formatAllSupported,
// 	})
// }

// UpdatePuppetPlanParam endpoint (PUT)
// - PathParams: id
func UpdatePuppetPlanParam(c *gin.Context) {
	puppetPlanParam, err := getPuppetPlanParam(c)
	if err != nil {
		return // error has already been logged
	}
	// Bind fields submitted
	err = c.Bind(puppetPlanParam)
	if err != nil {
		log.Error("ERROR binding puppetPlanParam: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ERROR binding puppetPlanParam: " + err.Error()})
		return
	}
	err = puppetPlanParam.Save()
	if err != nil {
		log.Error("ERROR SAVING puppetPlanParam: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ERROR SAVING puppetPlanParam: " + err.Error()})
		return
	}
	data := gin.H{"status": "success", "plan_param": puppetPlanParam}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlanParam-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// DeletePuppetPlanParam endpoint (DELETE)
// - PathParams: id
func DeletePuppetPlanParam(c *gin.Context) {
	puppetPlanParam, err := getPuppetPlanParam(c)
	if err != nil {
		return // error has already been logged
	}
	err = puppetPlanParam.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "message": "Deleted"}
	htmlData := gin.H{
		"status":     "success",
		"message":    "Deleted",
		"plan_param": puppetPlanParam,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlanParam-success-redirect.gohtml",
		Data:     data,
		HTMLData: htmlData,
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getPuppetPlanParam will get the id from context and return job
func getPuppetPlanParam(c *gin.Context) (puppetPlanParam *models.PuppetPlanParam, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	// Get puppetPlanParam from DB
	puppetPlanParam, err = models.GetPuppetPlanParamByID(id)
	if err != nil {
		log.Error("Error retrieving PuppetPlanParam from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving PuppetPlanParam from DB: " + err.Error()})
		return
	}
	// Another check to verify the PuppetPlanParam was retrieved, id should not be 0
	if puppetPlanParam.ID == 0 {
		err = errNotExist
		log.Error("Error puppetPlanParam id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error puppetPlanParam id should not be 0 (not found)"})
		return
	}
	return // success
}
