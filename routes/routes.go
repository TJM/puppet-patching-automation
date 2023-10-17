package routes

import (
	"html/template"

	method "github.com/bu/gin-method-override"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/thinkerou/favicon"

	"github.com/tjm/puppet-patching-automation/controllers"
	"github.com/tjm/puppet-patching-automation/controllers/jenkinsapi"
	"github.com/tjm/puppet-patching-automation/functions"
	"github.com/tjm/puppet-patching-automation/middleware"
	"github.com/tjm/puppet-patching-automation/models"
)

// StartService function
func StartService() {
	router := gin.New()

	// Handle methodOverride (DELETE/PUT using _method hidden form field)
	router.Use(method.ProcessMethodOverride(router))

	// The default handlers
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Load Functions for templates and load templates
	router.SetFuncMap(template.FuncMap{
		"FormatAsDateTimeLocal":  functions.FormatAsDateTimeLocal,
		"FormatAsISO8601":        functions.FormatAsISO8601,
		"jenkinsClassIsFolder":   jenkinsapi.IsFolder,
		"isPuppetServersEnabled": models.IsPuppetServersEnabled,
		"hasAccess":              middleware.HasAccess,
	})
	router.LoadHTMLGlob("templates/**/*")

	// Generic Handlers
	middleware.SetupTrustedProxies(router)
	router.Use(middleware.HandleCORS())
	router.Use(middleware.HandleLocation())
	router.Use(middleware.HandleSession()) // Sessions must be Use(d) before oidcauth, as oidcauth requires sessions

	// Standard(ish) path handlers (unauthenticated)
	router.GET("/", controllers.GetHome)
	router.Static("/assets", "./assets")
	router.GET("/favicon.ico", favicon.New("assets/favicon.ico"))
	middleware.SetupAuthentication(router) // Sets up authentication endpoints /login, /logout and /AUTHREDIRECTPATH
	router.GET("/healthz", controllers.GetPing)
	router.GET("/ping", controllers.GetPing) // Legacy health check (almost every gin-gonic app has a /ping)

	// Patching Automation Specific Paths and Handlers
	patchRun := router.Group("/patchRun", middleware.Authenticate())
	{
		patchRun.GET("", middleware.Authorize("patchRun", "read"), controllers.GetPatchRunList)

		patchRun.GET(":id", middleware.Authorize("patchRun", "read"), controllers.GetPatchRun)
		patchRun.PUT(":id", middleware.Authorize("patchRun", "write"), controllers.UpdatePatchRun)
		patchRun.POST(":id", middleware.Authorize("patchRun", "write"), controllers.UpdatePatchRun)
		patchRun.DELETE(":id", middleware.Authorize("patchRun", "delete"), controllers.DeletePatchRun)

		patchRun.POST(":id/createTrelloBoard", middleware.Authorize("trelloBoard", "write"), controllers.CreateTrelloBoard)
		patchRun.GET(":id/trelloBoards", middleware.Authorize("patchRun", "read"), controllers.GetTrelloBoards)

		patchRun.POST(":id/runQuery", middleware.Authorize("patchRun", "write"), controllers.RunPuppetDBQuery)

		patchRun.POST(":id/linkChatRoom", middleware.Authorize("patchRun", "write"), controllers.LinkChatRoomToPatchRun)

		patchRun.GET(":id/buildJenkinsJob/:jobID", middleware.Authorize("jenkinsJobRun", "read"), controllers.BuildJenkinsJob) // PREVIEW
		patchRun.POST(":id/buildJenkinsJob/:jobID", middleware.Authorize("jenkinsJobRun", "run"), controllers.BuildJenkinsJob)

		patchRun.GET(":id/applications", middleware.Authorize("application", "read"), controllers.GetAllApplications)
		patchRun.GET(":id/serverList", middleware.Authorize("server", "read"), controllers.GetServerList)
		patchRun.GET(":id/serverCSV", middleware.Authorize("server", "read"), controllers.GetDetailedServerList)
		patchRun.GET(":id/appsEnvs", middleware.Authorize("application", "read"), controllers.GetAppsEnvs)
	}

	application := router.Group("/application", middleware.Authenticate())
	{
		// Get application IDs from /patchRun/:id/applications
		application.GET(":id", middleware.Authorize("application", "read"), controllers.GetApplication)
		application.GET(":id/environments", middleware.Authorize("environment", "read"), controllers.GetAllEnvironments)
	}

	environment := router.Group("/environment", middleware.Authenticate())
	{
		// Get environment IDs from /application/:id/environments
		environment.GET(":id", middleware.Authorize("environment", "read"), controllers.GetEnvironment)
		environment.GET(":id/components", middleware.Authorize("component", "read"), controllers.GetAllComponents)
	}

	component := router.Group("/component", middleware.Authenticate())
	{
		// Get component IDs from /environment/:id/components
		component.GET(":id", middleware.Authorize("component", "read"), controllers.GetComponent)
		component.GET(":id/servers", middleware.Authorize("component", "read"), controllers.GetAllServers)
		component.POST(":id/runPatching", middleware.Authorize("puppetTaskRun", "run"), controllers.ComponentRunPatching)
		component.GET(":id/runPuppetPlan/:puppetServerID/:planID", middleware.Authorize("puppetTaskRun", "run"), controllers.ComponentRunPuppetPlan)
		component.POST(":id/runPuppetPlan/:puppetServerID/:planID", middleware.Authorize("puppetPlanRun", "run"), controllers.ComponentRunPuppetPlan)
		component.GET(":id/runPuppetTask/:puppetServerID/:taskID", middleware.Authorize("puppetTaskRun", "run"), controllers.ComponentRunPuppetTask)
		component.POST(":id/runPuppetTask/:puppetServerID/:taskID", middleware.Authorize("puppetTaskRun", "run"), controllers.ComponentRunPuppetTask)
	}

	server := router.Group("/server", middleware.Authenticate())
	{
		// Get Server IDs from /component/:id/servers
		server.GET(":id", middleware.Authorize("server", "read"), controllers.GetServer)
		server.POST(":id/runPatching", middleware.Authorize("puppetTaskRun", "run"), controllers.ServerRunPatching)
		server.GET(":id/facts", middleware.Authorize("server", "read"), controllers.GetServerFacts)
	}

	trelloboard := router.Group("/trelloboard", middleware.Authenticate())
	{
		// Get TrelloBoard IDs from /patchRun/:id/trelloBoards
		trelloboard.GET(":id", middleware.Authorize("trelloBoard", "read"), controllers.GetTrelloBoard)
		trelloboard.DELETE(":id", middleware.Authorize("trelloBoard", "delete"), controllers.DeleteTrelloBoard)
	}

	data := router.Group("/data", middleware.Authenticate())
	{
		data.GET("patchWindows", middleware.Authorize("patchRun", "read"), controllers.GetPatchWindows)
	}

	config := router.Group("/config", middleware.Authenticate())
	{

		role := config.Group("/role")
		{
			role.GET("", middleware.Authorize("role", "read"), controllers.ListRoles)
			role.GET(":name", middleware.Authorize("role", "read"), controllers.GetRole)
			role.POST(":name", middleware.Authorize("role", "write"), controllers.UpdateRole)
		}

		ChatRoom := config.Group("/ChatRoom")
		{
			ChatRoom.GET("", middleware.Authorize("chatRoom", "read"), controllers.ListChatRooms)
			ChatRoom.GET(":id", middleware.Authorize("chatRoom", "read"), controllers.GetChatRoom)
			ChatRoom.PUT(":id", middleware.Authorize("chatRoom", "write"), controllers.UpdateChatRoom)
			ChatRoom.POST(":id", middleware.Authorize("chatRoom", "write"), controllers.UpdateChatRoom)
			ChatRoom.DELETE(":id", middleware.Authorize("chatRoom", "delete"), controllers.DeleteChatRoom)
			ChatRoom.GET(":id/test", middleware.Authorize("chatRoom", "read"), controllers.TestChatRoom)
		}

		puppetServer := config.Group("/puppetServer")
		{
			puppetServer.GET("", middleware.Authorize("puppetServer", "read"), controllers.ListPuppetServers)

			puppetServer.GET(":id", middleware.Authorize("puppetServer", "read"), controllers.GetPuppetServer)
			puppetServer.PUT(":id", middleware.Authorize("puppetServer", "write"), controllers.UpdatePuppetServer)
			puppetServer.POST(":id", middleware.Authorize("puppetServer", "write"), controllers.UpdatePuppetServer)
			puppetServer.DELETE(":id", middleware.Authorize("puppetServer", "delete"), controllers.DeletePuppetServer)

			puppetServer.GET(":id/environments-pe", middleware.Authorize("puppetServer", "read"), controllers.GetPuppetServerEnvironmentsPE)
			puppetServer.GET(":id/environments", middleware.Authorize("puppetServer", "read"), controllers.GetPuppetServerEnvironments)

			puppetServer.GET(":id/apiTasks/:env", middleware.Authorize("puppetServer", "read"), controllers.GetPuppetServerAPITasks)
			puppetServer.GET(":id/apiTask/:env/:module/:taskName", middleware.Authorize("puppetServer", "read"), controllers.GetPuppetServerAPITask)

			puppetServer.GET(":id/apiPlans/:env", middleware.Authorize("puppetServer", "read"), controllers.GetPuppetServerAPIPlans)
			puppetServer.GET(":id/apiPlan/:env/:module/:planName", middleware.Authorize("puppetServer", "read"), controllers.GetPuppetServerAPIPlan)

			puppetServer.GET(":id/tasks", middleware.Authorize("puppetTask", "read"), controllers.GetPuppetServerTasks)
			puppetServer.POST(":id/addTask", middleware.Authorize("puppetTask", "write"), controllers.AddPuppetTaskToServer)
			puppetServer.POST(":id/associateTask", middleware.Authorize("puppetTask", "write"), controllers.AssociatePuppetTaskToServer)
			puppetServer.POST(":id/disassociateTask", middleware.Authorize("puppetTask", "write"), controllers.DisassociatePuppetTaskFromServer)

			puppetServer.GET(":id/plans", middleware.Authorize("puppetPlan", "read"), controllers.GetPuppetServerPlans)
			puppetServer.POST(":id/addPlan", middleware.Authorize("puppetPlan", "write"), controllers.AddPuppetPlanToServer)
			puppetServer.POST(":id/associatePlan", middleware.Authorize("puppetPlan", "write"), controllers.AssociatePuppetPlanToServer)
			puppetServer.POST(":id/disassociatePlan", middleware.Authorize("puppetPlan", "write"), controllers.DisassociatePuppetPlanFromServer)

			puppetServer.GET(":id/job/:jobID", middleware.Authorize("puppetServer", "read"), controllers.GetPuppetServerJob)
			puppetServer.GET(":id/jobReport/:jobID", middleware.Authorize("puppetServer", "read"), controllers.GetPuppetServerJobReport)
		}

		puppetTask := config.Group("/puppetTask")
		{
			puppetTask.GET("", middleware.Authorize("puppetTask", "read"), controllers.ListPuppetTasks)

			puppetTask.GET(":id", middleware.Authorize("puppetTask", "read"), controllers.GetPuppetTask)
			puppetTask.PUT(":id", middleware.Authorize("puppetTask", "write"), controllers.UpdatePuppetTask)
			puppetTask.POST(":id", middleware.Authorize("puppetTask", "write"), controllers.UpdatePuppetTask)
			puppetTask.POST(":id/updateFromAPI", middleware.Authorize("puppetTask", "write"), controllers.UpdatePuppetTaskFromAPI)
			puppetTask.DELETE(":id", middleware.Authorize("puppetTask", "delete"), controllers.DeletePuppetTask)
		}

		puppetTaskParam := config.Group("/puppetTaskParam")
		{
			// puppetTaskParam.GET("", middleware.Authorize("puppetTask", "read"), controllers.ListPuppetTaskParams)
			// puppetTaskParam.GET(":id", middleware.Authorize("puppetTask", "read"), controllers.GetPuppetTaskParam)
			puppetTaskParam.PUT(":id", middleware.Authorize("puppetTask", "write"), controllers.UpdatePuppetTaskParam)
			puppetTaskParam.POST(":id", middleware.Authorize("puppetTask", "write"), controllers.UpdatePuppetTaskParam)
			puppetTaskParam.DELETE(":id", middleware.Authorize("puppetTask", "delete"), controllers.DeletePuppetTaskParam)
		}

		puppetPlan := config.Group("/puppetPlan")
		{
			puppetPlan.GET("", middleware.Authorize("puppetPlan", "read"), controllers.ListPuppetPlans)
			puppetPlan.GET(":id", middleware.Authorize("puppetPlan", "read"), controllers.GetPuppetPlan)
			puppetPlan.PUT(":id", middleware.Authorize("puppetPlan", "write"), controllers.UpdatePuppetPlan)
			puppetPlan.POST(":id", middleware.Authorize("puppetPlan", "write"), controllers.UpdatePuppetPlan)
			puppetPlan.POST(":id/updateFromAPI", middleware.Authorize("puppetPlan", "write"), controllers.UpdatePuppetPlanFromAPI)
			puppetPlan.DELETE(":id", middleware.Authorize("puppetPlan", "delete"), controllers.DeletePuppetPlan)
		}

		puppetPlanParam := config.Group("/puppetPlanParam")
		{
			// puppetPlanParam.GET("", controllers.ListPuppetPlanParams)
			// puppetPlanParam.GET(":id", middleware.Authorize("puppetPlan", "read"), controllers.GetPuppetPlanParam)
			puppetPlanParam.PUT(":id", middleware.Authorize("puppetPlan", "write"), controllers.UpdatePuppetPlanParam)
			puppetPlanParam.POST(":id", middleware.Authorize("puppetPlan", "write"), controllers.UpdatePuppetPlanParam)
			puppetPlanParam.DELETE(":id", middleware.Authorize("puppetPlan", "delete"), controllers.DeletePuppetPlanParam)
		}

		jenkinsServer := config.Group("/jenkinsServer")
		{
			jenkinsServer.GET("", middleware.Authorize("jenkinsServer", "read"), controllers.ListJenkinsServers)
			jenkinsServer.GET(":id", middleware.Authorize("jenkinsServer", "read"), controllers.GetJenkinsServer)
			jenkinsServer.PUT(":id", middleware.Authorize("jenkinsServer", "write"), controllers.UpdateJenkinsServer)
			jenkinsServer.POST(":id", middleware.Authorize("jenkinsServer", "write"), controllers.UpdateJenkinsServer)
			jenkinsServer.DELETE(":id", middleware.Authorize("jenkinsServer", "delete"), controllers.DeleteJenkinsServer)

			jenkinsServer.GET(":id/info", middleware.Authorize("jenkinsServer", "read"), controllers.GetJenkinsServerInfo)

			jenkinsServer.GET(":id/apiJobs", middleware.Authorize("jenkinsServer", "read"), controllers.GetJenkinsAPIJobs)
			jenkinsServer.GET(":id/apiJobs/*path", middleware.Authorize("jenkinsServer", "read"), controllers.GetJenkinsAPIJobs)

			jenkinsServer.POST(":id/addJob/*path", middleware.Authorize("jenkinsJob", "write"), controllers.AddJenkinsJobToServer)
			jenkinsServer.GET(":id/jobs", middleware.Authorize("jenkinsServer", "read"), controllers.GetJenkinsJobs)
			//jenkinsServer.GET(":id/job/*path", middleware.Authorize("jenkinsServer", "read"), controllers.GetJenkinsJob)
		}

		jenkinJob := config.Group("/jenkinsJob")
		{
			jenkinJob.GET("", middleware.Authorize("jenkinsJob", "read"), controllers.ListJenkinsJobs)
			jenkinJob.GET(":id", middleware.Authorize("jenkinsJob", "read"), controllers.GetJenkinsJob)
			jenkinJob.PUT(":id", middleware.Authorize("jenkinsJob", "write"), controllers.UpdateJenkinsJob)
			jenkinJob.POST(":id", middleware.Authorize("jenkinsJob", "write"), controllers.UpdateJenkinsJob)
			jenkinJob.POST(":id/updateFromAPI", middleware.Authorize("jenkinsJob", "write"), controllers.UpdateJenkinsJobFromAPI)
			jenkinJob.DELETE(":id", middleware.Authorize("jenkinsJob", "delete"), controllers.DeleteJenkinsJob)
		}

		jenkinJobParam := config.Group("/jenkinsJobParam")
		{
			// jenkinJobParam.GET("", middleware.Authorize("jenkinsJob", "read"), controllers.ListJenkinsJobParams)
			// jenkinJobParam.GET(":id", middleware.Authorize("jenkinsJob", "read"), controllers.GetJenkinsJobParam)
			jenkinJobParam.PUT(":id", middleware.Authorize("jenkinsJob", "write"), controllers.UpdateJenkinsJobParam)
			jenkinJobParam.POST(":id", middleware.Authorize("jenkinsJob", "write"), controllers.UpdateJenkinsJobParam)
			jenkinJobParam.DELETE(":id", middleware.Authorize("jenkinsJob", "delete"), controllers.DeleteJenkinsJobParam)
		}
	} // END config group

	log.Info("Starting server.")
	err := router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err != nil {
		log.Error("Error from Router: ", err)
	}
}
