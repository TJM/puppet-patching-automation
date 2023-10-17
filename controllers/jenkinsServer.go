package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/jenkinsapi"
	"github.com/tjm/puppet-patching-automation/models"
)

// ListJenkinsServers endpoint (GET)
func ListJenkinsServers(c *gin.Context) {
	jenkinsServers := models.GetJenkinsServers()
	for _, jenkinsServer := range jenkinsServers {
		censorJenkinsServerFields(jenkinsServer)
	}
	data := gin.H{"status": "success", "jenkins_servers": jenkinsServers}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsServer-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, jenkinsServers.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// GetJenkinsServer endpoint (GET)
// PathParams: id
func GetJenkinsServer(c *gin.Context) {
	jenkinsServer, err := getJenkinsServer(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	censorJenkinsServerFields(jenkinsServer)
	data := gin.H{"status": "success", "jenkins_server": jenkinsServer}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsServer-show.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, jenkinsServer.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// UpdateJenkinsServer endpoint (PUT/POST)
// - PathParams: id
func UpdateJenkinsServer(c *gin.Context) {
	jenkinsServer, err := getJenkinsServer(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	token := jenkinsServer.Token
	// Bind fields submitted
	err = c.Bind(jenkinsServer)
	if err != nil {
		log.Error("ERROR binding jenkinsServer: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	// Handle not updating token
	if jenkinsServer.Token == "" {
		jenkinsServer.Token = token
	}
	// Handle DISABLE (Enable not checked)
	if c.PostForm("Enabled") == "" {
		jenkinsServer.Enabled = false
	}
	// Handle SSL DISABLE (SSL not checked)
	if c.PostForm("SSL") == "" {
		jenkinsServer.SSL = false
	}
	// Handle SSLSkipVerify DISABLE (SSL not checked)
	if c.PostForm("SSLSkipVerify") == "" {
		jenkinsServer.SSLSkipVerify = false
	}
	jenkinsServer.Save()
	censorJenkinsServerFields(jenkinsServer)
	data := gin.H{"status": "success", "jenkins_server": jenkinsServer}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsServer-success-redirect.gohtml",
		Data:     data,
		HTMLData: gin.H{},
		Offered:  formatAllSupported,
	})
}

// DeleteJenkinsServer endpoint (DELETE)
// - PathParams: id
func DeleteJenkinsServer(c *gin.Context) {
	jenkinsServer, err := getJenkinsServer(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	err = jenkinsServer.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "message": "Deleted"} // no jenkins_server_id, redirects to /config/jenkinsServer/
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsServer-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// GetJenkinsServerInfo endpoint (GET)
// PathParams: id - Jenkins Server ID
func GetJenkinsServerInfo(c *gin.Context) {
	jenkinsServer, err := getJenkinsServer(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}

	info, err := jenkinsapi.Info(c, jenkinsServer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": info})
}

// GetJenkinsAPIJobs endpoint (GET)
// PathParams: id - Jenkins Server ID
// PathParams: path - path to jenkins job in API
func GetJenkinsAPIJobs(c *gin.Context) {
	path := c.Param("path") // TODO: validate path
	jenkinsServer, err := getJenkinsServer(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}

	jobs, err := jenkinsapi.GetJobsMetadata(c, jenkinsServer, path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{
		"status":           "success",
		"jenkins_api_jobs": jobs,
	}
	// This one doesn't follow our typical pattern because it is getting data directly from the Jenkins API instead of a model
	htmlData := getHTMLData(c, models.JenkinsJobs{}.GetBreadCrumbs(jenkinsServer), data,
		gin.H{
			"isFolder":       jenkinsapi.IsFolder,
			"jenkins_server": jenkinsServer,
			"path":           path,
		})
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsServer-APIJobs.gohtml",
		Data:     data,
		HTMLData: htmlData,
		Offered:  formatAllSupported,
	})
}

// AddJenkinsJobToServer endpoint (POST)
// PathParams: id - Jenkins Server ID
// PathParams: path - path to jenkins job in API
func AddJenkinsJobToServer(c *gin.Context) {
	jenkinsServer, err := getJenkinsServer(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	jenkinsJob, err := jenkinsapi.AddJob(c, jenkinsServer, c.Param("path"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "jenkins_job": jenkinsJob}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsJob-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// GetJenkinsJobs endpoint (GET)
// PathParams: id - Jenkins Server ID
func GetJenkinsJobs(c *gin.Context) {
	jenkinsServer, err := getJenkinsServer(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	jobs, err := jenkinsServer.GetJobs()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{
		"status":         "success",
		"jenkins_jobs":   jobs,
		"jenkins_server": jenkinsServer,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "jenkinsJob-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, jobs.GetBreadCrumbs(jenkinsServer), data),
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

func getJenkinsServer(c *gin.Context) (jenkinsServer *models.JenkinsServer, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		if errors.Is(err, errIDNew) {
			jenkinsServer = models.NewJenkinsServer()
			err = nil
		}
		return
	}
	return getJenkinsServerByID(c, id)
}

func getJenkinsServerByID(c *gin.Context, id uint) (jenkinsServer *models.JenkinsServer, err error) {
	// Get jenkinsServer from DB
	jenkinsServer, err = models.GetJenkinsServerByID(id)
	if err != nil {
		log.Error("Error retrieving jenkinsServer from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving jenkinsServer from DB: " + err.Error()})
		return
	}
	// Another check to verify the jenkinsServer was retrieved, id should not be 0
	if jenkinsServer.ID == 0 {
		err = errNotExist
		log.Error("Error jenkinsServer id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error jenkinsServer id should not be 0 (not found)"})
		return
	}
	return // success
}

func censorJenkinsServerFields(jenkinsServer *models.JenkinsServer) {
	// Censor Token parameter (sensitive)
	if jenkinsServer.Token != "" {
		jenkinsServer.Token = "[censored]"
	}
}
