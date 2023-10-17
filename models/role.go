package models

import (
	"fmt"

	"gorm.io/gorm"
)

// ***** NOTICE: RBAC Role is *not* stored directly in the database.
// *****         The data comes from the casbin "enforcer" (which stores it in the database)

// Role defines an RBAC Role
type Role struct {
	gorm.Model
	Name        string `binding:"required"`
	Users       []string
	Description string
}

// Roles is a list of Role objects
type Roles []*Role

// NewRole returns a new Role object with defaults set
func NewRole() (r *Role) {
	r = new(Role)
	// Defaults
	r.Users = make([]string, 0)
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (r *Role) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, Roles{}.GetBreadCrumbs()...) // Patch Run List
	breadcrumbs = append(breadcrumbs, createBreadCrumb(fmt.Sprintf("Role: %s", r.Name), fmt.Sprintf("/config/role/%v", r.ID)))
	return
}

// GetBreadCrumbs returns a list of bread crumbs for navigation
func (roles Roles) GetBreadCrumbs() (breadcrumbs BreadCrumbs) {
	breadcrumbs = append(breadcrumbs, GetDefaultBreadCrumbs()...)
	breadcrumbs = append(breadcrumbs, createBreadCrumb("Roles", "/config/role"))
	return
}
