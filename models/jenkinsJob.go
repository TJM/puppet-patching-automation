package models

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// JenkinsJob defines a Jenkins Job
type JenkinsJob struct {
	gorm.Model
	Name             string
	Description      string
	URL              string
	APIJobPath       string // Used by gojenkins
	Enabled          bool
	IsForPatchRun    bool
	IsForApplication bool
	IsForServer      bool
	JenkinsServerID  uint               // Parent Jenkins Server ID
	JenkinsServer    *JenkinsServer     `json:"-" yaml:"-" xml:"-" form:"-"` // Parent Jenkins Server ID
	Params           []*JenkinsJobParam `json:"-" yaml:"-" xml:"-" form:"-"` // Params are owned by a JenkinsJob
}

// JenkinsJobs is a list of JenkinsJob objects
type JenkinsJobs []*JenkinsJob

// NewJenkinsJob returns a new JenkinsJob object
func NewJenkinsJob() (j *JenkinsJob) {
	j = new(JenkinsJob)
	// Defaults
	j.Enabled = true
	j.IsForPatchRun = true
	return
}

// Init : Create new PatchRun object
func (j *JenkinsJob) Init() error {
	return GetDB().Create(j).Error
}

// Save : Save PatchRun object
func (j *JenkinsJob) Save() error {
	return GetDB().Save(j).Error
}

// Delete : Delete PatchRun object
func (j *JenkinsJob) Delete(cascade bool) (err error) {
	if cascade {
		var params []*JenkinsJobParam
		params, err = j.GetParams()
		if err != nil {
			log.Error("Error RETRIEVING child params: ", err)
			return
		}
		for _, param := range params {
			err = param.Delete(cascade)
			if err != nil {
				return
			}
		}
		// TODO: Should we delete JenkinsBuild(s) ... currently not doing it
	}
	return GetDB().Delete(j).Error // TODO: Catch Error on delete from DB
}

// GetJenkinsJobByID returns patch run object by ID
func GetJenkinsJobByID(id uint) (j *JenkinsJob, err error) {
	j = new(JenkinsJob)
	err = GetDB().Preload("Params").First(j, id).Error
	return
}

// GetParams : Get JenkinsParams for this job
func (j *JenkinsJob) GetParams() (params []*JenkinsJobParam, err error) {
	err = GetDB().Model(j).Association("Params").Find(&params)
	return
}

// GetParamByName : Get JenkinsParam by Name for this job
func (j *JenkinsJob) GetParamByName(paramName string) (param *JenkinsJobParam, err error) {
	err = GetDB().Model(j).Association("Params").Find(&param, &JenkinsJobParam{Name: paramName})
	return
}

// Param : Return a JenkinsJobParam object by name (create if not exist)
func (j *JenkinsJob) Param(name string) (param *JenkinsJobParam) {
	param = new(JenkinsJobParam)
	GetDB().Where(JenkinsJobParam{Name: name, JenkinsJobID: j.ID}).FirstOrCreate(param)
	return
}

// AddParam : Add JenkinsParam for this job
func (j *JenkinsJob) AddParam(param *JenkinsJobParam) (err error) {
	err = GetDB().Model(j).Association("Params").Append(param)
	return
}

// // AddParams : Add JenkinsParam for this job
// func (j *JenkinsJob) AddParams(params []*JenkinsJobParam) (err error) {
// 	GetDB().Model(j).Association("Params").Append(params) // TODO catch errors?
// 	return
// }

// GetEnabledJenkinsJobs returns a list of enabled JenkinsJobs
func GetEnabledJenkinsJobs() (servers JenkinsJobs) {
	servers = make(JenkinsJobs, 0)
	GetDB().Where(&JenkinsJob{Enabled: true}).Order("name").Find(&servers)
	return
}

// IsEnabledJenkinsJobs returns a bool if there are enabled JenkinsJobs
func IsEnabledJenkinsJobs() (enabled bool) {
	var count int64
	GetDB().Model(&JenkinsJob{}).Where(&JenkinsJob{Enabled: true}).Count(&count)
	return count > 0
}

// GetJenkinsJobs returns a list of all JenkinsJobs
func GetJenkinsJobs() (servers JenkinsJobs) {
	servers = make(JenkinsJobs, 0)
	GetDB().Preload("JenkinsServer").Order("name").Find(&servers)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (j *JenkinsJob) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	var err error
	if j.JenkinsServer == nil { // Retrieve JenkinsServer
		j.JenkinsServer, err = GetJenkinsServerByID(j.JenkinsServerID)
		if err != nil {
			log.WithField("JenkinsServerID", j.JenkinsServerID).Warn("Error getting breadcrumbs: " + err.Error())
		}
	}

	if j.JenkinsServer.ID == 0 { // No Jenkins Server Found (this shouldn't happen), return generic list
		breadcrumbs = append(breadcrumbs, JenkinsJobs{}.GetBreadCrumbs()...)
	} else {
		breadcrumbs = append(breadcrumbs, JenkinsJobs{}.GetBreadCrumbs(j.JenkinsServer)...)
	}
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("Jenkins Job: %s", j.Name), fmt.Sprintf("/config/jenkinsJob/%v", j.ID)))
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (jobs JenkinsJobs) GetBreadCrumbs(jenkinsServer ...*JenkinsServer) (breadcrumbs BreadCrumbs) {
	if len(jenkinsServer) > 0 {
		s := jenkinsServer[0]
		// breadcrumbs = append(breadcrumbs, s.GetBreadCrumbs()...)
		// Just go back to the list of JenkinsServers
		breadcrumbs = append(breadcrumbs, JenkinsServers{}.GetBreadCrumbs()...)
		breadcrumbs = append(breadcrumbs, createBreadCrumb("Jenkins Jobs for: "+s.Name, fmt.Sprintf("/config/jenkinsServer/%v/jobs", s.ID)))
	} else {
		breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
		breadcrumbs = append(breadcrumbs, createBreadCrumb("Jenkins Jobs", "/config/jenkinsJob"))
	}
	return
}
