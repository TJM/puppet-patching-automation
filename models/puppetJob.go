package models

import (
	"gorm.io/gorm"
)

// PuppetJob defines a Puppet Job - A Puppet Job is what is created when a puppet deploy, task, task_plan, etc is "run"
type PuppetJob struct {
	gorm.Model
	Name             string
	Status           string
	ConsoleURL       string        // Puppet Console Link (manually generated)
	APIJobID         string        // Orchestrator Job ID (Remote)
	APIJobURL        string        // Orchestrator Job ID Link
	InitiatorID      uint          // Patching Automation Initiator ID
	InitiatorType    string        // Patching Automation Initator Type
	PuppetTaskID     uint          // Parent Puppet Task ID
	PuppetServerID   uint          // Parent Puppet Server ID
	PuppetParentID   uint          // Puppet Parent ID
	PuppetParentType string        // Puppet Parent Type (deploy, task or task_plan)
	PuppetServer     *PuppetServer `json:"-" yaml:"-" xml:"-" form:"-"` // Parent Puppet Server
	// PuppetParent     interface{}   `json:"-" yaml:"-" xml:"-" form:"-"` // Parent Puppet Object
}

// NewPuppetJob returns a new PuppetJob object
func NewPuppetJob() (j *PuppetJob) {
	j = new(PuppetJob)
	// Defaults
	return
}

// Init : Create new PatchRun object
func (j *PuppetJob) Init() error {
	return GetDB().Create(j).Error
}

// Save : Save PatchRun object
func (j *PuppetJob) Save() error {
	return GetDB().Save(j).Error
}

// Delete : Delete PatchRun object
func (j *PuppetJob) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// No Child Objects yet
	// }
	return GetDB().Delete(j).Error // TODO: Catch Error on delete from DB
}

// GetPuppetJobByID returns patch run object by ID
func GetPuppetJobByID(id uint) (j *PuppetJob, err error) {
	j = new(PuppetJob)
	err = GetDB().Preload("Params").First(j, id).Error
	return
}

// GetPuppetJobs returns a list of all PuppetJobs
func GetPuppetJobs() (jobs []*PuppetJob) {
	jobs = make([]*PuppetJob, 0)
	GetDB().Preload("PuppetServer").Order("name").Find(&jobs)
	return
}
