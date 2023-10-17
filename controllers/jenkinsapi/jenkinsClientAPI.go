package jenkinsapi

import (
	"context"
	"strings"

	"github.com/bndr/gojenkins"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

// getJenkinsClient will return the jenkins client
func getJenkinsClient(ctx context.Context, jenkinsServer *models.JenkinsServer) (apiClient *gojenkins.Jenkins, err error) {
	if jenkinsServer.APIClient == nil {
		apiClient = gojenkins.CreateJenkins(nil, jenkinsServer.GetURL(), jenkinsServer.Username, jenkinsServer.Token)
		// Provide CA certificate if server is SSL and using self-signed CA certificate
		if jenkinsServer.SSL && jenkinsServer.CACert != "" {
			apiClient.Requester.CACert = []byte(jenkinsServer.CACert)
		}
		// ctx.Value("debug")
		_, err = apiClient.Init(ctx)
		if err != nil {
			log.Error("ERROR getJenkinsClient: " + err.Error())
			return
		}
		jenkinsServer.APIClient = apiClient
	} else {
		apiClient = jenkinsServer.APIClient
	}
	return
}

// splitPath will split the path into a job name and an array of strings for parents (for passing to GetJob)
func splitPath(path string) (jobName string, parents []string) {
	// Catch Trailing slash
	path = strings.TrimSuffix(path, "/")

	// Split on / and extract job name and parents as an slice of strings
	paths := strings.Split(path, "/")
	jobName = paths[len(paths)-1]
	parents = paths[1 : len(paths)-1]
	return
}

// IsFolder determines if the job class is a "folder like" object
func IsFolder(jobClass string) bool {
	for _, t := range FolderTypes {
		if t == jobClass {
			return true
		}
	}
	return false
}

// Info returns JenkinsServer Info
func Info(ctx context.Context, jenkinsServer *models.JenkinsServer) (info *gojenkins.ExecutorResponse, err error) {
	apiClient, err := getJenkinsClient(ctx, jenkinsServer)
	if err != nil {
		return
	}

	info, err = apiClient.Info(ctx)
	return
}
