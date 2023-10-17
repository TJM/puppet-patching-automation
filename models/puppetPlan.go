package models

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PuppetPlan defines a Puppet Job
type PuppetPlan struct {
	gorm.Model
	Name             string
	Description      string
	APIPlanID        string // Used by pe-go-client as the "id" - Looks like a URL
	Environment      string
	Enabled          bool
	IsForPatchRun    bool
	IsForApplication bool
	IsForComponent   bool
	IsForServer      bool
	PuppetServers    *[]PuppetServer    `gorm:"many2many:puppetserver_plans" json:"-" yaml:"-" xml:"-" form:"-"` // Parent Puppet Server(s)
	Params           []*PuppetPlanParam `json:"-" yaml:"-" xml:"-" form:"-"`                                     // Params are owned by a PuppetPlan
}

// PuppetPlans is a list of PuppetPlan objects
type PuppetPlans []*PuppetPlan

// NewPuppetPlan returns a new PuppetPlan object
func NewPuppetPlan() (p *PuppetPlan) {
	p = new(PuppetPlan)
	// Defaults
	p.Enabled = true
	p.IsForServer = true
	return
}

// Init : Create new PatchRun object
func (p *PuppetPlan) Init() error {
	return GetDB().Create(p).Error
}

// Save : Save PatchRun object
func (p *PuppetPlan) Save() error {
	return GetDB().Save(p).Error
}

// Delete : Delete PatchRun object
func (p *PuppetPlan) Delete(cascade bool) (err error) {
	if cascade {
		var params []*PuppetPlanParam
		params, err = p.GetParams()
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
		// TODO: Should we delete PuppetPlanJob(s) ... currently not doing it
	}
	return GetDB().Delete(p).Error // TODO: Catch Error on delete from DB
}

// GetPuppetPlanByID returns patch run object by ID
func GetPuppetPlanByID(id uint) (p *PuppetPlan, err error) {
	p = new(PuppetPlan)
	err = GetDB().Preload("Params").First(p, id).Error
	return
}

// GetPuppetServers : Get PuppetServers for this Plan
func (p *PuppetPlan) GetPuppetServers() (puppetServers PuppetServers, err error) {
	err = GetDB().Model(p).Association("PuppetServers").Find(&puppetServers)
	return
}

// AddPuppetServer : Add PuppetServer for this Plan
func (p *PuppetPlan) AddPuppetServer(puppetServer *PuppetServer) (err error) {
	err = GetDB().Model(p).Association("PuppetServers").Append(puppetServer)
	return
}

// RemovePuppetServer : Remove PuppetServer from this Plan
func (p *PuppetPlan) RemovePuppetServer(puppetServer *PuppetServer) (err error) {
	err = GetDB().Model(p).Association("PuppetServers").Delete(puppetServer)
	return
}

// GetParams : Get PuppetPlanParams for this job
func (p *PuppetPlan) GetParams() (params []*PuppetPlanParam, err error) {
	err = GetDB().Model(p).Order("Name").Association("Params").Find(&params)
	return
}

// GetParamByName : Get PuppetPlanParam by Name for this job
func (p *PuppetPlan) GetParamByName(paramName string) (param *PuppetPlanParam, err error) {
	param = NewPuppetPlanParam()
	err = GetDB().Model(p).Where(&PuppetPlanParam{Name: paramName}).Association("Params").Find(&param)
	return
}

// Param : Return a PuppetPlanParam object by name (create if not exist)
func (p *PuppetPlan) Param(name string) (param *PuppetPlanParam) {
	param = new(PuppetPlanParam)
	GetDB().Where(PuppetPlanParam{Name: name, PuppetPlanID: p.ID}).FirstOrCreate(param)
	return
}

// AddParam : Add PuppetPlanParam for this job
func (p *PuppetPlan) AddParam(param *PuppetPlanParam) (err error) {
	err = GetDB().Model(p).Association("Params").Append(param)
	return
}

// // AddParams : Add PuppetPlanParam for this job
// func (p *PuppetPlan) AddParams(params []*PuppetPlanParam) (err error) {
// 	GetDB().Model(p).Association("Params").Append(params) // TODO catch errors?
// 	return
// }

// IsEnabledPuppetPlans returns a bool if there are enabled PuppetPlans
func IsEnabledPuppetPlans() (enabled bool) {
	var count int64
	GetDB().Model(&PuppetPlan{}).Where(&PuppetPlan{Enabled: true}).Count(&count)
	return count > 0
}

// GetPuppetPlans returns a list of all PuppetPlans
func GetPuppetPlans() (servers PuppetPlans) {
	servers = make(PuppetPlans, 0)
	GetDB().Preload("PuppetServers").Order("name").Find(&servers)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
// Optional parameter: puppetServer (passed up to list)
func (p *PuppetPlan) GetBreadCrumbs(puppetServer ...*PuppetServer) (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, PuppetPlans{}.GetBreadCrumbs(puppetServer...)...) // Patch Run List
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("Puppet Plan: %s", p.Name), fmt.Sprintf("/config/puppetTask/%v", p.ID)))
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
// Optional parameter: puppetServer
func (plans PuppetPlans) GetBreadCrumbs(puppetServer ...*PuppetServer) (breadcrumbs BreadCrumbs) {
	if len(puppetServer) > 0 {
		s := puppetServer[0]
		breadcrumbs = append(breadcrumbs, s.GetBreadCrumbs()...)
		breadcrumbs = append(breadcrumbs, createBreadCrumb("Puppet Plans for: "+s.Name, fmt.Sprintf("/config/puppetServer/%v/plans", s.ID)))
	} else {
		breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
		breadcrumbs = append(breadcrumbs, createBreadCrumb("Puppet Plans", "/config/puppetPlan"))
	}
	return
}
