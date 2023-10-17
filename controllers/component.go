package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/puppet"
	"github.com/tjm/puppet-patching-automation/models"
)

// GetAllComponents endpoint (GET)
// PathParams: id - environment ID
func GetAllComponents(c *gin.Context) {
	env, err := getEnvironment(c)
	if err != nil {
		return
	}

	app, err := getApplicationByID(c, env.ApplicationID)
	if err != nil {
		return
	}

	components := env.GetComponentsAndServers()
	data := gin.H{
		"status":      "success",
		"application": app,
		"environment": env,
		"components":  components,
	}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "component-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, components.GetBreadCrumbs(env), data),
		Offered:  formatAllSupported,
	})
}

// GetComponent by ID endpoint (GET)
// PathParams: id - Component ID
func GetComponent(c *gin.Context) {
	component, err := getComponent(c)
	if err != nil {
		return
	}
	//servers := component.GetServers()
	c.JSON(http.StatusOK, gin.H{"status": "success", "component_id": component.ID, "component": component})
	// data := gin.H{"status": "success", "Component": component.ID, "component_id": id, "servers": &servers}
	// c.Negotiate(http.StatusOK, gin.Negotiate{
	// 	HTMLName: "server-list.gohtml",
	// 	Data:     data,
	// 	Offered:  formatAllSupported,
	// })
}

// ComponentRunPatching endpoint (POST)
// PathParams: id
func ComponentRunPatching(c *gin.Context) {
	component, err := getComponent(c)
	if err != nil {
		return
	}
	if component.HealthCheckScript == "" || component.HealthCheckScript == "UNSET" {
		err = errors.New("HealthCheckScript is required to execute ComponentPatching")
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error PatchComponent: " + err.Error()})
		return
	}
	baseURL := fmt.Sprintf("%s/component/%v", location.Get(c).String(), component.ID)
	jobs, err := puppet.PatchComponent(component, baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error PatchComponent: " + err.Error()})
		return
	}
	// TODO: Create a proper UI, just dump JSON for now...
	data := gin.H{
		"status":  "success",
		"message": "Started Task(s) on Puppet Server, see link(s) for details.",
		"jobs":    jobs,
	}
	c.JSON(http.StatusOK, data)
	// c.Negotiate(http.StatusOK, gin.Negotiate{
	// 	HTMLName: "server-patch-redirect.gohtml",
	// 	Data:     data,
	// 	Offered:  formatAllSupported,
	// })
}

