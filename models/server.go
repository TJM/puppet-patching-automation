package models

import (
	"gorm.io/gorm"
)

// Server : Server Details Object
type Server struct {
	gorm.Model
	Name              string   `json:"name"`
	IPAddress         string   `json:"ip"`
	VMName            string   `json:"vm_name"`
	Notes             string   `json:"notes"`
	OperatingSystem   string   `json:"operating_system"`
	OSVersion         string   `json:"os_version"`
	PackageUpdates    int      `json:"package_updates"`
	PatchWindow       string   `json:"patch_window"`
	PinnedPackages    []string `json:"pinned_packages" gorm:"-"`
	SecurityUpdates   int      `json:"security_updates"`
	UUID              string   `json:"uuid"`
	TrelloItemID      string   `json:"trello_item_id"`
	TrelloChecklistID string   `json:"trello_checklist_id"`
	TrelloCardID      uint
	ComponentID       uint
	PuppetServerID    uint
	PuppetServer      *PuppetServer
}

// Servers - List of Servers
type Servers []*Server

// Init - Create new server
func (s *Server) Init() {
	s.PinnedPackages = make([]string, 0)
	GetDB().Create(s)
}

// Save - Save server
func (s *Server) Save() {
	GetDB().Save(s)
}

// Delete server
func (s *Server) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// server currently has no child objects
	// 	// I suppose we could remove the checklist items from a trello board,
	// 	// but I think this will only be used in a wholesale delete, in which case
	// 	// the entire trello board will be deleted anyhow.
	// }
	GetDB().Delete(s) // TODO: Catch Error Deleting
	return
}

// GetServerByID : Return a Server object by ID
func GetServerByID(id uint) (server *Server, err error) {
	server = new(Server)
	err = GetDB().First(server, id).Error
	return
}

// GetBreadCrumbs for a Server - NOT CURRENTLY USED
func (s Server) GetBreadCrumbs() BreadCrumbs {
	return GetDefaultBreadCrumbs()
}

// GetBreadCrumbs for a list of servers - NOT CURRENTLY USED
func (servers Servers) GetBreadCrumbs(component Component) BreadCrumbs {
	return GetDefaultBreadCrumbs()
}
