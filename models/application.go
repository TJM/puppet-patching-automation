package models

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Application : Application
type Application struct {
	gorm.Model
	Name              string         `json:"name"`
	PatchingProcedure string         `json:"patching_procedure"`
	PatchRunID        uint           `gorm:"index"`
	Environments      []*Environment `json:"environments"`
}

// Applications - List of Application
type Applications []*Application

// Init : Create new ApplictionType
func (a *Application) Init() {
	a.Environments = make([]*Environment, 0)
	GetDB().Create(a)
}

// Save : Create new ApplictionType
func (a *Application) Save() {
	GetDB().Save(a)
}

// Delete application
func (a *Application) Delete(cascade bool) (err error) {
	if cascade {
		for _, env := range a.GetEnvironments() {
			err = env.Delete(cascade)
			if err != nil {
				return
			}
		}
	}
	GetDB().Delete(a) // TODO: Catch Error Deleting
	return
}

// GetApplicationByID : Return an application object by ID
func GetApplicationByID(id uint) (a *Application, err error) {
	a = new(Application)
	err = GetDB().First(a, id).Error
	return
}

// GetOrCreateApplication : Return an application object by name (create one if not exist)
func GetOrCreateApplication(name string, patchRunID uint) (a *Application) {
	a = new(Application)
	GetDB().Where(Application{Name: name, PatchRunID: patchRunID}).FirstOrCreate(a)
	return
}

// GetApplication : Return an application object by name and patchRunID
func GetApplication(name string, patchRunID uint) (a *Application) {
	a = new(Application)
	result := GetDB().Where(Application{Name: name, PatchRunID: patchRunID}).First(a)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return
}

// GetApplications : Return all applications sorted by name
func GetApplications(patchRunID uint) (apps Applications) {
	apps = make(Applications, 0)
	GetDB().Where(Application{PatchRunID: patchRunID}).Order("name").Find(&apps)
	return
}

// Environment : Return an environment object by name (create if not exist)
func (a *Application) Environment(name string) (env *Environment) {
	env = new(Environment)
	GetDB().Where(Environment{Name: name, ApplicationID: a.ID}).FirstOrCreate(env)
	//env.Application = a
	a.Environments = append(a.Environments, env)
	return
}

// GetEnvironment : Return an environment object by name
func (a *Application) GetEnvironment(name string) (env *Environment) {
	env = new(Environment)
	result := GetDB().Where(Environment{Name: name, ApplicationID: a.ID}).First(env)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return
}

// GetPatchRunName : Return an environment object by name
func (a *Application) GetPatchRunName() (name string) {
	patchRun, err := GetPatchRunByID(a.PatchRunID)
	if err != nil {
		log.Error("Error in GetPatchRunName: " + err.Error())
		return "Not Found"
	}
	return patchRun.Name
}

// GetEnvironments : Return all environments sorted by name
func (a *Application) GetEnvironments() (envs Environments) {
	envs = make([]*Environment, 0)
	GetDB().Where(&Environment{ApplicationID: a.ID}).Order("name").Find(&envs)
	return
}

// GetEnvironmentList : Simple list of applications by name
func (a *Application) GetEnvironmentList() (names []string) {
	names = make([]string, 0)
	GetDB().Model(&Environment{}).Where(&Environment{ApplicationID: a.ID}).Order("name").Select([]string{"name"}).Find(&names)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (a *Application) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	patchRun, err := GetPatchRunByID(a.PatchRunID)
	if err != nil {
		log.Error("Error getting breadcrumbs: " + err.Error())
		return
	}
	breadcrumbs = append(breadcrumbs, Applications{}.GetBreadCrumbs(patchRun)...)                                                     // Application List
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("Application: %s", a.Name), fmt.Sprintf("/application/%v", a.ID))) // We don't navigate here
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (apps Applications) GetBreadCrumbs(patchRun *PatchRun) (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, patchRun.GetBreadCrumbs()...)
	breadcrumbs = append(breadcrumbs, createBreadCrumb("Applications", fmt.Sprintf("/patchRun/%v/applications", patchRun.ID)))
	return
}
