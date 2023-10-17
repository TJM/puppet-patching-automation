package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/jenkinsapi"
	"github.com/tjm/puppet-patching-automation/models"
)

// ListJenkinsJobs endpoint (GET)
func ListJenkinsJobs(c *gin.Context) {
	jobs := models.GetJenkinsJobs()
	data := gin.H{"status": "success", "jenkins_jobs": jobs}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsJob-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, jobs.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// GetJenkinsJob endpoint (GET)
// PathParams: id
func GetJenkinsJob(c *gin.Context) {
	jenkinsJob, err := getJenkinsJob(c)
	if err != nil {
		return // error has already been logged
	}
	data := gin.H{"status": "success", "jenkins_job": jenkinsJob}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsJob-show.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, jenkinsJob.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// UpdateJenkinsJob endpoint (PUT)
// - PathParams: id
func UpdateJenkinsJob(c *gin.Context) {
	jenkinsJob, err := getJenkinsJob(c)
	if err != nil {
		return // error has already been logged
	}
	// Bind fields submitted
	err = c.Bind(jenkinsJob)
	if err != nil {
		log.Error("ERROR binding jenkinsJob: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ERROR binding jenkinsJob: " + err.Error()})
		return
	}
	// Handle Checkboxes not checked
	if c.PostForm("Enabled") == "" {
		jenkinsJob.Enabled = false
	}
	if c.PostForm("IsForPatchRun") == "" {
		jenkinsJob.IsForPatchRun = false
	}
	if c.PostForm("IsForApplication") == "" {
		jenkinsJob.IsForApplication = false
	}
	if c.PostForm("IsForServer") == "" {
		jenkinsJob.IsForServer = false
	}
	err = jenkinsJob.Save()
	if err != nil {
		log.Error("Error saving jenkinsJob: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
	}
	data := gin.H{"status": "success", "jenkins_job": jenkinsJob}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsJob-success-redirect.gohtml",
		Data:     data,
		HTMLData: gin.H{},
		Offered:  formatAllSupported,
	})
}

// DeleteJenkinsJob endpoint (DELETE)
// - PathParams: id
func DeleteJenkinsJob(c *gin.Context) {
	jenkinsJob, err := getJenkinsJob(c)
	if err != nil {
		return // error has already been logged
	}
	err = jenkinsJob.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "message": "Deleted"} // no jenkins_job_id, redirects to /config/jenkinsJob/
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsJob-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// UpdateJenkinsJobFromAPI endpoint (POST)
// - PathParams: id
func UpdateJenkinsJobFromAPI(c *gin.Context) {
	jenkinsJob, err := getJenkinsJob(c)
	if err != nil {
		return // error has already been logged
	}
	jenkinsServer, err := getJenkinsServerByID(c, jenkinsJob.JenkinsServerID)
	if err != nil {
		return // error has already been logged
	}
	err = jenkinsapi.UpdateDBJobFromAPIJob(c, jenkinsServer, jenkinsJob)
	if err != nil {
		log.Error("Error updating DB JenkinsJob from API: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error updating DB JenkinsJob from API: " + err.Error()})
		return
	}
	data := gin.H{"status": "success", "jenkins_job": jenkinsJob}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsJob-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getJenkinsJob will get the id from context and return job
func getJenkinsJob(c *gin.Context) (jenkinsJob *models.JenkinsJob, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	return getJenkinsJobByID(c, id)
}

// getJenkinsJobByID retrives the jenkinsJob from the DB
func getJenkinsJobByID(c *gin.Context, id uint) (jenkinsJob *models.JenkinsJob, err error) {
	// Get jenkinsJob from DB
	jenkinsJob, err = models.GetJenkinsJobByID(id)
	if err != nil {
		log.Error("Error retrieving jenkinsServer from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving jenkinsServer from DB: " + err.Error()})
		return
	}
	// Another check to verify the jenkinsServer was retrieved, id should not be 0
	if jenkinsJob.ID == 0 {
		err = errNotExist
		log.Error("Error jenkinsJob id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error jenkinsJob id should not be 0 (not found)"})
		return
	}
	return // success
}