// ComponentRunPuppetPlan runs a PuppetPlan against a Component (POST), preview a run (GET)
// Path Params:
// - id - Component.ID
// - planID - puppetPlan.ID
func ComponentRunPuppetPlan(c *gin.Context) {
	var htmlTemplate string
	var data gin.H

	// Component
	component, err := getComponent(c)
	if err != nil {
		return
	}

	// Puppet Server
	puppetServerID, err := validateID(c, "puppetServerID")
	if err != nil {
		return
	}
	puppetServer, err := getPuppetServerByID(c, puppetServerID)
	if err != nil {
		return
	}

	// PuppetPlan
	planID, err := validateID(c, "planID")
	if err != nil {
		return
	}
	puppetPlan, err := getPuppetPlanByID(c, planID)
	if err != nil {
		return
	}

	// Limit component.Servers to puppet Server
	component.Servers = component.GetServersOnPuppetServer(puppetServerID)

	// params
	params, err := getComponentPuppetPlanParams(component, puppetPlan)
	if err != nil {
		return
	}

	if c.Request.Method == "POST" { // BUILD
		// Verify Job is Enabled
		if !puppetPlan.Enabled {
			err = errors.New("attempted too run plan that is disabled")
			log.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
			return
		}

		// Process any submitted params (overrides)
		submittedParams := c.PostFormMap("Params")
		for k, v := range submittedParams {
			params[k] = v
		}

		baseURL := fmt.Sprintf("%s/component/%v", location.Get(c).String(), component.ID)
		job, err := puppet.RunPuppetPlan(puppetServer, puppetPlan, params, baseURL)
		if err != nil {
			log.Error("Error in RunPuppetPlan: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}
		job.InitiatorID = component.ID
		job.InitiatorType = "Component"
		err = job.Save()
		if err != nil {
			log.Error("Error saving job: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}

		htmlTemplate = "common-success-redirect.gohtml"
		redirectURL := fmt.Sprintf("/environment/%v/components", component.EnvironmentID)
		data = gin.H{
			"status":       "success",
			"redirectURL":  redirectURL,
			"params":       params,
			"puppetPlan":   puppetPlan,
			"puppetServer": puppetServer.Name,
			"job":          job,
		}

	} else { // PREVIEW
		htmlTemplate = "puppetPlan-preview.gohtml"
		data = gin.H{
			"status":     "preview",
			"params":     params,
			"puppetPlan": puppetPlan,
			"component":  component,
		}
	}

	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: htmlTemplate,
		Data:     data,
		Offered:  formatAllSupported,
	})

}

// ComponentRunPuppetTask runs a PuppetTask against a Component (POST), preview a run (GET)
// Path Params:
// - id - Component.ID
// - taskID - puppetTask.ID
func ComponentRunPuppetTask(c *gin.Context) {
	var htmlTemplate string
	var data gin.H

	// Component
	component, err := getComponent(c)
	if err != nil {
		return
	}

	// Puppet Server
	puppetServerID, err := validateID(c, "puppetServerID")
	if err != nil {
		return
	}
	puppetServer, err := getPuppetServerByID(c, puppetServerID)
	if err != nil {
		return
	}

	// PuppetTask
	taskID, err := validateID(c, "taskID")
	if err != nil {
		return
	}
	puppetTask, err := getPuppetTaskByID(c, taskID)
	if err != nil {
		return
	}

	// Limit component.Servers to puppet Server
	component.Servers = component.GetServersOnPuppetServer(puppetServerID)

	// params
	params, err := getComponentPuppetTaskParams(component, puppetTask)
	if err != nil {
		return
	}

	if c.Request.Method == "POST" { // BUILD
		// Verify Job is Enabled
		if !puppetTask.Enabled {
			err = errors.New("attempted too run task that is disabled")
			log.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
			return
		}

		// Process any submitted params (overrides)
		submittedParams := c.PostFormMap("Params")
		for k, v := range submittedParams {
			params[k] = v
		}

		baseURL := fmt.Sprintf("%s/component/%v", location.Get(c).String(), component.ID)
		job, err := puppet.RunPuppetTask(puppetServer, puppetTask, component.GetServerList(), params, baseURL)
		if err != nil {
			log.Error("Error in RunPuppetPlan: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}
		job.InitiatorID = component.ID
		job.InitiatorType = "Component"
		err = job.Save()
		if err != nil {
			log.Error("Error saving job: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}

		htmlTemplate = "common-success-redirect.gohtml"
		redirectURL := fmt.Sprintf("/environment/%v/components", component.EnvironmentID)
		data = gin.H{
			"status":       "success",
			"redirectURL":  redirectURL,
			"params":       params,
			"puppetTask":   puppetTask,
			"puppetServer": puppetServer.Name,
			"job":          job,
		}
	} else { // PREVIEW
		htmlTemplate = "puppetTask-preview.gohtml"
		data = gin.H{
			"status":     "preview",
			"params":     params,
			"puppetTask": puppetTask,
			"component":  component,
		}
	}

	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: htmlTemplate,
		Data:     data,
		Offered:  formatAllSupported,
	})

}

// getComponentPuppetPlanParams will parse the Job Params templates and return them
func getComponentPuppetPlanParams(component *models.Component, puppetPlan *models.PuppetPlan) (params map[string]string, err error) {
	params = make(map[string]string)
	planParams, err := puppetPlan.GetParams()
	if err != nil {
		log.Error("Error retrieving planParams: ", err)
		return
	}

	for _, planParam := range planParams {
		if planParam.TemplateValue == "" {
			defaultValue, err := planParam.GetDefaultValue()
			if err != nil {
				log.WithFields(log.Fields{
					"puppetPlan": puppetPlan,
					"planParam":  planParam,
				}).Error("Error retrieving default value", err)
				// NOTE: defaultValue will be ""
			}
			params[planParam.Name] = defaultValue
		} else {
			// GetTemplate
			tpl, err := planParam.GetTemplate()
			if err != nil {
				log.WithField("puppetPlanParam", planParam.ID).Error("Error getting template.")
				return params, err
			}

			// Execute Template
			out := new(bytes.Buffer)
			err = tpl.Execute(out, component)
			if err != nil {
				log.WithFields(log.Fields{
					"planName":  puppetPlan.Name,
					"paramName": planParam.Name,
					"paramID":   planParam.ID,
				}).Error("Error Executing Template for puppetPlan: ", err)
			}
			params[planParam.Name] = out.String()
		}
	}
	return
}

// getComponentPuppetTaskParams will parse the Job Params templates and return them
func getComponentPuppetTaskParams(component *models.Component, puppetTask *models.PuppetTask) (params map[string]string, err error) {
	params = make(map[string]string)
	taskParams, err := puppetTask.GetParams()
	if err != nil {
		log.Error("Error retrieving taskParams: ", err)
		return
	}

	for _, taskParam := range taskParams {
		if taskParam.TemplateValue == "" {
			defaultValue, err := taskParam.GetDefaultValue()
			if err != nil {
				log.WithFields(log.Fields{
					"puppetTask": puppetTask,
					"taskParam":  taskParam,
				}).Error("Error retrieving default value", err)
				// NOTE: defaultValue will be ""
			}
			params[taskParam.Name] = defaultValue
		} else {
			// GetTemplate
			tpl, err := taskParam.GetTemplate()
			if err != nil {
				log.WithField("puppetTaskParam", taskParam.ID).Error("Error getting template.")
				return params, err
			}

			// Execute Template
			out := new(bytes.Buffer)
			err = tpl.Execute(out, component)
			if err != nil {
				log.WithFields(log.Fields{
					"taskName":  puppetTask.Name,
					"paramName": taskParam.Name,
					"paramID":   taskParam.ID,
				}).Error("Error Executing Template for puppetTask: ", err)
			}
			params[taskParam.Name] = out.String()
		}
	}
	return
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

func getComponent(c *gin.Context) (component *models.Component, err error) {
	// First retrieve "id" parameter
	id, err := validateID(c, "id")
	if err != nil {
		return
	}
	return getComponentByID(c, id)
}

func getComponentByID(c *gin.Context, id uint) (component *models.Component, err error) {
	// Get Component from DB
	component, err = models.GetComponentByID(id)
	if err != nil {
		log.Error("Error retrieving Component from DB: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving Component from DB: " + err.Error()})
		return
	}
	// Another check to verify the Component was retrieved, id should not be 0
	if component.ID == 0 {
		err = errNotExist
		log.Error("Error component id should not be 0 (not found)")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	return // success
}
