package models

import (
	"gorm.io/gorm"
)

// JenkinsBuild defines a Jenkins Job
type JenkinsBuild struct {
	gorm.Model
	Name            string
	Status          string
	URL             string
	APIBuildID      int64          // Used by gojenkins
	QueueID         int64          //used by gojenkins
	PatchRunID      uint           // Parent Patch Run ID
	JenkinsJobID    uint           // Parent Jenkins Job ID
	JenkinsServerID uint           // Parent Jenkins Server ID
	JenkinsServer   *JenkinsServer `json:"-" yaml:"-" xml:"-" form:"-"` // Parent Jenkins Server
	JenkinsJob      *JenkinsJob    `json:"-" yaml:"-" xml:"-" form:"-"` // Parent Jenkins Build
}

// NewJenkinsBuild returns a new JenkinsBuild object
func NewJenkinsBuild() (j *JenkinsBuild) {
	j = new(JenkinsBuild)
	// Defaults
	return
}

// Init : Create new PatchRun object
func (j *JenkinsBuild) Init() error {
	return GetDB().Create(j).Error
}

// Save : Save PatchRun object
func (j *JenkinsBuild) Save() error {
	return GetDB().Save(j).Error
}

// Delete : Delete PatchRun object
func (j *JenkinsBuild) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// No Child Objects yet
	// }
	return GetDB().Delete(j).Error // TODO: Catch Error on delete from DB
}

// GetJenkinsBuildByID returns patch run object by ID
func GetJenkinsBuildByID(id uint) (j *JenkinsBuild, err error) {
	j = new(JenkinsBuild)
	err = GetDB().Preload("Params").First(j, id).Error
	return
}

// GetJenkinsBuilds returns a list of all JenkinsBuilds
func GetJenkinsBuilds() (servers []*JenkinsBuild) {
	servers = make([]*JenkinsBuild, 0)
	GetDB().Preload("JenkinsServer").Order("name").Find(&servers)
	return
}
