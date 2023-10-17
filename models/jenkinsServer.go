package models

import (
	"fmt"

	"github.com/bndr/gojenkins"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// JenkinsServer defines a Jenkins Server
type JenkinsServer struct {
	gorm.Model
	Name          string `binding:"required"`
	Description   string
	Hostname      string `binding:"required,fqdn"`
	Port          uint   `binding:"required,numeric,gte=1,lte=65535"`
	Username      string `binding:"required"`
	Token         string // `binding:"required"` - Do not make "required"
	SSL           bool
	SSLSkipVerify bool
	CACert        string
	Enabled       bool
	Jobs          JenkinsJobs        `json:"-" yaml:"-" xml:"-" form:"-"`
	APIClient     *gojenkins.Jenkins `gorm:"-" json:"-" yaml:"-" xml:"-" form:"-"`
}

// JenkinsServers is a list of JenkinsServer
type JenkinsServers []*JenkinsServer

// NewJenkinsServer returns a new jenkinsServer object
func NewJenkinsServer() (j *JenkinsServer) {
	j = new(JenkinsServer)
	// Defaults
	j.Port = 443
	j.SSL = true
	j.Enabled = true
	return
}

// Init : Create new PatchRun object
func (j *JenkinsServer) Init() {
	GetDB().Create(j)
}

// Save : Save PatchRun object
func (j *JenkinsServer) Save() {
	GetDB().Save(j)
}

// Delete : Delete PatchRun object
func (j *JenkinsServer) Delete(cascade bool) (err error) {
	if cascade {
		var jobs JenkinsJobs
		jobs, err = j.GetJobs()
		if err != nil {
			log.Error("Error RETRIEVING child Jobs: ", err)
			return
		}
		for _, job := range jobs {
			err = job.Delete(cascade)
			if err != nil {
				return
			}
		}
	}
	err = GetDB().Delete(j).Error
	return
}

// GetJobs : Get JenkinsJobs
func (j *JenkinsServer) GetJobs() (jobs JenkinsJobs, err error) {
	err = GetDB().Model(j).Association("Jobs").Find(&jobs)
	return
}

// AddJob : Add JenkinsJob
func (j *JenkinsServer) AddJob(job *JenkinsJob) (err error) {
	err = GetDB().Model(j).Association("Jobs").Append(job)
	return
}

// DeleteJob : Delete JenkinsJob
func (j *JenkinsServer) DeleteJob(job *JenkinsJob) (err error) {
	err = GetDB().Model(j).Association("Jobs").Delete(job)
	return
}

// GetJenkinsServerByID returns patch run object by ID
func GetJenkinsServerByID(id uint) (j *JenkinsServer, err error) {
	j = new(JenkinsServer)
	err = GetDB().First(j, id).Error
	return
}

// GetEnabledJenkinsServers returns a list of enabled JenkinsServers
func GetEnabledJenkinsServers() (servers JenkinsServers) {
	servers = make(JenkinsServers, 0)
	GetDB().Where(&JenkinsServer{Enabled: true}).Order("name").Find(&servers)
	return
}

// IsEnabledJenkinsServers returns a bool if there are enabled JenkinsServers
func IsEnabledJenkinsServers() (enabled bool) {
	var count int64
	GetDB().Model(&JenkinsServer{}).Where(&JenkinsServer{Enabled: true}).Count(&count)
	return count > 0
}

// GetJenkinsServers returns a list of all JenkinsServers
func GetJenkinsServers() (servers JenkinsServers) {
	servers = make(JenkinsServers, 0)
	GetDB().Order("name").Find(&servers)
	return
}

// GetURL returns the URL for the Jenkins Server
func (j *JenkinsServer) GetURL() (url string) {
	if j.SSL {
		url = "https://"
	} else {
		url = "http://"
	}
	url += j.Hostname
	if j.SSL && j.Port != 443 {
		url += fmt.Sprintf(":%v", j.Port)
	}
	if !j.SSL && j.Port != 80 {
		url += fmt.Sprintf(":%v", j.Port)
	}
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (j *JenkinsServer) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, JenkinsServers{}.GetBreadCrumbs()...) // Patch Run List
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("Jenkins Server: %s", j.Name), fmt.Sprintf("/config/jenkinsServer/%v", j.ID)))
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (servers JenkinsServers) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
	breadcrumbs = append(breadcrumbs, createBreadCrumb("Jenkins Servers", "/config/jenkinsServer"))
	return
}
