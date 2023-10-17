package views

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/tjm/puppet-patching-automation/models"
)

// OutputAppsEnvs produces a list of application [env, env2]
func OutputAppsEnvs(apps models.Applications) (out string) {
	out = "*** Applications [environments] ***\n"
	for _, app := range apps {
		out += fmt.Sprintf("%v %v\n", app.Name, app.GetEnvironmentList())
	}
	return
}

// OutputServerCSV outputs a CSV of the server details
func OutputServerCSV(apps models.Applications) (out *bytes.Buffer, err error) {
	out = new(bytes.Buffer)
	writer := csv.NewWriter(out)

	_ = writer.Write([]string{
		"ServerName",
		"IP",
		"Application",
		"AppEnvironment",
		"AppComponent",
		"OS",
		"OSVersion",
		"Updates",
		"SecurityUpdates",
		"PatchWindow",
		"VMName",
	})
	for _, app := range apps {
		for _, env := range app.GetEnvironments() {
			for _, component := range env.GetComponents() {
				for _, server := range component.GetServers() {
					err = writer.Write([]string{
						server.Name,
						server.IPAddress,
						app.Name,
						env.Name,
						component.Name,
						server.OperatingSystem,
						server.OSVersion,
						fmt.Sprint(server.PackageUpdates),
						fmt.Sprint(server.SecurityUpdates),
						server.PatchWindow,
						server.VMName,
					})
					if err != nil {
						return
					}
				}
			}
		}
	}
	writer.Flush()
	err = writer.Error()
	return
}

// OutputServerList outputs a list of servers in apps
// ... used for the Jenkins Job to silence monitoring and possibly SNOW
func OutputServerList(apps models.Applications) (out string) {
	for _, app := range apps {
		for _, env := range app.GetEnvironments() {
			for _, component := range env.GetComponents() {
				out += strings.Join(component.GetServerList(), "\n")
				out += "\n" // Add a trailing newline
			}
		}
	}
	out += "\n" // Add a trailing newline
	return
}

// OutputAnsibleInventory will output the inventory file for ansible - saving for future use
// func OutputAnsibleInventory(app string, env string, patchRunID uint) string {
// 	var out string
// 	if application := models.GetApplication(app, patchRunID); application != nil {
// 		if environment := application.GetEnvironment(env); environment != nil {
// 			out += "## ANSIBLE INVENTORY ##\n"
// 			out += fmt.Sprintf("# Application: %v (%v)\n\n", app, env)
// 			for name, comp := range environment.Components {
// 				out += fmt.Sprintf("[%v]\n", name)
// 				for serverName := range comp.Servers {
// 					out += fmt.Sprintf("%v\n", serverName)
// 				}
// 				out += "\n"
// 			}
// 			return out
// 		}
// 		log.WithFields(log.Fields{
// 			"application": app,
// 			"environment": env,
// 		}).Error("Environment does not exist.")
// 	} else {
// 		log.WithFields(log.Fields{
// 			"application": app,
// 			"environment": env,
// 		}).Error("Application does not exist.")
// 	}
// 	return ""

// }
