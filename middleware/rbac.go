package middleware

import (
	"github.com/casbin/casbin/v2"
	log "github.com/sirupsen/logrus"
)

// // ReloadPoliciesFromDB will reload the RBAC policies from the database in the event
// //   that they have been modified during runtime.
// func ReloadPoliciesFromDB() (err error) {
// 	err = GetEnforcer().LoadPolicy()
// 	return
// }

// AddUserToRole will assign a user to a role (userGroup) if they are not already assigned
func AddUserToRole(user, role string) (result bool, err error) {
	return addUserToRole(GetEnforcer(), user, role)
}

func addUserToRole(e *casbin.Enforcer, user, role string) (result bool, err error) {
	if !e.HasGroupingPolicy(user, role) {
		log.WithFields(log.Fields{
			"user": user,
			"role": role,
		}).Info("Add user to role.")
		result, err = e.AddGroupingPolicy(user, role)
		if err != nil {
			log.WithFields(log.Fields{
				"user": user,
				"role": role,
			}).Error("Error adding user to role: " + err.Error())
		}
	} else {
		log.WithFields(log.Fields{
			"user": user,
			"role": role,
		}).Debug("User was already assigned the role.")
	}
	return
}

// AddResourceToGroup will add a resource object (path target) to a resource group if it does not exist already.
func AddResourceToGroup(resource, group string) (result bool, err error) {
	return addResourceToGroup(GetEnforcer(), resource, group)
}

func addResourceToGroup(e *casbin.Enforcer, resource, group string) (result bool, err error) {
	if !e.HasNamedGroupingPolicy("g2", resource, group) {
		log.WithFields(log.Fields{
			"resource": resource,
			"group":    group,
		}).Infof("Add resource to group")
		result, err = e.AddNamedGroupingPolicy("g2", resource, group)
		if err != nil {
			log.WithFields(log.Fields{
				"resource": resource,
				"group":    group,
			}).Error("Error adding resource to group: " + err.Error())
		}
	} else {
		log.WithFields(log.Fields{
			"resource": resource,
			"group":    group,
		}).Debug("Reource already found in group.")
	}
	return
}

// AddPolicy will add a policy to the enforcer if it does not exist.
func AddPolicy(subject, object, action string) (result bool, err error) {
	return addPolicy(GetEnforcer(), subject, object, action)
}

func addPolicy(e *casbin.Enforcer, subject, object, action string) (result bool, err error) {
	if !e.HasPolicy(subject, object, action) {
		log.WithFields(log.Fields{
			"subject": subject,
			"object":  object,
			"action":  action,
		}).Info("Add policy")
		result, err = e.AddPolicy(subject, object, action)
		if err != nil {
			log.WithFields(log.Fields{
				"subject": subject,
				"object":  object,
				"action":  action,
			}).Error("Error adding policy: " + err.Error())
		}
	} else {
		log.WithFields(log.Fields{
			"subject": subject,
			"object":  object,
			"action":  action,
		}).Debug("Policy was found")
	}
	return
}
