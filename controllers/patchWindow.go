package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/puppetlabs/go-pe-client/pkg/puppetdb"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/controllers/puppet"
	"github.com/tjm/puppet-patching-automation/models"
)

// GetPatchWindows endpoint (GET)
func GetPatchWindows(c *gin.Context) {
	patchWindows := make(map[string]puppetdb.Fact)
	// var errors []error

	for _, ps := range models.GetEnabledPuppetServers() {
		facts, err := puppet.GetPatchWindows(ps)
		if err != nil {
			log.Error("Error Querying Puppet Facts: ", err)
			// errors = append(errors, err)
		}

		// Collect results from multiple servers combining "counts"
		for _, fact := range facts {
			val := fact.Value.(string)
			if val != "" { // Exclude empty patch window values
				if _, ok := patchWindows[val]; ok {
					fact.Count += patchWindows[val].Count
					patchWindows[val] = fact
				} else {
					patchWindows[val] = fact
				}
			}
		}
	}

	list := make([]puppetdb.Fact, 0)
	for _, fact := range patchWindows {
		list = append(list, fact)
	}

	//data := gin.H{"status": "success", "patch_windows": list}
	// TODO: return errors list if they are there and maybe think about a partial success return code?
	c.JSON(http.StatusOK, list)
}
