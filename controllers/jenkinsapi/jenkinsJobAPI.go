package jenkinsapi

import (
	"context"
	"errors"

	"github.com/bndr/gojenkins"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

var (
	// Errors
	errInvalidJobPath = errors.New("invalid Job Path")
	errInvalidJobType = errors.New("invalid Job Type")

	// FolderTypes lists the types of jenkins classes that contain other jobs
	FolderTypes = []string{
		"jenkins.branch.OrganizationFolder",
		"com.cloudbees.hudson.plugins.folder.Folder",
		"org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject",
	}
)

// GetJobsMetadata will get metadata for all Jenkins Jobs in a folder (or for the specific job named in the path)
func GetJobsMetadata(ctx context.Context, jenkinsServer *models.JenkinsServer, path string) (metadata []gojenkins.InnerJob, err error) {
	apiClient, err := getJenkinsClient(ctx, jenkinsServer)
	if err != nil {
		return
	}

	// If no folder path was given, return GetAllJobNames
	if path == "/" || path == "" {
		metadata, err = apiClient.GetAllJobNames(ctx)
		if err != nil {
			log.Error("ERROR from getAllJobNames: ", err.Error())
		}
		return
	}

	jobID, parentIDs := splitPath(path)
	apiJob, err := apiClient.GetJob(ctx, jobID, parentIDs...)
	if err != nil {
		log.WithFields(log.Fields{
			"jobID":     jobID,
			"parentIDs": parentIDs,
		}).Error("GetJob: " + err.Error())
		return
	}

	// If it is a folder, get the job(s) metadata it contains
	if IsFolder(apiJob.GetDetails().Class) {
		metadata = apiJob.GetInnerJobsMetadata()
		return
	}

	// display some info like the "InnerJob Metadata" above and the parameters
	// jobParams, err := apiJob.GetParameters()
	// if err != nil {
	// 	log.WithField("job", apiJob.Base).Errorf("Error getting parameters: " + err.Error())
	// }

	metadata = make([]gojenkins.InnerJob, 1)
	metadata[1] = gojenkins.InnerJob{
		Class: apiJob.Raw.Class,
		Name:  apiJob.Raw.Name,
		Url:   apiJob.Raw.URL,
		Color: apiJob.Raw.Color,
	}

	return
}

// AddJob will Add a Jenkins Job to the DB for templating
func AddJob(ctx context.Context, jenkinsServer *models.JenkinsServer, path string) (dbJob *models.JenkinsJob, err error) {
	// If no folder path was given, ERROR
	if path == "/" || path == "" {
		return nil, errInvalidJobPath
	}

	dbJob = models.NewJenkinsJob()
	dbJob.JenkinsServerID = jenkinsServer.ID
	dbJob.APIJobPath = path

	err = UpdateDBJobFromAPIJob(ctx, jenkinsServer, dbJob)
	return
}

// UpdateDBJobFromAPIJob will update the DB Job from the API Job
func UpdateDBJobFromAPIJob(ctx context.Context, jenkinsServer *models.JenkinsServer, dbJob *models.JenkinsJob) (err error) {
	apiClient, err := getJenkinsClient(ctx, jenkinsServer)
	if err != nil {
		return
	}
	jobName, parents := splitPath(dbJob.APIJobPath)
	apiJob, err := apiClient.GetJob(ctx, jobName, parents...)
	if err != nil {
		log.WithFields(log.Fields{
			"jobName": jobName,
			"parents": parents,
		}).Error("GetJob: " + err.Error())
		return
	}

	// If it is a folder, ERROR
	if IsFolder(apiJob.GetDetails().Class) {
		log.WithFields(log.Fields{
			"jobName": jobName,
			"parents": parents,
			"Class":   apiJob.GetDetails().Class,
		}).Error(errInvalidJobType)
		return
	}

	dbJob.Name = apiJob.GetName()
	dbJob.Description = apiJob.GetDescription()
	dbJob.URL = apiJob.Raw.URL
	err = dbJob.Save()
	if err != nil {
		log.Error("Error saving dbJob to DB: ", err)
		return
	}
	dbJob.Params = parseAPIJobParameters(ctx, apiJob, dbJob)
	return
}

// parseAPIJobParameters will parse the parameters from a Jenkins API Job and create/return the JenkinsJobParams objects
// TODO: handle updating parameters
func parseAPIJobParameters(ctx context.Context, apiJob *gojenkins.Job, dbJob *models.JenkinsJob) (params []*models.JenkinsJobParam) {
	apiParams, err := apiJob.GetParameters(ctx)
	if err != nil {
		log.Error("Error getting parameters for job" + err.Error())
		return nil
	}

	// Find removed params
	for _, dbParam := range dbJob.Params {
		var found bool
		for _, apiParam := range apiParams {
			// This is crude, find a parameter name that matches
			if apiParam.Name == dbParam.Name {
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
				// Flag the values as "IsNotInJob"
				dbParam.IsNotInJob = true
				err = dbParam.Save()
				if err != nil {
					log.Error("ERROR saving dbParam: ", err)
				}
			}
		}
	}

	params = make([]*models.JenkinsJobParam, len(apiParams))
	for idx, apiParam := range apiParams {
		param := dbJob.Param(apiParam.Name)
		param.JenkinsJobID = dbJob.ID
		param.Description = apiParam.Description
		param.Type = apiParam.Type
		err = param.SetDefaultValue(apiParam.DefaultParameterValue.Value)
		if err != nil {
			log.WithFields(log.Fields{
				"Name": param.Name,
				"Type": param.Type,
			}).Error("Error setting default value")
		}
		err = param.Save()
		if err != nil {
			log.WithFields(log.Fields{
				"Name": param.Name,
				"Type": param.Type,
			}).Error("Error saving to DB: ", err)
		}
		params[idx] = param
	}
	return
}
