package puppet

import (
	"strings"

	"github.com/puppetlabs/go-pe-client/pkg/orch"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

// PatchServer will patch a specific server
func PatchServer(server *models.Server, baseURL string, test bool) (job *models.PuppetJob, err error) {
	puppetServer, err := models.GetPuppetServerByID(server.PuppetServerID)
	if err != nil {
		return
	}

	if test {
		job, err = runTestTask(puppetServer, []string{server.Name}, baseURL)
	} else {
		job, err = runPatchTask(puppetServer, []string{server.Name}, baseURL)
	}
	return
}

// RunPuppetTask will run a PuppetTask on a PuppetServer and return a PuppetJob
func RunPuppetTask(p *models.PuppetServer, task *models.PuppetTask, nodes []string, params map[string]string, baseURL string) (job *models.PuppetJob, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	taskParams, err := parseTaskParams(task, params)
	if err != nil {
		log.Error("Error Parsing Parameters: ", err)
		return
	}
	jobID, err := client.CommandTask(&orch.TaskRequest{
		Task:        task.Name,
		Params:      taskParams,
		Environment: task.Environment,
		Description: "Started from: " + baseURL,
		Scope: orch.Scope{
			Nodes: nodes,
		},
	})
	if err != nil {
		log.Error("Error RunPuppetTask: ", err)
		return
	}
	job, err = parseJobID(p, jobID)
	job.PuppetParentType = "Task"
	job.PuppetParentID = task.ID
	job.Name = task.Name
	// if err != nil, we have already logged an error, return it
	return
}

// runTestTask runs the test task on the nodes
func runTestTask(p *models.PuppetServer, nodes []string, baseURL string) (job *models.PuppetJob, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	patchingModule := strings.Split(p.FactName, ".")[0]
	jobID, err := client.CommandTask(&orch.TaskRequest{
		Description: "Started from: " + baseURL,
		Environment: "production",
		Task:        patchingModule + "::clean_cache",
		Params:      map[string]interface{}{},
		Scope: orch.Scope{
			Nodes: nodes,
		},
	})
	if err != nil {
		log.Error("Error runTestTask: ", err)
		return
	}
	job, err = parseJobID(p, jobID)
	job.PuppetParentType = "Default"
	job.Name = patchingModule + "::clean_cache"
	// if err != nil, we have already logged an error, return it
	return
}

// runPatchTask runs the patching task on the nodes
func runPatchTask(p *models.PuppetServer, nodes []string, baseURL string) (job *models.PuppetJob, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	patchingModule := strings.Split(p.FactName, ".")[0]
	jobID, err := client.CommandTask(&orch.TaskRequest{
		Description: "Started from: " + baseURL,
		Environment: "production",
		Task:        patchingModule + "::patch_server",
		Params: map[string]interface{}{
			// "clean_cache": "true",
			"reboot": "smart",
			// "yum_params": "",
		},
		Scope: orch.Scope{
			Nodes: nodes,
		},
	})
	if err != nil {
		log.Error("Error CommandTask: ", err)
		return
	}
	job, err = parseJobID(p, jobID)
	job.PuppetParentType = "Default"
	job.Name = patchingModule + "::patch_server"
	// if err != nil, we have already logged an error, return it
	return
}

// GetTask returns Task from PuppetServer
func GetTask(p *models.PuppetServer, env, module, taskName string) (task *orch.Task, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	task, err = client.Task(env, module, taskName)
	if err != nil {
		log.Error("Error Getting Task: ", err)
	}
	return
}

// GetTasks returns the available PuppetTasks for a PuppetServer
func GetTasks(p *models.PuppetServer, env string) (tasks *orch.Tasks, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	// Default env to "production"
	if env == "" {
		env = "production"
	}
	// Get Tasks for env
	tasks, err = client.Tasks(env)
	if err != nil {
		log.Error("Error Getting Tasks: ", err)
	}
	return
}

// UpdateTaskDetails associates a PuppetTask and Params to a PuppetServer for customization
func UpdateTaskDetails(p *models.PuppetServer, dbTask *models.PuppetTask) (err error) {
	nameSplit := strings.SplitN(dbTask.Name, "::", 2)
	module := nameSplit[0]
	taskName := "init"
	if len(nameSplit) > 1 {
		taskName = nameSplit[1]
	}
	log.WithFields(log.Fields{
		"puppetServer": p.Name,
		"nameSplit":    nameSplit,
		"environment":  dbTask.Environment,
	}).Info("UpdateTaskDetails")
	apiTask, err := GetTask(p, dbTask.Environment, module, taskName)
	if err != nil {
		log.WithFields(log.Fields{
			"puppetServer": p.Name,
			"module":       module,
			"taskName":     taskName,
		}).Error("Error GetTask", err)
	}
	dbTask.Description = apiTask.Metadata.Description
	dbTask.APITaskID = apiTask.ID
	err = dbTask.Save()
	if err != nil {
		log.Error("Error saving task to DB: ", err)
		return
	}
	err = p.AddPuppetTask(dbTask)
	if err != nil {
		log.Error("Error associating Task to PuppetServer: ", err)
		return
	}
	// Handle Parameters
	for apiName, apiParam := range apiTask.Metadata.Parameters {
		_ = apiName
		_ = apiParam
	}

	// Find removed params
	for _, dbParam := range dbTask.Params {
		var found bool
		for apiParamName := range apiTask.Metadata.Parameters {
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
				dbParam.IsNotInTask = true
				err = dbParam.Save()
				if err != nil {
					log.Error("ERROR saving dbParam: ", err)
				}
			}
		}
	}
	// Update DB Params from API
	params := make([]*models.PuppetTaskParam, 0)
	for apiParamName, apiParam := range apiTask.Metadata.Parameters {
		param := dbTask.Param(apiParamName)
		param.PuppetTaskID = dbTask.ID
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
	dbTask.Params = params
	return
}

// parseJobID returns the PuppetJob object given the "jobID" output from running a Puppet Task via the orchestrator API
func parseJobID(p *models.PuppetServer, jobID *orch.JobID) (job *models.PuppetJob, err error) {
	// NOTE: jobID is not what we expected, so here is example output:
	// jobID = {
	// "job": {
	//   "id": "https://puppetserver.company.com:8143/orchestrator/v1/jobs/338",
	//   "name": "338"
	// }

	log.WithField("jobID", jobID).Info("Parse JobID")
	job = models.NewPuppetJob()
	job.APIJobID = jobID.Job.Name // SEE example output above
	job.ConsoleURL = "https://" + p.Hostname + "/#/orchestration/tasks/task/" + jobID.Job.Name
	job.APIJobURL = jobID.Job.ID
	// job.PuppetParentID = plan.ID // We are not yet doing this
	job.PuppetParentType = "task"
	job.PuppetServerID = p.ID

	err = job.Save()
	if err != nil {
		log.Error("Error saving PuppetJob: ", err)
	}
	return
}

// parseTaskParams returns a map[string]interface from map[string]string for puppet plan parameters
func parseTaskParams(task *models.PuppetTask, params map[string]string) (newParams map[string]interface{}, err error) {
	var taskParam *models.PuppetTaskParam
	newParams = make(map[string]interface{})
	for k, v := range params {
		taskParam, err = task.GetParamByName(k)
		if err != nil {
			log.WithFields(log.Fields{
				"TaskID": task.ID,
				"Param":  k,
			}).Error("Error loading parameter: ", err)
			return
		}
		if v != "" {
			newParams[k] = v
		}
		val := getInterfaceValue(taskParam.Type, v)
		if val != nil {
			newParams[k] = val
		}
	}
	return
}
