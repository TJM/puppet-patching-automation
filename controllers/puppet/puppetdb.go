package puppet

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/puppetlabs/go-pe-client/pkg/puppetdb"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
)

var pdbTimeout = 30 * time.Second

// PDBEnvironments will return puppet environments for the selected puppet server
func PDBEnvironments(p *models.PuppetServer) (envs []puppetdb.Environment, err error) {
	client, err := getPDBClient(p)
	if err != nil {
		return // already logged
	}

	envs, err = client.Environments()
	return
}

// GetFacts will return the facts for a host from PuppetDB
func GetFacts(s *models.Server) (result []puppetdb.Inventory, err error) {
	p := s.PuppetServer
	if p == nil {
		p, err = models.GetPuppetServerByID(s.PuppetServerID)
		if err != nil {
			return
		}
	}
	client, err := getPDBClient(p)
	if err != nil {
		return // already logged
	}
	query := fmt.Sprintf(`[ "=", "certname", "%s" ]`, s.Name) // Get facts for server.Name
	log.Debugf("Query PuppetServer: %s (%s) for facts on server %s...", p.Name, p.GetPuppetDBUrl(), s.Name)
	pagination := puppetdb.Pagination{
		IncludeTotal: true,
	}
	result, err = client.Inventory(query, &pagination, nil)
	if err != nil {
		log.Error("Error querying puppetDB: " + err.Error())
		return
	}
	factCount := 0
	if len(result) > 0 {
		factCount = len(result[0].Facts)
	}
	log.Debugf("Inventory Results: %v / Facts: %v", len(result), factCount)
	return
}

// GetPatchWindows will return available patch windows from PuppetDB
func GetPatchWindows(p *models.PuppetServer) (facts []puppetdb.Fact, err error) {
	client, err := getPDBClient(p)
	if err != nil {
		return // already logged
	}
	query := fmt.Sprintf(`[ "extract", [ "value", [ "function", "count" ] ], [ "=", "path", %s ], [ "group_by", "value" ] ]`, p.GetFactNamePath()) // Get patch_window values and counts
	log.Infof("Query PuppetServer: %s (%s) for patch windows...", p.Name, p.GetPuppetDBUrl())
	if p.SSLSkipVerify {
		log.Warnf("Skipping SSL Verification on %s", p.GetPuppetDBUrl())
	}
	pagination := puppetdb.Pagination{
		IncludeTotal: true,
	}
	orderBy := puppetdb.OrderBy{
		Field: "value",
		Order: "asc",
	}
	facts, err = client.FactContents(query, &pagination, &orderBy)
	if err != nil {
		log.Error("Error querying puppetDB: " + err.Error())
		return
	}
	log.Info("Inventory Results: ", len(facts))
	return
}

