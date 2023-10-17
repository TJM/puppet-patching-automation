package models

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Environment : Application Environment Type
type Environment struct {
	gorm.Model
	Name              string     `json:"name"`
	PatchingProcedure string     `json:"patching_procedure"`
	Components        Components `json:"components"`
	TrelloCardID      string     `json:"trello_card_id"`
	TrelloCardURL     string     `json:"trello_card_url"`
	ApplicationID     uint
	//Application       *Application
}

// Environments : List of Environments
type Environments []*Environment

// Init : Create New Environment
func (e *Environment) Init() {
	e.Components = make(Components, 0)
	GetDB().Create(e)
}

// Save : Save Environment
func (e *Environment) Save() {
	GetDB().Save(e)
}

// Delete environment
func (e *Environment) Delete(cascade bool) (err error) {
	if cascade {
		for _, component := range e.GetComponents() {
			err = component.Delete(cascade)
			if err != nil {
				return
			}
		}
	}
	GetDB().Delete(e) // TODO: Catch Error Deleting
	return
}

// GetEnvironmentByID : Return an environment object by ID
func GetEnvironmentByID(id uint) (e *Environment, err error) {
	e = new(Environment)
	err = GetDB().First(e, id).Error
	return
}

// Component : Return an component object by name (create if not exist)
func (e *Environment) Component(name string) (component *Component) {
	component = new(Component)
	GetDB().Where(Component{Name: name, EnvironmentID: e.ID}).FirstOrCreate(component)
	//component.Environment = e
	e.Components = append(e.Components, component)
	return
}

// GetComponent : Return an component object by name
func (e *Environment) GetComponent(name string) (component *Component) {
	component = new(Component)
	result := GetDB().Where(Component{Name: name, EnvironmentID: e.ID}).First(component)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return
}

// GetComponents : Return all Component sorted by name
func (e *Environment) GetComponents() (components Components) {
	components = make(Components, 0)
	GetDB().Where(&Component{EnvironmentID: e.ID}).Order("name").Find(&components)
	return
}

// GetComponentsAndServers : Return all Component sorted by name, with their servers
func (e *Environment) GetComponentsAndServers() (components Components) {
	components = make(Components, 0)
	GetDB().Preload("Servers").Where(&Component{EnvironmentID: e.ID}).Order("name").Find(&components)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (e *Environment) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	app, err := GetApplicationByID(e.ApplicationID)
	if err != nil {
		log.Error("Error getting breadcrumbs: " + err.Error())
		return
	}
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("App: %s", app.Name), fmt.Sprintf("/application/%v", app.ID)))
	breadcrumbs = append(breadcrumbs, app.GetBreadCrumbs()...)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (envs Environments) GetBreadCrumbs(app *Application) (breadcrumbs BreadCrumbs) {
	patchRun, err := GetPatchRunByID(app.PatchRunID)
	if err != nil {
		log.Warn("Unable to retrieve patchRun for application breadcrumb.")
		breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
	} else {
		breadcrumbs = append(breadcrumbs, Applications{}.GetBreadCrumbs(patchRun)...)
	}
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("%s: Environments", app.Name), fmt.Sprintf("/application/%v/environments", app.ID)))
	return
}
