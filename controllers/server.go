package controllers

import (
	"net/http"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/puppet"
	"github.com/tjm/puppet-patching-automation/models"
)

// GetAllServers endpoint (GET)
// PathParams: id
func GetAllServers(c *gin.Context) {
	component, err := getComponent(c)
	if err != nil {
		return
	}
	servers := component.GetServers()
	data := gin.H{
		"status":         "success",
		"environment_id": component.EnvironmentID,
		"component_id":   component.ID,
		"servers":        &servers,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "server-list.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// GetServer endpoint (GET)
// PathParams: id
func GetServer(c *gin.Context) {
	server, err := getServer(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "server": server})
}

// ServerRunPatching endpoint (POST)
// PathParams: id
func ServerRunPatching(c *gin.Context) {
	server, err := getServer(c)
	if err != nil {
		return
	}
	job, err := puppet.PatchServer(server, location.Get(c).String(), false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error PatchServer: " + err.Error()})
		return
	}
	data := gin.H{
		"status":  "success",
		"message": "Started Task on Puppet Server, see link for details.",
		"jobID":   job,
		"link":    job.ConsoleURL,
	}
	c.JSON(http.StatusOK, data)
	// c.Negotiate(http.StatusOK, gin.Negotiate{
	// 	HTMLName: "server-patch-redirect.gohtml",
	// 	Data:     data,
	// 	Offered:  formatAllSupported,
	// })
}

// GetServerFacts endpoint (GET)
// PathParams: id
func GetServerFacts(c *gin.Context) {
	server, err := getServer(c)
	if err != nil {
		return
	}
	facts, err := puppet.GetFacts(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error GetFacts: " + err.Error()})
		return
	}
	// data := gin.H{
	// 	"status": "success",
	// 	"facts":  facts,
	// }
	c.JSON(http.StatusOK, facts)
	// c.Negotiate(http.StatusOK, gin.Negotiate{
	// 	HTMLName: "server-patch-redirect.gohtml",
	// 	Data:     data,
	// 	Offered:  formatAllSupported,
	// })
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

func getServer(c *gin.Context) (server *models.Server, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	return getServerByID(c, id)
}

func getServerByID(c *gin.Context, id uint) (server *models.Server, err error) {
	// Get Server from DB
	server, err = models.GetServerByID(id)
	if err != nil {
		log.Error("Error retrieving Server from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving Server from DB: " + err.Error()})
		return
	}
	// Another check to verify the Server was retrieved, id should not be 0
	if server.ID == 0 {
		err = errNotExist
		log.Error("Error server id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	return // success
}
