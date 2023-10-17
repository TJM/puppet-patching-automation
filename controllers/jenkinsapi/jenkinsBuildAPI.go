package jenkinsapi

import (
	"context"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/events"
	"github.com/tjm/puppet-patching-automation/models"
)

// BuildJob creates a Build in a Job
func BuildJob(ctx context.Context, jenkinsServer *models.JenkinsServer, jobPath string, dbBuild *models.JenkinsBuild, buildParams map[string]string, waitForBuild bool) (queueID int64, err error) {
	apiClient, err := getJenkinsClient(ctx, jenkinsServer)
	if err != nil {
		return
	}
	// Get API Job
	jobName, parents := splitPath(jobPath)
	apiJob, err := apiClient.GetJob(ctx, jobName, parents...)
	if err != nil {
		log.WithFields(log.Fields{
			"jobName": jobName,
			"parents": parents,
		}).Error("GetJob: " + err.Error())
		return
	}
	// INVOKE! (build)
	queueID, err = apiJob.InvokeSimple(ctx, buildParams)
	if err != nil {
		log.Error("ERROR Building Job: "+jobPath, err)
		return
	}
	log.WithFields(log.Fields{
		"jobPath": jobPath,
		"queueID": queueID,
	}).Info("jenkinsapi.BuildJob")

	dbBuild.QueueID = queueID
	// Save to DB
	err = dbBuild.Init()
	if err != nil {
		log.Error("Error saving build to DB: ", err)
		return
	}

	if waitForBuild {
		err = GetBuildFromQueue(ctx, apiClient, apiJob, dbBuild)
		if err != nil {
			log.Error("Error populating trelloBoard: ", err)
		}
	} else {
		go func() {
			err = GetBuildFromQueue(ctx, apiClient, apiJob, dbBuild)
			if err != nil {
				log.Error("Error populating trelloBoard: ", err)
			}
		}()
	}
	return
}

// GetBuildFromQueue updates build object once its populated
func GetBuildFromQueue(ctx context.Context, apiClient *gojenkins.Jenkins, apiJob *gojenkins.Job, dbBuild *models.JenkinsBuild) (err error) {
	task, err := apiClient.GetQueueItem(ctx, dbBuild.QueueID)
	if err != nil {
		log.Error("Error Getting Task: ", err)
		return
	}

	log.WithFields(log.Fields{
		"QueueID":    dbBuild.QueueID,
		"Name":       task.Raw.Task.Name,
		"QueueURL":   task.Raw.URL,
		"Why":        task.Raw.Why,
		"Executable": task.Raw.Executable,
	}).Infof("apiTask before polling. (background)")

	// TODO: Do this in the background?
	// POLL Task - https://github.com/bndr/gojenkins/issues/161
	retry := 0
	maxRetry := 30
	for retry < maxRetry {
		if task.Raw.Executable.URL != "" {
			break
		}
		time.Sleep(1 * time.Second)
		_, _ = task.Poll(ctx) // TODO: decide whether the return values of Poll() are useful
		retry++
	}

	log.WithFields(log.Fields{
		"Name":       task.Raw.Task.Name,
		"QueueURL":   task.Raw.URL,
		"Why":        task.Raw.Why,
		"Executable": task.Raw.Executable,
		"Retries":    retry,
	}).Infof("apiTask after polling.")

	// get the build using the build number
	apiBuild, err := apiJob.GetBuild(ctx, task.Raw.Executable.Number)
	if err != nil {
		log.Error("Error Getting Build: ", err)
		return
	}

	dbBuild.Name = apiBuild.Job.Raw.Name
	dbBuild.APIBuildID = apiBuild.GetBuildNumber()
	dbBuild.URL = apiBuild.GetUrl()
	dbBuild.Status = apiBuild.GetResult()
	// TODO: See if there are any other useful fields

	// Save to DB
	err = dbBuild.Save()
	if err != nil {
		log.Error("Error saving build to DB: ", err)
		return
	}

	event := models.NewEvent(models.ActionJenkinsBuildCreated)
	patchRun, _ := models.GetPatchRunByID(dbBuild.PatchRunID)
	event.Target = dbBuild
	events.PatchRunEvent(new(gin.Context), patchRun, event) // NOTE: empty gin.Context here

	return
}

// GetBuild will return the build response
func GetBuild(ctx context.Context, jenkinsServer *models.JenkinsServer, dbJob *models.JenkinsJob, buildID int64) (apiBuild *gojenkins.Build, err error) {
	apiClient, err := getJenkinsClient(ctx, jenkinsServer)
	if err != nil {
		return
	}
	// Get API Job
	jobName, parents := splitPath(dbJob.APIJobPath)
	apiJob, err := apiClient.GetJob(ctx, jobName, parents...)
	if err != nil {
		log.WithFields(log.Fields{
			"jobName": jobName,
			"parents": parents,
		}).Error("GetJob: " + err.Error())
		return
	}
	// Get Build
	apiBuild, err = apiJob.GetBuild(ctx, buildID)
	if err != nil {
		log.WithFields(log.Fields{
			"job":     apiJob.Raw.Name,
			"buildID": buildID,
		}).Error("GetBuild: " + err.Error())
		return
	}

	return
}
