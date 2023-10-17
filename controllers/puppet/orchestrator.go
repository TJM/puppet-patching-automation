package puppet

import (
	"net/url"

	"github.com/puppetlabs/go-pe-client/pkg/orch"

	"github.com/tjm/puppet-patching-automation/models"
)

// getOrchClient Create a Puppet Orchestrator client
func getOrchClient(p *models.PuppetServer) (client *orch.Client, err error) {
	if p.OrchClient == nil {
		// Validate Orchestrator URL
		var u *url.URL
		u, err = url.Parse(p.GetOrchURL())
		if err != nil {
			return
		}

		client = orch.NewClient(u.String(), p.Token, getTLSconfig(p.CACert, p.SSLSkipVerify))
		// pdb.Version() // TODO: Add some simple query to quickly "verify" that the connection is working
		p.OrchClient = client
	} else {
		client = p.OrchClient
	}
	return
}
