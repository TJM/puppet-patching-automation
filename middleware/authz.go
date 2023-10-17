package middleware

import (
	"fmt"
	"strings"

	casbin "github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/config"
	"github.com/tjm/puppet-patching-automation/models"
)

const casbinConfig = "config/rbac_model.conf"
const casbinPolicy = "config/authz_policy.csv"

var globalEnforcer *casbin.Enforcer

// Authorize determines if current user has been authorized to take an action on an object.
func Authorize(obj string, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Initialized Enforcer
		enforcer := GetEnforcer()

		// Get current user/subject
		sub, existed := c.Get("user")
		if !existed {
			c.AbortWithStatusJSON(401, gin.H{"msg": "User hasn't logged in yet"})
			return
		}
		user := strings.ToLower(fmt.Sprint(sub))

		// Casbin enforces policy
		ok, pols, err := enforcer.EnforceEx(user, obj, act)

		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"msg": "Error occurred when authorizing user"})
			return
		}

		if !ok {
			log.WithFields(log.Fields{
				"user":     user,
				"object":   obj,
				"action":   act,
				"policies": pols,
			}).Info("Access Denied.")
			c.AbortWithStatusJSON(403, gin.H{"msg": "You are not authorized"})
			return
		}
		log.WithFields(log.Fields{
			"user":     user,
			"object":   obj,
			"action":   act,
			"policies": pols,
		}).Info("Access Granted.")
		c.Next()
	}
}

// HasAccess will return true or false if the subject has access
func HasAccess(sub, obj, act string) bool {
	// Get Initialized Enforcer
	enforcer := GetEnforcer()

	// Casbin enforces policy
	ok, err := enforcer.Enforce(fmt.Sprint(sub), obj, act)
	if err != nil {
		return false
	}
	return ok
}

// createEnforcer will create and initialize the "enforcer" for CASBIN
func createEnforcer() (e *casbin.Enforcer) {
	var err error
	var adapter *gormadapter.Adapter

	log.Info("Creating RBAC enforcer...")

	// Initialize casbin GORM (database) adapter
	adapter, err = gormadapter.NewAdapterByDB(models.GetDB())
	if err != nil {
		log.Fatal("failed to initialize casbin gorm adapter: " + err.Error())
	}
	// CASBIN Enforcer from database adapter
	e, err = casbin.NewEnforcer(casbinConfig, adapter)
	if err != nil {
		log.Fatal("Error with NewEnforcer: " + err.Error())
	}
	return
}

// initEnforcer will initialize the enforcer with default policies, mostly for first run
// or if an upgrade provides new features or policies. It will also add initial admins.
func initEnforcer(e *casbin.Enforcer) {
	log.Info("Initializing RBAC enforcer...")

	args := config.GetArgs()

	for _, user := range args.InitAdmins {
		_, err := addUserToRole(e, user, "admin")
		if err != nil {
			log.Error("Error addUserToRole: " + err.Error())
		}
	}

	for _, user := range args.InitUsers {
		_, err := addUserToRole(e, user, "patcher")
		if err != nil {
			log.Error("Error addUserToRole: " + err.Error())
		}
	}

	// Read Policy from CSV
	csvEnforcer, err := casbin.NewEnforcer(casbinConfig, casbinPolicy)
	if err != nil {
		log.Fatal("Failed to load initial policies from CSV")
	}

	// Process "p" objects (policies)
	for _, p := range csvEnforcer.GetPolicy() {
		_, err := addPolicy(e, p[0], p[1], p[2])
		if err != nil {
			log.Error("Error addPolicy: " + err.Error())
		}
	}

	// Process "g" objects (groups/roles)
	for _, g := range csvEnforcer.GetGroupingPolicy() {
		_, err := addResourceToGroup(e, g[0], g[1])
		if err != nil {
			log.Error("Error addResourceToGroup: " + err.Error())
		}
	}

	// Process "g2" objects (resource groups)
	for _, g2 := range csvEnforcer.GetNamedGroupingPolicy("g2") {
		_, err := addResourceToGroup(e, g2[0], g2[1])
		if err != nil {
			log.Error("Error addResourceToGroup: " + err.Error())
		}
	}

	if args.DebugAuth {
		fmt.Printf("\n\n*** initEnforcer Results\n")
		fmt.Printf("*** - Roles (User Groups): %+v\n", e.GetGroupingPolicy())
		fmt.Printf("*** - Resource Groups: %+v\n", e.GetNamedGroupingPolicy("g2"))
		fmt.Printf("*** - Policies: %+v\n", e.GetPolicy())
		fmt.Printf("*** END initEnforcer Results\n\n")
		log.Info("initEnforer: Done!")
	}
}

// GetEnforcer returns the current enforcer and initializes if needed
func GetEnforcer() *casbin.Enforcer {
	// Uses global variable `globalEnforcer`
	if globalEnforcer == nil {
		globalEnforcer = createEnforcer()
	}
	return globalEnforcer
}
