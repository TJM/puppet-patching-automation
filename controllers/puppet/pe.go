package puppet

import (
	"fmt"
	"net/url"

	"github.com/puppetlabs/go-pe-client/pkg/pe"

	"github.com/tjm/puppet-patching-automation/models"
)

// PEEnvironments will return puppet environments for the selected puppet server
func PEEnvironments(p *models.PuppetServer) (environments []string, err error) {
	client, err := getPEClient(p)
	if err != nil {
		return // already logged
	}
	environments, err = client.Environments()
	return
}

// getPEClient Create a Puppet Enterprise client
func getPEClient(p *models.PuppetServer) (client *pe.Client, err error) {
	if p.PEClient == nil {
		// Validate Orchestrator URL
		var u *url.URL
		u, err = url.Parse(p.GetPEURL())
		if err != nil {
			return
		}
		fmt.Println(u)

		client = pe.NewClient(u.String(), p.Token, getTLSconfig(p.CACert, p.SSLSkipVerify))
		// pdb.Version() // TODO: Add some simple query to quickly "verify" that the connection is working
		p.PEClient = client
	} else {
		client = p.PEClient
	}
	return
}
