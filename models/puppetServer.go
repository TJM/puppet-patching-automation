package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/puppetlabs/go-pe-client/pkg/orch"
	"github.com/puppetlabs/go-pe-client/pkg/pe"
	"github.com/puppetlabs/go-pe-client/pkg/puppetdb"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PuppetServer defines a Puppet Server
type PuppetServer struct {
	gorm.Model
	Name          string `binding:"required"`
	Description   string
	Hostname      string `binding:"required,fqdn"`
	PuppetDBPort  uint   `binding:"required,numeric,gte=1024,lte=65535"`
	OrchPort      uint   `binding:"required,numeric,gte=1024,lte=65535"`
	RBACPort      uint   `binding:"required,numeric,gte=1024,lte=65535"`
	Token         string // `binding:"required"` - This can't be "required"
	SSL           bool
	SSLSkipVerify bool
	CACert        string
	Enabled       bool
	FactName      string           `binding:"required"` // TODO: Validate FactString
	PuppetTasks   PuppetTasks      `gorm:"many2many:puppetserver_tasks" json:"-" yaml:"-" xml:"-" form:"-"`
	PuppetPlans   PuppetPlans      `gorm:"many2many:puppetserver_plans" json:"-" yaml:"-" xml:"-" form:"-"`
	OrchClient    *orch.Client     `gorm:"-" json:"-" yaml:"-" xml:"-" form:"-"`
	PDBClient     *puppetdb.Client `gorm:"-" json:"-" yaml:"-" xml:"-" form:"-"`
	PEClient      *pe.Client       `gorm:"-" json:"-" yaml:"-" xml:"-" form:"-"`
}

// PuppetServers is a list of PuppetServer objects
type PuppetServers []*PuppetServer

// NewPuppetServer returns a new PuppetServer object
func NewPuppetServer() (p *PuppetServer) {
	p = new(PuppetServer)
	// Defaults
	p.PuppetDBPort = 8081
	p.RBACPort = 4443
	p.OrchPort = 8143
	p.SSL = true
	p.Enabled = true
	p.FactName = "pe_patch.patch_group"
	return
}

// Init : Create new PatchRun object
func (p *PuppetServer) Init() {
	GetDB().Create(p)
}

// Save : Save PatchRun object
func (p *PuppetServer) Save() {
	GetDB().Save(p)
}

// Delete : Delete PatchRun object
func (p *PuppetServer) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// No Child Objects yet
	// }
	GetDB().Delete(p) // TODO: Catch Error on delete from DB
	return
}

// GetPuppetServerByID returns patch run object by ID
func GetPuppetServerByID(id uint) (p *PuppetServer, err error) {
	p = new(PuppetServer)
	err = GetDB().First(p, id).Error
	p.setDefaults() // Finds "missing" attributes and sets them
	return
}

// GetEnabledPuppetServers returns a list of enabled PuppetServers
func GetEnabledPuppetServers() (servers PuppetServers) {
	servers = make(PuppetServers, 0)
	GetDB().Where(&PuppetServer{Enabled: true}).Order("name").Find(&servers)
	for _, server := range servers {
		server.setDefaults()
	}
	return
}

// GetTasks : Get Puppet Tasks
func (p *PuppetServer) GetTasks() (tasks PuppetTasks, err error) {
	err = GetDB().Model(p).Association("PuppetTasks").Find(&tasks)
	return
}

// GetPlans : Get Puppet Plans
func (p *PuppetServer) GetPlans() (plans PuppetPlans, err error) {
	err = GetDB().Model(p).Association("PuppetPlans").Find(&plans)
	return
}

// GetUnassociatedPlans : Get Puppet Plans that are not associated to this PuppetServer, that match names
func (p *PuppetServer) GetUnassociatedPlans(names []string) (plans PuppetPlans, err error) {
	notPlanIDs := make([]uint, 0)
	err = GetDB().Model(p).Select("ID").Association("PuppetPlans").Find(&notPlanIDs)
	if err != nil {
		return
	}
	err = GetDB().Model(PuppetPlan{}).Not(notPlanIDs).Find(&plans).Error
	return
}

// GetUnassociatedTasks : Get Puppet Tasks that are not associated to this PuppetServer, that match names
func (p *PuppetServer) GetUnassociatedTasks(names []string) (tasks PuppetTasks, err error) {
	notTaskIDs := make([]uint, 0)
	err = GetDB().Model(p).Select("ID").Association("PuppetTasks").Find(&notTaskIDs)
	if err != nil {
		return
	}
	log.WithField("names", names).Info("Tasks for names")
	err = GetDB().Model(PuppetTask{}).Not(notTaskIDs).Where("name IN ?", names).Find(&tasks).Error
	return
}

