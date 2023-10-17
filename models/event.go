package models

import (
	"net/url"

	"gorm.io/gorm"
)

// Action is the type of the event that happend
type Action string

// Event defines an Event (that we may store/notify)
type Event struct {
	gorm.Model
	Name      string
	Action    Action
	Message   string
	ThreadKey string
	PatchRun  *PatchRun
	Target    interface{}
	URL       *url.URL
	Sent      bool
}

// All possible Actions for use in events
const (
	ActionTest                Action = "TEST_MESSAGE"
	ActionPatchRunCreated     Action = "PATCH_RUN_CREATED"
	ActionPatchRunUpdated     Action = "PATCH_RUN_UPDATED"
	ActionPatchRunDeleted     Action = "PATCH_RUN_DELETED"
	ActionTrelloBoardCreated  Action = "TRELLO_BOARD_CREATED"
	ActionTrelloBoardDeleted  Action = "TRELLO_BOARD_DELETED"
	ActionJenkinsBuildCreated Action = "JENKINS_BUILD_CREATED"
)

// NewEvent returns a new Event object
func NewEvent(action Action) (e *Event) {
	e = new(Event)
	e.Action = action
	// Defaults
	return
}

// Init : Create new PatchRun object
func (e *Event) Init() {
	GetDB().Create(e)
}

// Save : Save PatchRun object
func (e *Event) Save() {
	GetDB().Save(e)
}

// Delete : Delete PatchRun object
func (e *Event) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// No Child Objects yet
	// }
	GetDB().Delete(e) // TODO: Catch Error on delete from DB
	return
}

// GetEventByID returns patch run object by ID
func GetEventByID(id uint) (e *Event) {
	e = new(Event)
	GetDB().First(e, id)
	return
}

// GetEvents returns a list of all Events
func GetEvents() (servers []*Event) {
	servers = make([]*Event, 0)
	GetDB().Order("name").Find(&servers)
	return
}