// GetInventoryForPatchRun will query all enabled Puppet Servers for the inventory results
func GetInventoryForPatchRun(patchRun *models.PatchRun) (errors []error) {
	// TODO: Make this operation asynchronous
	// First DELETE all existing applications for this patch run
	for _, app := range patchRun.GetApplications() {
		err := app.Delete(true)
		if err != nil {
			errors = append(errors, err)
			return
		}
	}
	// Now loop through each enabled puppet server
	for _, ps := range models.GetEnabledPuppetServers() {
		log.Infof("Query PuppetServer: %s (%s)", ps.Name, ps.GetPuppetDBUrl())
		if ps.SSLSkipVerify {
			log.Warnf("Skipping SSL Verification on %s", ps.GetPuppetDBUrl())
		}
		err := QueryPuppetDBInventory(ps, patchRun)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return
}

// QueryPuppetDBInventory will actually do the query and return the results
func QueryPuppetDBInventory(p *models.PuppetServer, patchRun *models.PatchRun) (err error) {
	client, err := getPDBClient(p)
	if err != nil {
		return // already logged
	}
	query := fmt.Sprintf(`["~","facts.%s","%v"]`, p.FactName, patchRun.PatchWindow)
	pagination := puppetdb.Pagination{
		IncludeTotal: true,
	}
	orderBy := puppetdb.OrderBy{
		Field: "certname",
		Order: "asc",
	}
	items, err := client.Inventory(query, &pagination, &orderBy)
	if err != nil {
		log.Error("Error querying puppetDB: " + err.Error())
		return
	}
	log.Info("Inventory Results: ", len(items))

	// Create Applications by parsing query output
	for _, server := range items {
		appName := getFactString(server, "application")
		envName := getFactString(server, "application_environment")
		componentName := getFactString(server, "application_component")
		url := getFactString(server, "patching-automation.patching_procedure_url")
		healthcheckScript := getFactString(server, "patching-automation.post_reboot_scriptpath")

		log.WithFields(log.Fields{
			"server":            server.Certname,
			"application":       appName,
			"environment":       envName,
			"component":         componentName,
			"URL":               url,
			"HealthCheckScript": healthcheckScript,
		}).Debug("Processing PuppetDB Inventory Result")

		// Add Server to application/environment/component
		app := models.GetOrCreateApplication(appName, patchRun.ID)
		if strings.HasPrefix(url, "http") {
			app.PatchingProcedure = url
		}
		app.Save()

		component := app.Environment(envName).Component(componentName)
		if component.HealthCheckScript == "" && healthcheckScript != "UNSET" {
			component.HealthCheckScript = healthcheckScript
			component.Save()
		} else if component.HealthCheckScript != healthcheckScript && healthcheckScript != "UNSET" {
			log.WithFields(log.Fields{
				"application": appName,
				"component":   componentName,
				"old":         component.HealthCheckScript,
				"new":         healthcheckScript,
				"server":      server.Certname,
			}).Warn("Attempted to change HealthCheck Script, not doing it! Look at Puppet Config for issues.")
		}
		dbServer := component.Server(server.Certname)
		parseServerResult(dbServer, server, p)
		dbServer.Save()

	}
	return nil
}

// parseServerResult Insert necessary fact values into server struct
// NOTE: This is the part that converts from puppetDB facts to Server Model
func parseServerResult(s *models.Server, Server puppetdb.Inventory, p *models.PuppetServer) {
	patchingFact := strings.Split(p.FactName, ".")
	s.PuppetServerID = p.ID
	s.IPAddress = getFactString(Server, "ipaddress")
	s.OperatingSystem = getFactString(Server, "os.name")
	s.OSVersion = getFactString(Server, "os.release.full")
	s.PackageUpdates = getFactInt(Server, patchingFact[0]+".package_update_count")
	s.PatchWindow = getFactString(Server, p.FactName)
	s.PinnedPackages = getFactArrayOfStrings(Server, patchingFact[0]+".pinned_packages")
	s.SecurityUpdates = getFactInt(Server, patchingFact[0]+".security_package_update_count")

	// Get cqjw-xxxxx name for legacy cliqa servers (example)
	if strings.Contains(Server.Certname, "cliqa") {
		s.VMName = getFactString(Server, "cliqr.cliqrNodeHostname")
	} else {
		s.VMName = getFactString(Server, "hostname")
	}
	s.UUID = getFactString(Server, "dmi.product.uuid")
}

// getPDBClient Create a Puppet Enterprise client
func getPDBClient(p *models.PuppetServer) (client *puppetdb.Client, err error) {
	if p.PDBClient == nil {
		// Validate Orchestrator URL
		var u *url.URL
		u, err = url.Parse(p.GetPuppetDBUrl())
		if err != nil {
			return
		}
		fmt.Println(u)

		client = puppetdb.NewClient(u.String(), p.Token, getTLSconfig(p.CACert, p.SSLSkipVerify), pdbTimeout)
		// pdb.Version() // TODO: Add some simple query to quickly "verify" that the connection is working
		p.PDBClient = client
	} else {
		client = p.PDBClient
	}
	return
}

func getFact(facts map[string]interface{}, factPath string) (fact interface{}, err error) {
	if len(factPath) == 0 {
		err = errors.New("empty factPath is invalid")
		return
	}
	factPathList := strings.SplitN(factPath, ".", 2) // "one.two.three" -> ["one", "two.three"] or "one" -> ["one"]
	factName := factPathList[0]
	if f, ok := facts[factName]; ok {
		if len(factPathList) == 2 && f != nil {
			fact, err = getFact(f.(map[string]interface{}), factPathList[1])
			return
		}
		fact = f
	} else {
		err = errors.New("fact not found")
	}
	return
}

// getFactString : Return a string from a fact, with default "UNSET"
func getFactString(server puppetdb.Inventory, factPath string) (result string) {
	fact, err := getFact(server.Facts, factPath)
	if err != nil {
		log.WithFields(log.Fields{
			"certname": server.Certname,
			"factPath": factPath,
		}).Info("Error Retrieving Fact: ", err)
		return "UNSET"
	}
	if fact != nil {
		return fact.(string)
	}
	return "NULL"
}

// getFactArrayOfStrings : Return an array of strings from a fact (using default values getFactString, which shouldn't occur)
func getFactArrayOfStrings(server puppetdb.Inventory, factPath string) (result []string) {
	fact, err := getFact(server.Facts, factPath)
	if err != nil {
		log.WithFields(log.Fields{
			"certname": server.Certname,
			"factPath": factPath,
		}).Info("Error Retrieving Fact: ", err)
		return
	}
	val := fact.([]interface{})
	for _, v := range val {
		result = append(result, v.(string))
	}
	return result
}

// getFactInt : Return an integer from a fact, default 0
func getFactInt(server puppetdb.Inventory, factPath string) (result int) {
	fact, err := getFact(server.Facts, factPath)
	if err != nil {
		log.WithFields(log.Fields{
			"certname": server.Certname,
			"factPath": factPath,
		}).Info("Error Retrieving Fact: ", err)
		return 0
	}

	if fact != nil {
		return int(fact.(float64))
	}
	return -1
}
