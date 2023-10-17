package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/functions"
	"github.com/tjm/puppet-patching-automation/middleware"
	"github.com/tjm/puppet-patching-automation/models"
)

// ListRoles endpoint (GET) will list all roles
func ListRoles(c *gin.Context) {
	roles := middleware.GetEnforcer().GetAllNamedRoles("g")
	// if err != nil {
	// 	return
	// }

	// Required Roles - admin and user
	// - The "GetAllNamedRoles" will not return an empty role
	if !functions.Contains(roles, "admin") {
		roles = append(roles, "admin")
	}
	if !functions.Contains(roles, "patcher") {
		roles = append(roles, "patcher")
	}

	data := gin.H{
		"roles": roles,
	}
	// c.JSON(http.StatusOK, data)
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "role-list.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, models.Roles{}.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// GetRole endpoint (GET)
// PathParams: id
func GetRole(c *gin.Context) {
	role, err := getRole(c)
	if err != nil {
		// Error has already been sent, just return
		return
	}
	data := gin.H{"roleName": role.Name, "users": role.Users}
	// c.JSON(http.StatusOK, data)
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "role-show.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, role.GetBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// UpdateRole endpoint (PUT/POST)
// - PathParams: name
func UpdateRole(c *gin.Context) {
	var err error
	var status string
	roleName := c.Param("name")
	e := middleware.GetEnforcer()
	currentUser, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "User hasn't logged in yet"})
		return
	}

	err = c.Request.ParseForm()
	if err != nil {
		log.Error("Parse Form Error: " + err.Error())
	}

	if user := c.PostForm("addUser"); user != "" { // Handle Add User
		user = strings.ToLower(user)
		if user == currentUser {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "Not allowed to add yourself!"})
			return
		}
		log.WithFields(log.Fields{
			"roleName": roleName,
			"target":   user,
			"user":     currentUser,
		}).Info("AUDIT: Add user to role.")
		ok, err = e.AddRoleForUser(user, roleName)
		if err != nil {
			log.WithFields(log.Fields{
				"roleName": roleName,
				"target":   user,
				"error":    err,
			}).Error("Error adding target user to role.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "Error adding user to role", "error": err.Error()})
			return
		}

	} else if user := c.PostForm("removeUser"); user != "" { // Handle Remove User
		if user == currentUser {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "Not allowed to remove yourself!"})
			return
		}
		log.WithFields(log.Fields{
			"roleName": roleName,
			"target":   user,
			"user":     currentUser,
		}).Info("AUDIT: Remove user from role.")
		ok, err = e.DeleteRoleForUser(user, roleName)
		if err != nil {
			log.WithFields(log.Fields{
				"roleName": roleName,
				"target":   user,
				"error":    err,
			}).Error("Error removing target user from role.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "Error removing user from role", "error": err.Error()})
			return
		}
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "Unknown operation!"})
		return
	}

	if ok {
		status = "success"
	} else {
		status = "failed"
	}
	users, err := e.GetUsersForRole(roleName)
	if err != nil {
		log.WithFields(log.Fields{
			"roleName": roleName,
			"error":    err,
		}).Error("Error getting users for role.")
	}
	data := gin.H{"status": status, "roleName": roleName, "users": users}
	c.Negotiate(http.StatusOK, gin.Negotiate{
		HTMLName: "role-success-redirect.gohtml",
		Data:     data,
		HTMLData: getHTMLData(c, models.GetDefaultBreadCrumbs(), data),
		Offered:  formatAllSupported,
	})
}

// ------------------------- STANDARD PATTERN HELPERS ---------------------------------

// getRole will get the id from context and return job
func getRole(c *gin.Context) (role *models.Role, err error) {
	// Retrieve "name" path parameter
	name := c.Param("name")
	return getRoleByName(c, name)
}

// getRoleByName retrives the patchRun from the DB
func getRoleByName(c *gin.Context, name string) (role *models.Role, err error) {
	// Get Role from enforcer (not from DB)
	role = models.NewRole()
	role.Name = name
	role.Users, err = middleware.GetEnforcer().GetUsersForRole(name)
	if err != nil {
		log.WithFields(log.Fields{
			"roleName": role.Name,
			"error":    err,
		}).Error("Error getting users for role.")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	return // success
}
