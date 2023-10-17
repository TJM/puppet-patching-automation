package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/puppet"
	"github.com/tjm/puppet-patching-automation/models"
)

// ListPuppetServers endpoint (GET)
func ListPuppetServers(c *gin.Context) {
	puppetServers := models.GetPuppetServers()
	for _, puppetServer := range puppetServers {
		censorPuppetServerFields(puppetServer)
	}
	data := gin.H{"status": "success", "puppet_servers": puppetServers}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetserver-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, puppetServers.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// GetPuppetServer endpoint (GET)
// PathParams: id
func GetPuppetServer(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	censorPuppetServerFields(puppetServer)
	data := gin.H{"status": "success", "puppet_server": puppetServer}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		Offered:  formatAllSupported,
		HTMLName: "puppetserver-show.gohtml",
		HTMLData: getHTMLData(c, puppetServer.GetBreadCrumbs(), data),
		JSONData: data,
		XMLData:  data,
		YAMLData: data,
		Data:     data,
	})
}

// UpdatePuppetServer endpoint (PUT)
// - PathParams: id
func UpdatePuppetServer(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	token := puppetServer.Token
	err = c.Bind(puppetServer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	// Handle not updating token
	if puppetServer.Token == "" {
		puppetServer.Token = token
	}
	// Handle DISABLE (Enable not checked)
	if c.PostForm("Enabled") == "" {
		puppetServer.Enabled = false
	}
	// Handle SSL DISABLE (SSL not checked)
	if c.PostForm("SSL") == "" {
		puppetServer.SSL = false
	}
	// Handle SSLSkipVerify DISABLE (SSL not checked)
	if c.PostForm("SSLSkipVerify") == "" {
		puppetServer.SSLSkipVerify = false
	}
	puppetServer.Save()
	censorPuppetServerFields(puppetServer)
	data := gin.H{"status": "success", "puppet_server": puppetServer}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "success-redirect-puppetserver.gohtml",
		Data:     data,
		HTMLData: gin.H{},
		Offered:  formatAllSupported,
	})
}

// DeletePuppetServer endpoint (DELETE)
// - PathParams: id
func DeletePuppetServer(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	censorPuppetServerFields(puppetServer)
	err = puppetServer.Delete(true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "message": "Deleted"} // no puppet_server_id, redirects to /config/puppetServer/
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "success-redirect-puppetserver.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// GetPuppetServerAPIPlans endpoint (GET)
//
//	PathParams:
//	  - id (PuppetServerID)
//	  - env (Puppet Environment)
func GetPuppetServerAPIPlans(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}

	env := c.Param("env")

	plans, err := puppet.GetPlans(puppetServer, env)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error Getting Plans: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// GetPuppetServerAPIPlan endpoint (GET)
//   - PathParams: id (PuppetServerID)
//     env (Puppet Environment)
//     module (Puppet Module Name)
//     taskName (Puppet Task Name)
func GetPuppetServerAPIPlan(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	// TODO: Validate Parameters
	env := c.Param("env")
	module := c.Param("module")
	planName := c.Param("planName")

	plan, err := puppet.GetPlan(puppetServer, env, module, planName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error Getting Tasks: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

// GetPuppetServerAPITasks endpoint (GET)
//   - PathParams: id (PuppetServerID)
//     env (Puppet Environment)
func GetPuppetServerAPITasks(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}

	env := c.Param("env")

	tasks, err := puppet.GetTasks(puppetServer, env)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error Getting Tasks: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// GetPuppetServerAPITask endpoint (GET)
//   - PathParams: id (PuppetServerID)
//     env (Puppet Environment)
//     module (Puppet Module Name)
//     taskName (Puppet Task Name)
func GetPuppetServerAPITask(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}

	// TODO: Validate Parameters
	env := c.Param("env")
	module := c.Param("module")
	taskName := c.Param("taskName")

	task, err := puppet.GetTask(puppetServer, env, module, taskName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error Getting Tasks: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// GetPuppetServerJob endpoint (GET)
//   - PathParams: id (PuppetServerID)
//     jobID (Job ID)
func GetPuppetServerJob(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}

	jobID := c.Param("jobID")

	job, err := puppet.GetJobByID(puppetServer, jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error Getting Job: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"job": job})
}

// GetPuppetServerJobReport endpoint (GET)
//   - PathParams: id (PuppetServerID)
//     jobID (Job ID)
func GetPuppetServerJobReport(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}

	jobID := c.Param("jobID")

	jobReport, err := puppet.GetJobReportByID(puppetServer, jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error Getting JobReport: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": jobReport})
}

// GetPuppetServerEnvironments endpoint (GET)
// - PathParams: id (PuppetServerID)
func GetPuppetServerEnvironments(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	envs, err := puppet.PDBEnvironments(puppetServer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error Getting Environments: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"environments": envs})
}

// GetPuppetServerEnvironmentsPE endpoint (GET)
// - PathParams: id (PuppetServerID)
func GetPuppetServerEnvironmentsPE(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	envs, err := puppet.PEEnvironments(puppetServer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error Getting Environments: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"environments": envs})
}

