package models

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PatchRun : PatchRun Type
type PatchRun struct {
	gorm.Model
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	PatchWindow string    `json:"patch_window" binding:"required"`
	StartTime   time.Time `binding:"required" time_format:"2006-01-02T15:04"`
	EndTime     time.Time `binding:"required" time_format:"2006-01-02T15:04"`
	ChatRooms   ChatRooms `json:"chat_roooms,omitempty" gorm:"many2many:patchrun_ChatRooms;"`
}

// PatchRuns is a list of PatchRun object pointers
type PatchRuns []*PatchRun

// NewPatchRun returns a new PuppetServer object
func NewPatchRun() (p *PatchRun) {
	p = new(PatchRun)
	// Defaults
	now := time.Now()
	p.StartTime = now
	p.EndTime = now
	return
}

// Init : Create new PatchRun object
func (p *PatchRun) Init() (err error) {
	p.Name = generatePatchRunName()
	return GetDB().Create(p).Error
}

// Save : Save PatchRun object
func (p *PatchRun) Save() (err error) {
	return GetDB().Save(p).Error
}

// Delete : Delete PatchRun object
func (p *PatchRun) Delete(cascade bool) (err error) {
	if cascade {
		for _, app := range p.GetApplications() {
			err = app.Delete(cascade)
			if err != nil {
				return
			}
		}
		for _, tb := range p.GetTrelloBoards() {
			err = tb.Delete(cascade)
			if err != nil {
				return
			}
		}
	}
	return GetDB().Delete(p).Error
}

// GetPatchRunByID returns patch run object by ID
func GetPatchRunByID(id uint) (patchRun *PatchRun, err error) {
	patchRun = new(PatchRun)
	err = GetDB().Preload("ChatRooms").First(patchRun, id).Error
	return
}

// GetApplications returns a list of apps in this patchrun
func (p *PatchRun) GetApplications() (apps []*Application) {
	apps = make([]*Application, 0)
	GetDB().Where(&Application{PatchRunID: p.ID}).Order("name").Find(&apps)
	return
}

// GetTrelloBoards returns a list of trello boards in this patchrun
func (p *PatchRun) GetTrelloBoards() (boards []*TrelloBoard) {
	return GetTrelloBoards(p.ID)
}

// GetJenkinsJobs returns a list of enabled Jenkins Jobs for a patchRun
func (p *PatchRun) GetJenkinsJobs() (jobs JenkinsJobs) {
	var err error
	jobs = make(JenkinsJobs, 0)
	err = GetDB().Where(&JenkinsJob{IsForPatchRun: true}).Order("name").Find(&jobs).Error
	if err != nil {
		log.Error("ERROR in patchRun.GetJenkinsJobs: ", err)
	}
	return
}

// GetJenkinsBuilds returns a list of enabled Jenkins Builds for a patchRun
func (p *PatchRun) GetJenkinsBuilds() (builds []*JenkinsBuild) {
	var err error
	builds = make([]*JenkinsBuild, 0)
	err = GetDB().Where(&JenkinsBuild{PatchRunID: p.ID}).Preload("JenkinsJob").Find(&builds).Error
	if err != nil {
		log.Error("ERROR in patchRun.GetJenkinsBuilds: ", err)
	}
	return
}

// GetEnabledChatRooms returns a list of enabled ChatRooms
func (p *PatchRun) GetEnabledChatRooms() (rooms ChatRooms) {
	var err error
	rooms = make(ChatRooms, 0)
	err = GetDB().Where(ChatRoom{Enabled: true}).Order("name").Find(&rooms).Error
	if err != nil {
		log.Error("ERROR in patchRun.GetEnabledChatRooms: ", err)
	}
	return
}

// IsChatRoomLinked returns true if ChatRoom ID is linked
func (p *PatchRun) IsChatRoomLinked(id uint) (linked bool) {
	for _, room := range p.ChatRooms {
		if room.ID == id {
			return true
		}
	}
	return
}

// LinkChatRooms returns a list of enabled ChatRooms
func (p *PatchRun) LinkChatRooms(rooms ChatRooms) (err error) {
	err = GetDB().Model(p).Association("ChatRooms").Replace(rooms)
	if err != nil {
		log.Error("ERROR in patchRun.LinkChatRooms: ", err)
	}
	return
}

// GetServers returns a list of servers patchrun
func (p *PatchRun) GetServers() (servers []string) {
	for _, app := range p.GetApplications() {
		for _, env := range app.GetEnvironments() {
			for _, component := range env.GetComponents() {
				servers = append(servers, component.GetServerList()...)
			}
		}
	}
	return
}

// GetServersCommaSeparated returns a list of servers patchrun
func (p *PatchRun) GetServersCommaSeparated() (servers string) {
	return strings.Join(p.GetServers(), ",")
}

// CreatePatchRun makes a patchrun object
func CreatePatchRun(name string, patchWindow string) (patchRun *PatchRun) {
	patchRun = new(PatchRun)
	err := patchRun.Init()
	if err != nil {
		log.Error("ERROR in patchRun.Init: ", err)
	}
	if name != "" {
		patchRun.Name = name
	}
	patchRun.PatchWindow = patchWindow
	err = patchRun.Save()
	if err != nil {
		log.Error("ERROR in patchRun.Save: ", err)
	}
	return
}

// GetPatchRuns returns a list of patch runs
func GetPatchRuns() (patchRuns PatchRuns) {
	patchRuns = make(PatchRuns, 0)
	GetDB().Order("updated_at desc").Find(&patchRuns)
	return
}

// GetLatestPatchRunID will return the most recent Patch Run ID
func GetLatestPatchRunID() (id uint, err error) {
	patchRun := new(PatchRun)
	err = GetDB().Select("id").Last(patchRun).Error
	if err != nil {
		log.Error("Error retrieving latest PatchRun ID: ", err)
		return
	}
	id = patchRun.ID
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (p *PatchRun) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, PatchRuns{}.GetBreadCrumbs()...) // Patch Run List
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("PatchRun: %s", p.Name), fmt.Sprintf("/patchRun/%v", p.ID)))
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (patchRuns PatchRuns) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
	breadcrumbs = append(breadcrumbs, createBreadCrumb("Patch Runs", "/patchRun"))
	return
}

func generatePatchRunName() (name string) {
	year, week := time.Now().ISOWeek()
	name = fmt.Sprintf("Patching: %v-W%v", year, week)
	return
}
