package puppet

import (
	"github.com/puppetlabs/go-pe-client/pkg/orch"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

// GetJobByID returns the Job details for a specific Job ID on a PuppetServer
func GetJobByID(p *models.PuppetServer, jobID string) (job *orch.Job, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	job, err = client.Job(jobID)
	if err != nil {
		log.Error("Error Getting Job: ", err)
	}
	return
}

// GetJobReportByID returns the Job Report for a specific Job ID on a PuppetServer
func GetJobReportByID(p *models.PuppetServer, jobID string) (report *orch.JobReport, err error) {
	client, err := getOrchClient(p)
	if err != nil {
		return // already logged
	}
	report, err = client.JobReport(jobID)
	if err != nil {
		log.Error("Error Getting JobReport: ", err)
	}
	return
}
