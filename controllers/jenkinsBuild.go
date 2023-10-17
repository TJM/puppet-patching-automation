package controllers

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/jenkinsapi"
	"github.com/tjm/puppet-patching-automation/models"
)

// BuildJenkinsJob builds a job (POST), preview a build (GET)
// Path Params:
// - id - PatchRunID
// - jobID - JenkinsJobID
func BuildJenkinsJob(c *gin.Context) {
	var htmlTemplate = "patchRun-success-redirect.gohtml"
	var data gin.H
	var waitForBuild bool

	// PatchRun
	patchRun, err := getPatchRun(c)
	if err != nil {
		return
	}
	// JenkinsJob
	jobID, err := validateID(c, "jobID")
	if err != nil {
		return
	}
	jenkinsJob, err := getJenkinsJobByID(c, jobID)
	if err != nil {
		return
	}

	//JenkinsServer
	jenkinsServer, err := getJenkinsServerByID(c, jenkinsJob.JenkinsServerID)
	if err != nil {
		return
	}

	// BuildParams
	buildParams, err := getBuildParams(patchRun, jenkinsJob)
	if err != nil {
		return
	}

	if c.Request.Method == "POST" { // BUILD
		// Verify Job is Enabled
		if !jenkinsJob.Enabled {
			err = errors.New("attempted too run job that is disabled")
			log.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
			return
		}

		if c.Request.FormValue("waitForBuild") == "true" {
			waitForBuild = true
		}

		// Create jenkinsBuild and add associations
		jenkinsBuild := models.NewJenkinsBuild()
		jenkinsBuild.PatchRunID = patchRun.ID
		jenkinsBuild.JenkinsJobID = jenkinsJob.ID
		jenkinsBuild.JenkinsServerID = jenkinsServer.ID
		queueID, err := jenkinsapi.BuildJob(c, jenkinsServer, jenkinsJob.APIJobPath, jenkinsBuild, buildParams, waitForBuild)
		if err != nil {
			log.Error("Error building job: ", err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error building job: " + err.Error()})
			return
		}

		data = gin.H{
			"status":           "success",
			"buildParams":      buildParams,
			"jenkins_build":    jenkinsBuild,
			"patch_run_id":     patchRun.ID,
			"jenkins_queue_id": queueID,
		}
	} else { // PREVIEW
		htmlTemplate = "jenkinsBuild-preview.gohtml"
		data = gin.H{
			"status":       "preview",
			"build_params": buildParams,
			"jenkins_job":  jenkinsJob,
			"patch_run":    patchRun,
		}
	}

	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: htmlTemplate,
		Data:     data,
		Offered:  formatAllSupported,
	})

}

// getBuildParams will parse the Job Params templates and return them
func getBuildParams(patchRun *models.PatchRun, jenkinsJob *models.JenkinsJob) (buildParams map[string]string, err error) {
	buildParams = make(map[string]string)
	jobParams, err := jenkinsJob.GetParams()
	if err != nil {
		log.Error("Error retrieving jobParams: ", err)
		return
	}

	for _, jobParam := range jobParams {
		if jobParam.TemplateValue == "" {
			defaultValue, err := jobParam.GetDefaultValue()
			if err != nil {
				log.WithFields(log.Fields{
					"jenkinsJob": jenkinsJob,
					"jobParam":   jobParam,
				}).Error("Error retrieving default value", err)
				// NOTE: defaultValue will be ""
			}
			buildParams[jobParam.Name] = defaultValue
		} else {
			// GetTemplate
			tpl, err := jobParam.GetTemplate()
			if err != nil {
				log.WithField("jenkinsJobParam", jobParam.ID).Error("Error getting template.")
				return buildParams, err
			}

			// Execute Template
			out := new(bytes.Buffer)
			err = tpl.Execute(out, patchRun)
			if err != nil {
				log.WithFields(log.Fields{
					"jobName":   jenkinsJob.Name,
					"paramName": jobParam.Name,
					"paramID":   jobParam.ID,
				}).Error("Error Executing Template for job: ", err)
			}
			buildParams[jobParam.Name] = out.String()
		}
	}
	return
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// // getJenkinsBuild will get the id from context and return job
// func getJenkinsBuild(c *gin.Context) (jenkinsBuild *models.JenkinsBuild, err error) {
// 	// First retrieve "id" parameter
// 	id, err := validateID(c, "id")
// 	if err != nil {
// 		return
// 	}
// 	// Get jenkinsBuild from DB
// 	jenkinsBuild, err = models.GetJenkinsBuildByID(id)
// 	if err != nil {
// 		log.Error("Error retrieving jenkinsServer from DB: ", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error retrieving jenkinsServer from DB: " + err.Error()})
// 		return
// 	}
// 	// Another check to verify the jenkinsServer was retrieved, id should not be 0
// 	if jenkinsBuild.ID == 0 {
// 		err = errNotExist
// 		log.Error("Error jenkinsBuild id should not be 0 (not found)")
// 		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Error jenkinsBuild id should not be 0 (not found)"})
// 		return
// 	}
// 	return // success
// }