// AddPuppetPlan : Add PuppetPlan to this PuppetServer
func (p *PuppetServer) AddPuppetPlan(plan *PuppetPlan) (err error) {
	err = GetDB().Model(p).Association("PuppetPlans").Append(plan)
	return
}

// RemovePuppetPlan : Remove PuppetPlan from this PuppetServer
func (p *PuppetServer) RemovePuppetPlan(plan *PuppetPlan) (err error) {
	err = GetDB().Model(p).Association("PuppetPlans").Delete(plan)
	return
}

// AddPuppetTask : Add PuppetTask to this PuppetServer
func (p *PuppetServer) AddPuppetTask(task *PuppetTask) (err error) {
	err = GetDB().Model(p).Association("PuppetTasks").Append(task)
	return
}

// RemovePuppetTask : Remove PuppetTask from this PuppetServer
func (p *PuppetServer) RemovePuppetTask(task *PuppetTask) (err error) {
	err = GetDB().Model(p).Association("PuppetTasks").Delete(task)
	return
}

// IsPuppetServersEnabled returns a bool if there are enabled PuppetServers
func IsPuppetServersEnabled() (enabled bool) {
	var count int64
	GetDB().Model(&PuppetServer{}).Where(&PuppetServer{Enabled: true}).Count(&count)
	return count > 0
}

// GetPuppetServers returns a list of all PuppetServers
func GetPuppetServers() (servers PuppetServers) {
	servers = make(PuppetServers, 0)
	GetDB().Order("name").Find(&servers)
	for _, server := range servers {
		server.setDefaults()
	}
	return
}

// GetPuppetDBUrl returns the PuppetDB URL for the Puppet Server
func (p *PuppetServer) GetPuppetDBUrl() (url string) {
	return p.GetBaseURL(p.PuppetDBPort)
}

// GetOrchURL returns the Puppet Orchestrator API URL for the Puppet Server
func (p *PuppetServer) GetOrchURL() (url string) {
	return p.GetBaseURL(p.OrchPort)
}

// GetPEURL returns the Puppet Enterprise API URL for the Puppet Server
func (p *PuppetServer) GetPEURL() (url string) {
	return p.GetBaseURL(8140) // NOTE: hardcoded to 8140
}

// GetConsoleURL returns the Puppet Console URL for the Puppet Server
func (p *PuppetServer) GetConsoleURL() (url string) {
	return p.GetBaseURL(443) // NOTE: hardcoded to 443
}

// GetFactNamePath returns the FactName as a "path" for PuppetDB API
func (p *PuppetServer) GetFactNamePath() (factPath string) {
	factList := strings.Split(p.FactName, ".")
	factJSON, err := json.Marshal(factList)
	if err != nil {
		log.WithField("FactName", p.FactName).Error("Error json-ing the puppetServer.FactName", err)
	}
	factPath = string(factJSON)
	return
}

// setDefaults sets defaults on "existing" entries. This handles adding new attributes.
// NOTE: We cannot handle setting a boolean default to true on existing entries.
func (p *PuppetServer) setDefaults() {
	new := NewPuppetServer()
	changed := false
	if p.FactName == "" {
		p.FactName = new.FactName
		changed = true
	}
	if p.PuppetDBPort == 0 {
		p.PuppetDBPort = new.PuppetDBPort
		changed = true
	}
	if p.RBACPort == 0 {
		p.RBACPort = new.RBACPort
		changed = true
	}
	if p.OrchPort == 0 {
		p.OrchPort = new.OrchPort
		changed = true
	}
	if changed {
		p.Save()
	}
}

// GetBaseURL returns a base URL (up to the first slash)
func (p *PuppetServer) GetBaseURL(port uint) (url string) {
	if p.SSL {
		url = "https://"
	} else {
		url = "http://"
	}
	url += fmt.Sprintf("%s:%v", p.Hostname, port)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (p *PuppetServer) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, PuppetServers{}.GetBreadCrumbs()...) // Patch Run List
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("Puppet Server: %s", p.Name), fmt.Sprintf("/config/puppetServer/%v", p.ID)))
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (servers PuppetServers) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
	breadcrumbs = append(breadcrumbs, createBreadCrumb("Puppet Servers", "/config/puppetServer"))
	return
}
