package controllers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

// // ListJenkinsJobParams endpoint (GET)
// func ListJenkinsJobParams(c *gin.Context) {
// 	data := gin.H{"status": "success", "jenkins_jobs": models.GetJenkinsJobParams()}
// 	// Add empty jenkins_job for html (for default values)
// 	htmlData := gin.H{"status": "success", "jenkins_job_params": models.GetJenkinsJobParams(), "jenkins_job": models.NewJenkinsJobParam()}
// 	c.Negotiate(http.StatusOK, gin.Negotiate{
// 		HTMLName: "jenkinsJobParam-list.gohtml",
// 		Data:     data,
// 		HTMLData: htmlData,
// 		Offered:  formatAllSupported,
// 	})
// }

// // GetJenkinsJobParam endpoint (GET)
// // PathParams: id
// func GetJenkinsJobParam(c *gin.Context) {
// 	jenkinsJobParam, err := getJenkinsJobParam(c)
// 	if err != nil {
// 		return // error has already been logged
// 	}
// 	data := gin.H{"status": "success", "jenkins_job_param": jenkinsJobParam}
// 	c.Negotiate(http.StatusOK, gin.Negotiate{
// 		HTMLName: "jenkinsJobParam-show.gohtml",
// 		Data:     data,
// 		Offered:  formatAllSupported,
// 	})
// }

// UpdateJenkinsJobParam endpoint (PUT)
// - PathParams: id
func UpdateJenkinsJobParam(c *gin.Context) {
	jenkinsJobParam, err := getJenkinsJobParam(c)
	if err != nil {
		return // error has already been logged
	}
	// Bind fields submitted
	err = c.Bind(jenkinsJobParam)
	if err != nil {
		log.Error("ERROR binding jenkinsJobParam: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ERROR binding jenkinsJobParam: " + err.Error(),
		})
		return
	}
	err = jenkinsJobParam.Save()
	if err != nil {
		log.Error("ERROR SAVING jenkinsJobParam: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ERROR SAVING jenkinsJobParam: " + err.Error(),
		})
		return
	}
	data := gin.H{"status": "success", "jenkins_job_param": jenkinsJobParam}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsJobParam-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// DeleteJenkinsJobParam endpoint (DELETE)
// - PathParams: id
func DeleteJenkinsJobParam(c *gin.Context) {
	jenkinsJobParam, err := getJenkinsJobParam(c)
	if err != nil {
		return // error has already been logged
	}
	err = jenkinsJobParam.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "message": "Deleted"}
	htmlData := gin.H{
		"session":           sessions.Default(c),
		"status":            "success",
		"message":           "Deleted",
		"jenkins_job_param": jenkinsJobParam,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsJobParam-success-redirect.gohtml",
		Data:     data,
		HTMLData: htmlData,
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getJenkinsJobParam will get the id from context and return job
func getJenkinsJobParam(c *gin.Context) (jenkinsJobParam *models.JenkinsJobParam, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	// Get jenkinsJobParam from DB
	jenkinsJobParam, err = models.GetJenkinsJobParamByID(id)
	if err != nil {
		log.Error("Error retrieving jenkinsServer from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving jenkinsServer from DB: " + err.Error()})
		return
	}
	// Another check to verify the jenkinsServer was retrieved, id should not be 0
	if jenkinsJobParam.ID == 0 {
		err = errNotExist
		log.Error("Error jenkinsJobParam id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error jenkinsJobParam id should not be 0 (not found)"})
		return
	}
	return // success
}
