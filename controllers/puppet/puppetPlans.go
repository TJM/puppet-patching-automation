package puppet

import (
	"strings"

	"github.com/puppetlabs/go-pe-client/pkg/orch"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

// PatchComponent - Patch all servers for a component
// NOTE: Currently there are no health checks done and all servers may patch at the same time.
// Loops through each of the servers in a component, and then tries to patch them in a group, based
// on the puppet server they are associated to.
func PatchComponent(component *models.Component, baseURL string) (jobs []*models.PuppetJob, err error) {

	puppetServers, serverList, err := getComponentDetails(component)
	if err != nil {
		return
	}

	// Loop through each of the discovered lists of servers
	for psid, servers := range serverList {
		var job *models.PuppetJob
		nodes := make([]string, 0)
		for _, server := range servers {
			nodes = append(nodes, server.Name)
		}

		// TODO: Figure out how to make this function reusable by only calling this if its "legacy patching"

		job, err = runPatchPlan(puppetServers[psid], component.HealthCheckScript, nodes, baseURL)
		if err != nil {
			log.Error("ERROR in PatchComponent: ", err)
			// return
		} else {
			job.InitiatorID = component.ID
			job.InitiatorType = "Component"
			err = job.Save()
			if err != nil {
				log.Error("ERROR Saving job in PatchComponent: ", err)
				// return
			}
			jobs = append(jobs, job)
		}
	}
	return
}

// RunPuppetPlan will run a Puppet Plan on a Puppet Server and return the PuppetJob
func RunPuppetPlan(p *models.PuppetServer, plan *models.PuppetPlan, params map[string]string, baseURL string) (job *models.PuppetJob, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	planParams, err := parsePlanParams(plan, params)
	if err != nil {
		log.Error("Error Parsing Parameters: ", err)
		return
	}
	jobID, err := client.CommandPlanRun(&orch.PlanRunRequest{
		Name:        plan.Name,
		Params:      planParams,
		Environment: plan.Environment,
		Description: "Started from: " + baseURL,
	})
	if err != nil {
		log.Error("Error RunPuppetPlan: ", err)
		return
	}
	job, err = parsePlanRunJobID(p, jobID)
	job.PuppetParentType = "Plan"
	job.PuppetParentID = plan.ID
	job.Name = plan.Name
	// if err != nil, we have already logged an error, return it
	return
}

// getComponentDetails returns a list of PuppetServers and a serverList indexed by puppetServer.ID
func getComponentDetails(component *models.Component) (puppetServers map[uint]*models.PuppetServer, serverList map[uint][]*models.Server, err error) {
	puppetServers = make(map[uint]*models.PuppetServer) // List of Puppet Servers by ID
	serverList = make(map[uint][]*models.Server)        // List of servers by Puppet Server ID

	for _, server := range component.GetServers() {
		// Make sure we have a PuppetServer
		if _, ok := puppetServers[server.PuppetServerID]; !ok {
			puppetServers[server.PuppetServerID], err = models.GetPuppetServerByID(server.PuppetServerID)
			if err != nil {
				return
			}
		}
		serverList[server.PuppetServerID] = append(serverList[server.PuppetServerID], server)
	}
	return
}

// runPatchPlan runs the test task on the nodes
func runPatchPlan(p *models.PuppetServer, postRebootScriptPath string, nodes []string, baseURL string) (job *models.PuppetJob, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	jobID, err := client.CommandPlanRun(&orch.PlanRunRequest{
		Name: "patchy::cluster_patching",
		Params: map[string]interface{}{
			"targets":                nodes,
			"post_reboot_scriptpath": postRebootScriptPath,
		},
		Environment: "production",
		Description: "Started from: " + baseURL,
	})
	if err != nil {
		log.Error("Error runPatchPlan: ", err)
		return
	}
	job, err = parsePlanRunJobID(p, jobID)
	job.PuppetParentType = "Default"
	job.Name = "patchy::cluster_patching"
	// if err != nil, we have already logged an error, return it
	return
}

// GetPlan returns Plan from PuppetServer
func GetPlan(p *models.PuppetServer, env, module, taskName string) (plan *orch.Plan, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	plan, err = client.Plan(env, module, taskName)
	if err != nil {
		log.Error("Error Getting Plan: ", err)
	}
	return
}

// GetPlans returns the available Puppet Plans for a PuppetServer
func GetPlans(p *models.PuppetServer, env string) (plans *orch.Plans, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	// Default env to "production"
	if env == "" {
		env = "production"
	}
	// Get Tasks for env
	plans, err = client.Plans(env)
	if err != nil {
		log.Error("Error Getting Plans: ", err)
	}
	return
}

// UpdatePlanDetails associates a PuppetPlan and Params to a PuppetServer for customization
func UpdatePlanDetails(p *models.PuppetServer, dbPlan *models.PuppetPlan) (err error) {
	nameSplit := strings.SplitN(dbPlan.Name, "::", 2)
	module := nameSplit[0]
	planName := "init"
	if len(nameSplit) > 1 {
		planName = nameSplit[1]
	}
	log.WithFields(log.Fields{
		"puppetServer": p.Name,
		"nameSplit":    nameSplit,
		"environment":  dbPlan.Environment,
	}).Info("UpdatePlanDetails")
	apiPlan, err := GetPlan(p, dbPlan.Environment, module, planName)
	if err != nil {
		log.WithFields(log.Fields{
			"puppetServer": p.Name,
			"module":       module,
			"planName":     planName,
		}).Error("Error GetPlan", err)
	}
	dbPlan.Description = apiPlan.Metadata.Description
	dbPlan.APIPlanID = apiPlan.ID
	err = dbPlan.Save()
	if err != nil {
		log.Error("Error saving task to DB: ", err)
		return
	}
	err = p.AddPuppetPlan(dbPlan)
	if err != nil {
		log.Error("Error associating Plan to PuppetServer: ", err)
		return
	}
	// Handle Parameters
	for apiName, apiParam := range apiPlan.Metadata.Parameters {
		_ = apiName
		_ = apiParam
	}

	// Find removed params
	for _, dbParam := range dbPlan.Params {
		var found bool
		for apiParamName := range apiPlan.Metadata.Parameters {
			// This is crude, find a parameter name that matches
			if apiParamName == dbParam.Name {
				found = true
				break
			}
		}
		if !found {
			log.WithField("paramName", dbParam.Name).Info("Found a parameter that is no longer in the job.")
			if dbParam.TemplateValue == "" {
				// If no TemplateValue is set, just delete
				err = dbParam.Delete(true)
				if err != nil {
					log.Error("ERROR deleting dbParam: ", err)
				}
			} else {
				// Flag the values as "IsNotInTask"
				dbParam.IsNotInPlan = true
				err = dbParam.Save()
				if err != nil {
					log.Error("ERROR saving dbParam: ", err)
				}
			}
		}
	}

	// Update DB Params from API
	params := make([]*models.PuppetPlanParam, 0)
	for apiParamName, apiParam := range apiPlan.Metadata.Parameters {
		param := dbPlan.Param(apiParamName)
		param.PuppetPlanID = dbPlan.ID
		param.Description = apiParam.Description
		param.Type = apiParam.Type
		err = param.Save()
		if err != nil {
			log.WithFields(log.Fields{
				"Name": param.Name,
				"Type": param.Type,
			}).Error("Error saving to DB: ", err)
		}
		params = append(params, param)
	}
	dbPlan.Params = params
	return
}

// parsePlanRunJobID returns the PuppetJob object given the "jobID" output from running a Puppet Plan via the orchestrator API
func parsePlanRunJobID(p *models.PuppetServer, jobID *orch.PlanRunJobID) (job *models.PuppetJob, err error) {
	// NOTE: jobID is not what we expected, so here is example output:
	// jobID = {
	//   "name": "338"
	// }
	log.WithField("jobID", jobID).Info("Parse Plan Run JobID")
	job = models.NewPuppetJob()
	job.APIJobID = jobID.Name // SEE example output above
	job.ConsoleURL = "https://" + p.Hostname + "/#/orchestration/plans/plan/" + jobID.Name
	// job.APIJobURL = jobID.ID // NOT AVAILABLE IN PlanRunJobID
	job.PuppetServerID = p.ID

	err = job.Save()
	if err != nil {
		log.Error("Error saving PuppetJob: ", err)
	}
	return
}

// parsePlanParams returns a map[string]interface from map[string]string for puppet plan parameters
func parsePlanParams(plan *models.PuppetPlan, params map[string]string) (newParams map[string]interface{}, err error) {
	var planParam *models.PuppetPlanParam
	newParams = make(map[string]interface{})
	for k, v := range params {
		planParam, err = plan.GetParamByName(k)
		if err != nil {
			log.WithFields(log.Fields{
				"PlanID": plan.ID,
				"Param":  k,
			}).Error("Error loading parameter: ", err)
			return
		}
		val := getInterfaceValue(planParam.Type, v)
		if val != nil {
			newParams[k] = val
		}
	}
	return
}
