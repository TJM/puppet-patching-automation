package models

import (
	"fmt"

	"gorm.io/gorm"
)

// ChatRoom defines a ChatRoom
type ChatRoom struct {
	gorm.Model
	Name        string `binding:"required"`
	Description string
	WebhookURL  string `binding:"required,url"`
	Enabled     bool
}

// ChatRooms is a list of ChatRoom
type ChatRooms []*ChatRoom

// NewChatRoom returns a new ChatRoom object
func NewChatRoom() (r *ChatRoom) {
	r = new(ChatRoom)
	// Defaults
	r.Enabled = true
	return
}

// Init : Create new ChatRoom object in DB
func (r *ChatRoom) Init() {
	GetDB().Create(r)
}

// Save : Save ChatRoom object
func (r *ChatRoom) Save() {
	GetDB().Save(r)
}

// Delete : Delete PatchRun object
func (r *ChatRoom) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// no child objects (yet)
	// }
	err = GetDB().Delete(r).Error
	return
}

// GetChatRoomByID returns patch run object by ID
func GetChatRoomByID(id uint) (r *ChatRoom, err error) {
	r = new(ChatRoom)
	err = GetDB().First(r, id).Error
	return
}

// GetChatRoomsByIDs returns patch run object by ID
func GetChatRoomsByIDs(ids []uint) (rooms ChatRooms, err error) {
	rooms = make(ChatRooms, 0)
	if len(ids) == 0 {
		return
	}
	GetDB().Order("name").Find(&rooms, ids)
	return
}

// GetEnabledChatRooms returns a list of enabled ChatRooms
func GetEnabledChatRooms() (rooms ChatRooms, err error) {
	rooms = make(ChatRooms, 0)
	err = GetDB().Where(&ChatRoom{Enabled: true}).Order("name").Find(&rooms).Error
	return
}

// GetChatRooms returns a list of all ChatRooms
func GetChatRooms() (rooms ChatRooms) {
	rooms = make(ChatRooms, 0)
	GetDB().Order("name").Find(&rooms)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (r *ChatRoom) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, ChatRooms{}.GetBreadCrumbs()...) // Patch Run List
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("ChatRoom: %s", r.Name), fmt.Sprintf("/config/ChatRoom/%v", r.ID)))
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (rooms ChatRooms) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
	breadcrumbs = append(breadcrumbs, createBreadCrumb("Chat Rooms", "/config/ChatRoom"))
	return
}
