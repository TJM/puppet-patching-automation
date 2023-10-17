package models

import (
	"gorm.io/gorm"
)

// TrelloBoard : Application TrelloBoard Type
type TrelloBoard struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Background  string `json:"background"`
	URL         string `json:"url"`
	PatchRunID  uint   `json:"patch_run_id"`
	RemoteID    string `json:"trello_board_id"`
}

// Init : Create new board object
func (tb *TrelloBoard) Init() {
	GetDB().Create(tb)
}

// Save Trello Board to Database
func (tb *TrelloBoard) Save() {
	GetDB().Save(tb)
}

// GetTrelloBoardByID : Return TrelloBoard object by ID
func GetTrelloBoardByID(id uint) (tb *TrelloBoard, err error) {
	tb = new(TrelloBoard)
	err = GetDB().First(tb, id).Error
	return
}

// Delete Trello Board
func (tb *TrelloBoard) Delete(cascade bool) (err error) {
	// if cascade {
	// 	// NOOP - no child objects
	// }
	err = GetDB().Delete(tb).Error // TODO: Check for error deleting from DB?
	return
}

// GetTrelloBoards returns trello boards associated to patchRunID
func GetTrelloBoards(patchRunID uint) (boards []*TrelloBoard) {
	boards = make([]*TrelloBoard, 0)
	GetDB().Where(TrelloBoard{PatchRunID: patchRunID}).Order("name").Find(&boards)
	return
}
