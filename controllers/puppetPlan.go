package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/puppet"
	"github.com/tjm/puppet-patching-automation/models"
)

// ListPuppetPlans endpoint (GET)
func ListPuppetPlans(c *gin.Context) {
	plans := models.GetPuppetPlans()
	data := gin.H{"status": "success", "plans": plans}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlan-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, plans.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// GetPuppetPlan endpoint (GET)
// PathParams: id
func GetPuppetPlan(c *gin.Context) {
	puppetPlan, err := getPuppetPlan(c)
	if err != nil {
		return // error has already been logged
	}
	data := gin.H{"status": "success", "plan": puppetPlan}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlan-show.gohtml",
		HTMLData: getHTMLData(c, puppetPlan.GetBreadCrumbs(), data),
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// UpdatePuppetPlan endpoint (PUT)
// - PathParams: id
func UpdatePuppetPlan(c *gin.Context) {
	puppetPlan, err := getPuppetPlan(c)
	if err != nil {
		return // error has already been logged
	}
	// Bind fields submitted
	err = c.Bind(puppetPlan)
	if err != nil {
		log.Error("ERROR binding puppetPlan: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ERROR binding puppetPlan: " + err.Error()})
		return
	}
	// Handle Checkboxes not checked
	if c.PostForm("Enabled") == "" {
		puppetPlan.Enabled = false
	}
	if c.PostForm("IsForPatchRun") == "" {
		puppetPlan.IsForPatchRun = false
	}
	if c.PostForm("IsForApplication") == "" {
		puppetPlan.IsForApplication = false
	}
	if c.PostForm("IsForComponent") == "" {
		puppetPlan.IsForComponent = false
	}
	if c.PostForm("IsForServer") == "" {
		puppetPlan.IsForServer = false
	}
	err = puppetPlan.Save()
	if err != nil {
		log.Error("Error saving puppetPlan: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
	}
	data := gin.H{"status": "success", "plan": puppetPlan}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlan-success-redirect.gohtml",
		Data:     data,
		HTMLData: gin.H{},
		Offered:  formatAllSupported,
	})
}

// DeletePuppetPlan endpoint (DELETE)
// - PathParams: id
func DeletePuppetPlan(c *gin.Context) {
	puppetPlan, err := getPuppetPlan(c)
	if err != nil {
		return // error has already been logged
	}
	err = puppetPlan.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "message": "Deleted"} // no plan_id, redirects to /config/puppetPlan/
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlan-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// UpdatePuppetPlanFromAPI endpoint (POST)
// - PathParams: id
func UpdatePuppetPlanFromAPI(c *gin.Context) {
	puppetPlan, err := getPuppetPlan(c)
	if err != nil {
		return // error has already been logged
	}
	puppetServers, err := puppetPlan.GetPuppetServers()
	if err != nil {
		return // error has already been logged
	}
	puppetServer := puppetServers[0] // TODO: Decide whether to check multiple servers, which server or all? use first server for now
	err = puppet.UpdatePlanDetails(puppetServer, puppetPlan)
	if err != nil {
		log.Error("Error updating DB PuppetPlan from API: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error updating DB PuppetPlan from API: " + err.Error()})
		return
	}
	data := gin.H{"status": "success", "plan": puppetPlan}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlan-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getPuppetPlan will get the id from context and return job
func getPuppetPlan(c *gin.Context) (puppetPlan *models.PuppetPlan, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	return getPuppetPlanByID(c, id)
}

// getPuppetPlanByID retrives the puppetPlan from the DB
func getPuppetPlanByID(c *gin.Context, id uint) (puppetPlan *models.PuppetPlan, err error) {
	// Get puppetPlan from DB
	puppetPlan, err = models.GetPuppetPlanByID(id)
	if err != nil {
		log.Error("Error retrieving puppetServer from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving puppetServer from DB: " + err.Error()})
		return
	}
	// Another check to verify the puppetServer was retrieved, id should not be 0
	if puppetPlan.ID == 0 {
		err = errNotExist
		log.Error("Error puppetPlan id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error puppetPlan id should not be 0 (not found)"})
		return
	}
	return // success
}
