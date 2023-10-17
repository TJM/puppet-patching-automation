package models

import (
	"errors"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Component : Application Component Type
type Component struct {
	gorm.Model
	Name              string  `json:"name"`
	PatchingProcedure string  `json:"patching_procedure"`
	HealthCheckScript string  `json:"healthcheck_script"`
	Servers           Servers `json:"servers"`
	TrelloChecklistID string  `json:"trello_checklist_id"`
	EnvironmentID     uint    `json:"environment_id"`
	//Environment       *Environment
}

// Components - List of Components
type Components []*Component

// Init : Create new component object
func (c *Component) Init() {
	c.Servers = make(Servers, 0)
	GetDB().Create(c)
}

// Save : Save component object
func (c *Component) Save() {
	GetDB().Save(c)
}

// Delete component
func (c *Component) Delete(cascade bool) (err error) {
	if cascade {
		for _, server := range c.GetServers() {
			err = server.Delete(cascade)
			if err != nil {
				return
			}
		}
		for _, puppetJob := range c.GetPuppetJobs(0, "") { // NOTE "zero" value matches anything in GoRM
			fmt.Printf("Job to be deleted: %+v", puppetJob)
			err = puppetJob.Delete(cascade)
			if err != nil {
				return
			}
		}
	}
	GetDB().Delete(c) // TODO: Catch Error Deleting
	return
}

// GetComponentByID : Return an environment object by ID
func GetComponentByID(id uint) (c *Component, err error) {
	c = new(Component)
	err = GetDB().First(c, id).Error
	return
}

// Server : Return server object by name (create if not exist)
func (c *Component) Server(name string) (server *Server) {
	server = new(Server)
	GetDB().Where(Server{Name: name, ComponentID: c.ID}).FirstOrCreate(server)
	//server.Component = component
	c.Servers = append(c.Servers, server)
	return
}

// GetServer : Return server object by name
func (c *Component) GetServer(name string) (server *Server) {
	server = new(Server)
	result := GetDB().Where(Server{Name: name, ComponentID: c.ID}).First(server)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return
}

// GetServerList : Return list of servers, sorted by name
func (c *Component) GetServerList() (names []string) {
	names = make([]string, 0, len(c.Servers))
	for _, server := range c.GetServers() {
		names = append(names, server.Name)
	}
	sort.Strings(names)
	return
}

// GetServers : Return all Server sorted by name
func (c *Component) GetServers() (servers Servers) {
	if len(c.Servers) == 0 {
		servers = make(Servers, 0)
		GetDB().Where(&Server{ComponentID: c.ID}).Model(&Server{}).Order("name").Find(&servers)
		c.Servers = servers
	}
	return c.Servers
}

// GetServersOnPuppetServer : Return Servers Associated to puppetServer sorted by name
func (c *Component) GetServersOnPuppetServer(puppetServerID uint) (servers Servers) {
	servers = make(Servers, 0)
	GetDB().Where(&Server{ComponentID: c.ID, PuppetServerID: puppetServerID}).Order("name").Find(&servers)
	return
}

// GetPuppetTasks returns a list of puppet tasks for this component
func (c *Component) GetPuppetTasks(puppetServerID uint) (tasks PuppetTasks) {
	tasks = make(PuppetTasks, 0)
	err := GetDB().Model(&PuppetServer{Model: gorm.Model{ID: puppetServerID}}).Where(&PuppetTask{Enabled: true, IsForComponent: true}).Distinct().Order("name").Association("PuppetTasks").Find(&tasks)
	if err != nil {
		log.WithField("componentID", c.ID).Error("Error Retrieving PuppetTasks from Component")
	}
	return
}

// GetPuppetPlans returns a list of puppet plans for this component
func (c *Component) GetPuppetPlans(puppetServerID uint) (plans PuppetPlans) {
	plans = make(PuppetPlans, 0)
	err := GetDB().Model(&PuppetServer{Model: gorm.Model{ID: puppetServerID}}).Where(&PuppetPlan{Enabled: true, IsForComponent: true}).Distinct().Order("name").Association("PuppetPlans").Find(&plans)
	if err != nil {
		log.WithField("componentID", c.ID).Error("Error Retrieving PuppetPlans from Component", err)
	}
	return
}

// GetPuppetServers returns a list of PuppetServers from the component's servers
func (c *Component) GetPuppetServers() (puppetServers PuppetServers) {
	puppetServers = make(PuppetServers, 0)
	err := GetDB().Model(c.GetServers()).Association("PuppetServer").Find(&puppetServers)
	if err != nil {
		log.WithField("c.ID", c.ID).Error("Error retrieving puppetServers from component:", err)
	}
	return
}

// GetPuppetJobs returns a list of PuppetJobs for this component
func (c *Component) GetPuppetJobs(puppetServerID uint, kind string) (jobs []*PuppetJob) {
	jobs = make([]*PuppetJob, 0)
	GetDB().Where(&PuppetJob{PuppetServerID: puppetServerID, InitiatorID: c.ID, InitiatorType: "Component", PuppetParentType: kind}).Find(&jobs)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (c *Component) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	env, err := GetEnvironmentByID(c.EnvironmentID)
	if err != nil {
		log.Error("Error getting breadcrumbs: " + err.Error())
		return
	}
	breadcrumbs = append(breadcrumbs, Components{}.GetBreadCrumbs(env)...)                                                        // Component List
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("Component: %s", c.Name), fmt.Sprintf("/component/%v", c.ID))) // SELF
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (components Components) GetBreadCrumbs(env *Environment) (breadcrumbs BreadCrumbs) {
	app, err := GetApplicationByID(env.ApplicationID)
	if err != nil {
		log.Warn("Problem Accessing Application for breadcrumbs")
		breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
	} else {
		breadcrumbs = append(breadcrumbs, Environments{}.GetBreadCrumbs(app)...)
	}
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("%s (%s): Components", app.Name, env.Name), fmt.Sprintf("/environment/%v/components", env.ID))) // SELF
	return
}
