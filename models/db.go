package models

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/tjm/puppet-patching-automation/config"
)

var db *gorm.DB

// TestApp is for testing
type TestApp struct {
	gorm.Model
	Name              string `json:"name"`
	PatchingProcedure string `json:"patching_procedure"`
}

// Connect to database and run migrations
func Connect() {
	var err error
	args := config.GetArgs()
	dbType := args.DBType
	host := args.DBHost
	name := args.DBName
	user := args.DBUser
	password := args.DBPassword
	port := args.DBPort

	switch dbType {
	case "sqlite3", "sqlite":
		dsn := fmt.Sprintf("db/%s.db", name)
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})

	case "postgresql", "postgres":
		if port == 0 {
			port = 5432
		}
		// https://github.com/jackc/pgx
		dsn := fmt.Sprintf("host=%s user=%s password=%s DB.name=%s port=%v sslmode=disable", host, user, password, name, port)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	case "mysql":
		if port == 0 {
			port = 3306
		}
		// https://github.com/go-sql-driver/mysql
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, name)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	default:
		panic("Unsupported database type: " + dbType)
	}

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	if args.DebugDB {
		db = db.Debug()
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&Application{},
		&Environment{},
		&Component{},
		&Server{},
		&TrelloBoard{},
		&PatchRun{},
		&PuppetServer{},
		&PuppetTask{},
		&PuppetTaskParam{},
		&PuppetPlan{},
		&PuppetPlanParam{},
		&PuppetJob{},
		&JenkinsServer{},
		&JenkinsJob{},
		&JenkinsJobParam{},
		&JenkinsBuild{},
		ChatRoom{},
	)
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}
}

// GetDB returns a handle to the DB object
func GetDB() *gorm.DB {
	return db
}
