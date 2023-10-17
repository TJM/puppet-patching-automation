package config

import (
	"errors"

	"github.com/alexflint/go-arg"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/version"
)

// Arguments Type
type Arguments struct {
	Debug          bool     `arg:"env:DEBUG" help:"enable application debug mode (more logs)"`
	DebugDB        bool     `arg:"env:DEBUG_DB" help:"enable database debug mode (show all queries)"`
	DebugAuth      bool     `arg:"env:DEBUG_AUTH" help:"enable authentication/authorization debug mode"`
	TrelloAppKey   string   `default:"5a453a8d5b4ab0ae9a5746b34cc0b09e" arg:"env:TRELLO_APP_KEY" help:"Trello App Key (Identifies this App)"`
	TrelloToken    string   `arg:"env:TRELLO_TOKEN" help:"Trello Access Token: https://trello.com/1/connect?key=5a453a8d5b4ab0ae9a5746b34cc0b09e&name=PuppetPatchingAutomation&response_type=token&scope=read,write&expiration=1day"`
	DBType         string   `default:"sqlite3" arg:"env:DB_TYPE" help:"Database Type (eg. sqlite3, postgresql, mysql) (env: DB_TYPE)"`
	DBHost         string   `default:"localhost" arg:"env:DB_HOST" help:"Database Host (env: DB_HOST)"`
	DBPort         int      `default:"0" arg:"env:DB_PORT" help:"Database Port (env: DB_PORT) Default based on dbtype."`
	DBUser         string   `default:"padb" arg:"env:DB_USER" help:"Database Username (env: DB_USER)"`
	DBPassword     string   `default:"padb" arg:"env:DB_PASSWORD" help:"Database Password (env: DB_PASSWORD)"`
	DBName         string   `default:"padb" arg:"env:DB_NAME" help:"Database Name (env: DB_NAME)"`
	SessionName    string   `default:"PatchingAutomation" arg:"env:SESSION_NAME" help:"Session Name (cookie name) (env: SESSION_NAME)"`
	SessionAuthKey string   `default:"PatchingAutomationDefaultAuthKey" arg:"env:SESSION_AUTH_KEY" help:"Session Authentication Key, should be 32 or 64 bytes (env: SESSION_AUTH_KEY)"`
	SessionEncKey  string   `default:"PatchingAutomationDefaultEncrKey" arg:"env:SESSION_ENC_KEY" help:"Session Encrpytion Key, must be 16, 24 or 32 bytes (env: SESSION_ENC_KEY)"`
	TrustedProxies []string `arg:"env:TRUSTED_PROXIES" help:"Trusted Proxies - to provide Client Remote IP (comma separated) (env: TRUSTED_PROXIES)"`
	InitAdmins     []string `arg:"env:INIT_ADMINS" help:"Initial Admins - to provide initial administrative users. (comma separated) (env: INIT_ADMINS)"`
	InitUsers      []string `arg:"env:INIT_USERS" help:"Initial Users - to provide initial authorized users (aka patchers). (comma separated) (env: INIT_USERS)"`
	LogAudit       bool     `arg:"env:LOG_AUDIT" help:"Audit Log Authorization Messages (env: LOG_AUDIT)"`
}

var args *Arguments

func parseArgs() {
	// Load environment from .env
	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	arg.MustParse(args)

	// Validate
	err = args.Validate()
	if err != nil {
		log.Fatal("Error Validating Args: " + err.Error())
	}
}

// Validate arguments
func (a *Arguments) Validate() (err error) {
	var l int

	if a.SessionAuthKey == "PatchingAutomationDefaultAuthKey" {
		log.Warn("Using default value for --sessionauthkey (SESSION_AUTH_KEY). This should be set in production!")
	}
	l = len(a.SessionAuthKey)
	if !(l == 16 || l == 32) {
		log.Warn("Session Authentication Key should be 16 or 32 bytes.")
	}

	if a.SessionEncKey == "PatchingAutomationDefaultEncrKey" {
		log.Warn("Using default value for --sessionenckey (SESSION_ENC_KEY). This should be set in production!")
	}
	l = len(a.SessionEncKey)
	if !(l == 16 || l == 24 || l == 32) {
		err = errors.New("SessionEncKey must be 16, 24, or 32 bytes")
	}
	// if len(a.TrustedProxies) > 0 {
	// TODO: Implement validation, TrustedProxies should "look like" IPv4 or IPv6 address or CIDR
	// SetTrustedProxies set a list of network origins (IPv4 addresses, IPv4 CIDRs, IPv6 addresses or IPv6 CIDRs)
	// }
	return
}

// Version string
func (a *Arguments) Version() string {
	return version.FormattedVersion()
}

// GetArgs returns Args
func GetArgs() *Arguments {
	if args == nil {
		args = new(Arguments)
		parseArgs()
	}
	return args
}
