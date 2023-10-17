package models

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PuppetTask defines a Puppet Task
type PuppetTask struct {
	gorm.Model
	Name             string
	Description      string
	APITaskID        string // Used by pe-go-client as the "id" - Looks like a URL
	Environment      string
	Enabled          bool
	IsForPatchRun    bool
	IsForApplication bool
	IsForComponent   bool
	IsForServer      bool
	PuppetServers    *[]PuppetServer    `gorm:"many2many:puppetserver_tasks" json:"-" yaml:"-" xml:"-" form:"-"` // Parent Puppet Server(s)
	Params           []*PuppetTaskParam `json:"-" yaml:"-" xml:"-" form:"-"`                                     // Params are owned by a PuppetTask
}

// PuppetTasks is a list of PuppetTask objects
type PuppetTasks []*PuppetTask

// NewPuppetTask returns a new PuppetTask object
func NewPuppetTask() (t *PuppetTask) {
	t = new(PuppetTask)
	// Defaults
	t.Enabled = true
	t.IsForServer = true
	return
}

// Init : Create new PatchRun object
func (t *PuppetTask) Init() error {
	return GetDB().Create(t).Error
}

// Save : Save PatchRun object
func (t *PuppetTask) Save() error {
	return GetDB().Save(t).Error
}

// Delete : Delete PatchRun object
func (t *PuppetTask) Delete(cascade bool) (err error) {
	if cascade {
		var params []*PuppetTaskParam
		params, err = t.GetParams()
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
		// TODO: Should we delete PuppetTaskJob(s) ... currently not doing it
	}
	return GetDB().Delete(t).Error // TODO: Catch Error on delete from DB
}

// GetPuppetTaskByID returns patch run object by ID
func GetPuppetTaskByID(id uint) (t *PuppetTask, err error) {
	t = new(PuppetTask)
	err = GetDB().Preload("Params").First(t, id).Error
	return
}

// GetPuppetServers : Get PuppetServers for this job
func (t *PuppetTask) GetPuppetServers() (puppetServers PuppetServers, err error) {
	err = GetDB().Model(t).Association("PuppetServers").Find(&puppetServers)
	return
}

// AddPuppetServer : Add PuppetServer for this Task
func (t *PuppetTask) AddPuppetServer(puppetServer *PuppetServer) (err error) {
	err = GetDB().Model(t).Association("PuppetServers").Append(puppetServer)
	return
}

// RemovePuppetServer : Remove PuppetServer from this Task
func (t *PuppetTask) RemovePuppetServer(puppetServer *PuppetServer) (err error) {
	err = GetDB().Model(t).Association("PuppetServers").Delete(puppetServer)
	return
}

// GetParams : Get PuppetTaskParams for this job
func (t *PuppetTask) GetParams() (params []*PuppetTaskParam, err error) {
	err = GetDB().Model(t).Association("Params").Find(&params)
	return
}

// GetParamByName : Get PuppetTaskParam by Name for this job
func (t *PuppetTask) GetParamByName(paramName string) (param *PuppetTaskParam, err error) {
	param = NewPuppetTaskParam()
	err = GetDB().Model(t).Where(&PuppetTaskParam{Name: paramName}).Association("Params").Find(&param)
	return
}

// Param : Return a PuppetTaskParam object by name (create if not exist)
func (t *PuppetTask) Param(name string) (param *PuppetTaskParam) {
	param = new(PuppetTaskParam)
	GetDB().Where(PuppetTaskParam{Name: name, PuppetTaskID: t.ID}).FirstOrCreate(param)
	return
}

// AddParam : Add PuppetTaskParam for this job
func (t *PuppetTask) AddParam(param *PuppetTaskParam) (err error) {
	err = GetDB().Model(t).Association("Params").Append(param)
	return
}

// // AddParams : Add PuppetTaskParam for this job
// func (t *PuppetTask) AddParams(params []*PuppetTaskParam) (err error) {
// 	GetDB().Model(t).Association("Params").Append(params) // TODO catch errors?
// 	return
// }

// IsEnabledPuppetTasks returns a bool if there are enabled PuppetTasks
func IsEnabledPuppetTasks() (enabled bool) {
	var count int64
	GetDB().Model(&PuppetTask{}).Where(&PuppetTask{Enabled: true}).Count(&count)
	return count > 0
}

// GetPuppetTasks returns a list of all PuppetTasks
func GetPuppetTasks() (servers PuppetTasks) {
	servers = make(PuppetTasks, 0)
	GetDB().Preload("PuppetServers").Order("name").Find(&servers)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
// Optional parameter: puppetServer (passed up to list)
func (t *PuppetTask) GetBreadCrumbs(puppetServer ...*PuppetServer) (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, PuppetTasks{}.GetBreadCrumbs(puppetServer...)...) // Patch Run List
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("Puppet Task: %s", t.Name), fmt.Sprintf("/config/puppetTask/%v", t.ID)))
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
// Optional parameter: puppetServer
func (tasks PuppetTasks) GetBreadCrumbs(puppetServer ...*PuppetServer) (breadcrumbs BreadCrumbs) {
	if len(puppetServer) > 0 {
		s := puppetServer[0]
		breadcrumbs = append(breadcrumbs, s.GetBreadCrumbs()...)
		breadcrumbs = append(breadcrumbs, createBreadCrumb("Puppet Tasks for: "+s.Name, fmt.Sprintf("/config/puppetServer/%v/tasks", s.ID)))
	} else {
		breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
		breadcrumbs = append(breadcrumbs, createBreadCrumb("Puppet Tasks", "/config/puppetTask"))
	}
	return
}
