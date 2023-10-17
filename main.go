package main

import (
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/config"
	"github.com/tjm/puppet-patching-automation/middleware"
	"github.com/tjm/puppet-patching-automation/models"
	"github.com/tjm/puppet-patching-automation/routes"
	"github.com/tjm/puppet-patching-automation/version"
)

func main() {
	args := config.GetArgs()
	log.Infof("[INFO] PatchingAutomation version: %s [%s %s %s]",
		version.FormattedVersion(),
		runtime.Version(),
		runtime.GOOS, runtime.GOARCH)

	if args.Debug {
		log.SetLevel(log.TraceLevel)
		log.SetReportCaller(true)
		log.Debugf("DEBUG LOGS ENABLED")
	}

	models.Connect()
	middleware.Init()

	routes.StartService()
}
