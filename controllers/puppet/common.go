package puppet

import (
	"crypto/tls"
	"crypto/x509"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func getTLSconfig(cacert string, SSLSkipVerify bool) (config *tls.Config) {
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Error("Error getting SystemCertPool:" + err.Error())
		// continue
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if cacert != "" {
		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM([]byte(cacert)); !ok {
			log.Error("No CA certs appended")
			// continue
		}
	}

	// Trust the augmented cert pool in our client

	/* #nosec G402 - SSLSkipVerify defaults to false, users are warned when switching it on. */
	config = &tls.Config{
		RootCAs:            rootCAs,
		InsecureSkipVerify: SSLSkipVerify,
	}
	return
}

// getInterfaceValue will return an interface{} with an appropriate type for the puppetType
// NOTE: This is *very* basic and will revert to just returning the string value by default.
func getInterfaceValue(puppetType string, val string) (retVal interface{}) {
	var err error
	if val == "" {
		return nil // Don't set a parameter that is "" (empty)
	}
	// Handle Optional[WHATEVER]
	if strings.HasPrefix(puppetType, "Optional[") {
		puppetType = strings.TrimPrefix(puppetType, "Optional[")
		puppetType = strings.TrimSuffix(puppetType, "]")
		return getInterfaceValue(puppetType, val)
	}
	if puppetType == "Integer" {
		retVal, err = strconv.Atoi(val)
		if err != nil {
			log.WithField("val", val).Error("Error converting to integer: ", err)
			return nil
		}
		return
	}
	if puppetType == "Boolean" {
		switch val {
		case "yes", "true", "1":
			return true
		default:
			return false
		}
	}
	// Otherwise just return the "string" (shrug)
	return val
}