// AddPuppetTaskToServer endpoint (POST)
// PathParams: id - Puppet Server ID
func AddPuppetTaskToServer(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	task := models.NewPuppetTask()
	err = c.Bind(task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	err = puppet.UpdateTaskDetails(puppetServer, task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "task": task}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTask-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// GetPuppetServerTasks endpoint (GET)
// PathParams: id - Puppet Server ID
func GetPuppetServerTasks(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	tasks, err := puppetServer.GetTasks()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	apiTasks, err := puppet.GetTasks(puppetServer, "production") // TODO: Determine whether env should be selectable
	if err != nil {
		log.WithField("puppetServerID", puppetServer.ID).Error("ERROR retrieving puppetTasks from API", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	availableTaskNames := make([]string, 0)
	for _, item := range apiTasks.Items {
		availableTaskNames = append(availableTaskNames, item.Name)
	}
	unassociatedTasks, err := puppetServer.GetUnassociatedTasks(availableTaskNames)
	if err != nil {
		log.WithField("puppetServerID", puppetServer.ID).Error("ERROR retrieving unassociated PuppetTasks from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	data := gin.H{
		"status":            "success",
		"tasks":             tasks,
		"unassociatedTasks": unassociatedTasks,
		"puppet_server":     puppetServer,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTask-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, tasks.GetBreadCrumbs(puppetServer), data),
		Offered:  formatAllSupported,
	})
}

// AddPuppetPlanToServer endpoint (POST)
// PathParams: id - Puppet Server ID
func AddPuppetPlanToServer(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	plan := models.NewPuppetPlan()
	err = c.Bind(plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	err = puppet.UpdatePlanDetails(puppetServer, plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "plan": plan}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlan-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// GetPuppetServerPlans endpoint (GET)
// PathParams: id - Puppet Server ID
func GetPuppetServerPlans(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	plans, err := puppetServer.GetPlans()
	if err != nil {
		// NOTE: "Could not find route /orchestrator/v1/plans" means the puppet server is too old to support plans via orchestrator API
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	apiPlans, err := puppet.GetPlans(puppetServer, "production") // TODO: Determine whether env should be selectable
	if err != nil {
		// NOTE: "Could not find route /orchestrator/v1/plans" means the puppet server is too old to support plans via orchestrator API
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	availablePlanNames := make([]string, 0)
	for _, item := range apiPlans.Items {
		availablePlanNames = append(availablePlanNames, item.Name)
	}
	unassociatedPlans, err := puppetServer.GetUnassociatedPlans(availablePlanNames)
	if err != nil {
		log.WithField("puppetServerID", puppetServer.ID).Error("ERROR retrieving puppetPlans from DB:", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	data := gin.H{
		"status":            "success",
		"plans":             plans,
		"unassociatedPlans": unassociatedPlans,
		"puppet_server":     puppetServer,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlan-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, plans.GetBreadCrumbs(puppetServer), data),
		Offered:  formatAllSupported,
	})
}

// AssociatePuppetPlanToServer : Associates a puppet plan (submitted ID) to the puppetServer in the path
func AssociatePuppetPlanToServer(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	plan := models.NewPuppetPlan()
	err = c.Bind(plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	err = puppetServer.AddPuppetPlan(plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "plan": plan, "puppet_server": puppetServer}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlan-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// DisassociatePuppetPlanFromServer : Disassociates a puppet plan (submitted ID) from the puppetServer in the path
func DisassociatePuppetPlanFromServer(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	plan := models.NewPuppetPlan()
	err = c.Bind(plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	err = puppetServer.RemovePuppetPlan(plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "plan": plan, "puppet_server": puppetServer}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetPlan-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// AssociatePuppetTaskToServer : Associates a PuppetTask (submitted ID) to the PuppetServer in the path
func AssociatePuppetTaskToServer(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	task := models.NewPuppetTask()
	err = c.Bind(task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	err = puppetServer.AddPuppetTask(task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "task": task, "puppet_server": puppetServer}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTask-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// DisassociatePuppetTaskFromServer : Disassociates a puppet task (submitted ID) from the puppetServer in the path
func DisassociatePuppetTaskFromServer(c *gin.Context) {
	puppetServer, err := getPuppetServer(c)
	if err != nil {
		return
	}
	task := models.NewPuppetTask()
	err = c.Bind(task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	err = puppetServer.RemovePuppetTask(task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	data := gin.H{"status": "success", "task": task, "puppet_server": puppetServer}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "puppetTask-success-redirect.gohtml",
		Data:     data,
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

func getPuppetServer(c *gin.Context) (puppetServer *models.PuppetServer, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		if errors.Is(err, errIDNew) {
			puppetServer = models.NewPuppetServer()
			err = nil
		}
		return
	}
	return getPuppetServerByID(c, id)
}

func getPuppetServerByID(c *gin.Context, id uint) (puppetServer *models.PuppetServer, err error) {
	// Get puppetServer from DB
	puppetServer, err = models.GetPuppetServerByID(id)
	if err != nil {
		log.Error("Error retrieving puppetServer from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving puppetServer from DB: " + err.Error()})
		return
	}
	// Another check to verify the puppetServer was retrieved, id should not be 0
	if puppetServer.ID == 0 {
		err = errNotExist
		log.Error("Error server id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	return // success
}

func censorPuppetServerFields(puppetServer *models.PuppetServer) {
	// Censor Token parameter (sensitive)
	if puppetServer.Token != "" {
		puppetServer.Token = "[censored]"
	}
}
