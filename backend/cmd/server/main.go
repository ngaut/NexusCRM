package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/bootstrap"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/internal/interfaces/middleware"
	"github.com/nexuscrm/backend/internal/interfaces/rest"
	"github.com/nexuscrm/mcp/pkg/client"
	"github.com/nexuscrm/mcp/pkg/contextstore"
	"github.com/nexuscrm/mcp/pkg/mcp"
	mcp_models "github.com/nexuscrm/mcp/pkg/models"
	mcp_server "github.com/nexuscrm/mcp/pkg/server"
	"github.com/nexuscrm/shared/pkg/constants"
)

func main() {
	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001" // Default to 3001 (standard NexusCRM port)
	}

	// Initialize database connection
	db, err := database.GetInstance()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connection established")

	// Initialize core schema - SchemaManager auto-registers all tables in _System_Table during creation
	if err := bootstrap.InitializeSchema(db); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Initialize service manager
	svcMgr := services.NewServiceManager(db)
	log.Println("🔧 Service manager initialized")

	// Refresh metadata cache - Load all metadata before running bootstrap scripts
	if err := svcMgr.RefreshMetadataCache(); err != nil {
		log.Printf("⚠️  Warning: Failed to refresh metadata cache: %v", err)
	} else {
		log.Println("📦 Metadata cache loaded")
	}

	// Initialize system data (profiles, admin user, etc.)
	if err := bootstrap.InitializeSystemData(svcMgr.System); err != nil {
		log.Fatalf("Failed to initialize system data: %v", err)
	}

	// Initialize standard actions - Ensures Edit/Delete actions exist for core objects
	if err := bootstrap.InitializeStandardActions(svcMgr.Metadata); err != nil {
		log.Printf("⚠️  Warning: Failed to initialize standard actions: %v", err)
	}

	// Initialize default permissions for all profiles
	if err := bootstrap.InitializePermissions(svcMgr.Permissions, svcMgr.Metadata); err != nil {
		log.Printf("⚠️  Warning: Failed to initialize permissions: %v", err)
	}

	// Initialize standard themes
	if err := bootstrap.InitializeThemes(svcMgr); err != nil {
		log.Printf("⚠️  Warning: Failed to initialize themes: %v", err)
	}

	// Initialize UI components
	if err := bootstrap.InitializeUIComponents(svcMgr); err != nil {
		log.Printf("⚠️  Warning: Failed to initialize UI components: %v", err)
	}

	// Initialize setup pages
	if err := bootstrap.InitializeSetupPages(svcMgr); err != nil {
		log.Printf("⚠️  Warning: Failed to initialize setup pages: %v", err)
	}

	// Initialize flows
	if err := bootstrap.InitializeFlows(svcMgr.Metadata); err != nil {
		log.Printf("⚠️  Warning: Failed to initialize flows: %v", err)
	}

	// Run startup assertions to detect design violations
	// By default, violations are fatal (strict mode). Set SKIP_ASSERTIONS=true to skip.
	if os.Getenv("SKIP_ASSERTIONS") != "true" {
		if _, err := bootstrap.RunAssertions(db, true); err != nil {
			log.Fatalf("❌ Startup assertions failed: %v", err)
		}
	} else {
		log.Println("⚠️  Skipping startup assertions (SKIP_ASSERTIONS=true)")
	}

	// Create Gin router
	router := gin.Default()

	// CORS middleware - Allow credentials from any origin
	router.Use(middleware.Cors())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"server": "golang",
		})
	})

	// Debug/pprof endpoints for goroutine debugging
	// Access: http://localhost:3001/debug/pprof/
	// Goroutine stacks: http://localhost:3001/debug/pprof/goroutine?debug=2
	debug := router.Group("/debug/pprof")
	{
		debug.GET("/", gin.WrapF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/debug/pprof/", http.StatusMovedPermanently)
		})))
		debug.GET("/goroutine", gin.WrapH(http.DefaultServeMux))
		debug.GET("/heap", gin.WrapH(http.DefaultServeMux))
		debug.GET("/threadcreate", gin.WrapH(http.DefaultServeMux))
		debug.GET("/block", gin.WrapH(http.DefaultServeMux))
		debug.GET("/mutex", gin.WrapH(http.DefaultServeMux))
		debug.GET("/profile", gin.WrapH(http.DefaultServeMux))
		debug.GET("/trace", gin.WrapH(http.DefaultServeMux))
	}

	// Initialize handlers
	formulaHandler := rest.NewFormulaHandler()
	authHandler := rest.NewAuthHandler(svcMgr)
	userHandler := rest.NewUserHandler(svcMgr) // UserHandler init
	metadataHandler := rest.NewMetadataHandler(svcMgr)
	uiHandler := rest.NewUIHandler(svcMgr) // Add UIHandler initialization
	dataHandler := rest.NewDataHandler(svcMgr)
	actionHandler := rest.NewActionHandler(svcMgr)
	flowHandler := rest.NewFlowHandler(svcMgr)
	adminHandler := rest.NewAdminHandler(svcMgr)
	analyticsHandler := rest.NewAnalyticsHandler(svcMgr)
	fileHandler := rest.NewFileHandler(svcMgr)
	approvalHandler := rest.NewApprovalHandler(svcMgr.Approval)
	feedHandler := rest.NewFeedHandler(svcMgr)
	notificationHandler := rest.NewNotificationHandler(svcMgr)
	roleHandler := rest.NewRoleHandler(svcMgr)
	// Initialize Agent Handler (MCP-based)
	// Function to extract and map backend user to MCP user
	agentUserExtractor := func(c *gin.Context) *mcp_models.UserSession {
		user := rest.GetUserFromContext(c)
		if user == nil {
			return nil
		}
		// Map backend user to MCP user
		return &mcp_models.UserSession{
			ID:            user.ID,
			Name:          user.Name,
			Email:         user.Email,
			ProfileID:     user.ProfileID,
			IsSystemAdmin: user.IsSuperUser(),
		}
	}
	// Initialize MCP Handler
	// Decoupled: Create standalone ToolBus using NexusClient
	apiBaseURL := "http://localhost:3001"
	if url := os.Getenv("API_BASE_URL"); url != "" {
		apiBaseURL = url
	}
	mcpClient := client.NewNexusClient(apiBaseURL)

	// SHARED Context Store with Persistence
	// Use data/context_store.json for persistence (relative to CWD: backend)
	persistencePath := "data/context_store.json"
	sharedContextStore := contextstore.NewContextStore(persistencePath)

	toolBus := mcp_server.NewToolBusService(mcpClient, sharedContextStore)
	mcpHandler := mcp_server.NewHandler(toolBus)

	// Inject shared store into Agent Handler too
	agentHandler := mcp_server.NewAgentHandler(agentUserExtractor, sharedContextStore)

	// Initialize middleware
	requireAuth := middleware.RequireAuth(svcMgr.Auth)
	requireSystemAdmin := middleware.RequireSystemAdmin()

	// MCP Endpoint (Model Context Protocol)
	// Supports JSON-RPC 2.0 over HTTP
	// 1. Require Auth (Validates Bearer token)
	// 2. Propagate User Context (Gin -> Stdlib Context) -> WrapH(mcpHandler)
	router.POST("/mcp", requireAuth, func(c *gin.Context) {
		// Extract user from Gin context (set by RequireAuth)
		if user, exists := c.Get(constants.ContextKeyUser); exists {
			ctx := c.Request.Context()
			// Inject into standard context
			ctx = context.WithValue(ctx, constants.ContextKeyUser, user)

			// Inject Auth Token (needed for ContextStore)
			authHeader := c.GetHeader(constants.HeaderAuthorization)
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token := authHeader[7:]
				ctx = context.WithValue(ctx, mcp.ContextKeyAuthToken, token)
			}

			c.Request = c.Request.WithContext(ctx)
		}
		// Forward to standard HTTP handler
		mcpHandler.ServeHTTP(c.Writer, c.Request)
	})

	// API routes
	api := router.Group("/api")
	{
		// Public Auth routes (no authentication required)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", requireAuth, authHandler.Logout)
			auth.GET("/me", requireAuth, authHandler.GetMe)
			auth.GET("/permissions/me", requireAuth, authHandler.GetMyPermissions)
			auth.POST("/change-password", requireAuth, authHandler.ChangePassword)

			// Moved to UserHandler
			auth.POST("/register", requireAuth, requireSystemAdmin, userHandler.Register)
			auth.PUT("/users/:id", requireAuth, requireSystemAdmin, userHandler.UpdateUser)
			auth.DELETE("/users/:id", requireAuth, requireSystemAdmin, userHandler.DeleteUser)
			auth.GET("/users", requireAuth, userHandler.GetUsers)
			auth.GET("/profiles", requireAuth, userHandler.GetProfiles)
			auth.GET("/profiles/:id/permissions", requireAuth, userHandler.GetProfilePermissions)
			auth.PUT("/profiles/:id/permissions", requireAuth, requireSystemAdmin, userHandler.UpdateProfilePermissions)
			auth.GET("/profiles/:id/permissions/fields", requireAuth, userHandler.GetProfileFieldPermissions)
			auth.PUT("/profiles/:id/permissions/fields", requireAuth, requireSystemAdmin, userHandler.UpdateProfileFieldPermissions)

			// Permission Set permissions
			auth.POST("/permission-sets", requireAuth, requireSystemAdmin, userHandler.CreatePermissionSet)
			auth.PUT("/permission-sets/:id", requireAuth, requireSystemAdmin, userHandler.UpdatePermissionSet)
			auth.DELETE("/permission-sets/:id", requireAuth, requireSystemAdmin, userHandler.DeletePermissionSet)

			auth.GET("/permission-sets/:id/permissions", requireAuth, userHandler.GetPermissionSetPermissions)
			auth.PUT("/permission-sets/:id/permissions", requireAuth, requireSystemAdmin, userHandler.UpdatePermissionSetPermissions)
			auth.GET("/permission-sets/:id/permissions/fields", requireAuth, userHandler.GetPermissionSetFieldPermissions)
			auth.PUT("/permission-sets/:id/permissions/fields", requireAuth, requireSystemAdmin, userHandler.UpdatePermissionSetFieldPermissions)

			// Effective Permissions (User)
			auth.GET("/users/:id/permissions/effective", requireAuth, requireSystemAdmin, userHandler.GetUserEffectivePermissions)
			auth.GET("/users/:id/permissions/fields/effective", requireAuth, requireSystemAdmin, userHandler.GetUserEffectiveFieldPermissions)

			// Role Management routes
			auth.POST("/roles", requireAuth, requireSystemAdmin, roleHandler.CreateRole)
			auth.GET("/roles", requireAuth, roleHandler.GetRoles)
			auth.GET("/roles/:id", requireAuth, roleHandler.GetRole)
			auth.PUT("/roles/:id", requireAuth, requireSystemAdmin, roleHandler.UpdateRole)
			auth.DELETE("/roles/:id", requireAuth, requireSystemAdmin, roleHandler.DeleteRole)
		}

		// Protected Formula routes
		formula := api.Group("/formula")
		formula.Use(requireAuth)
		{
			formula.POST("/evaluate", formulaHandler.Evaluate)
			formula.POST("/condition", formulaHandler.EvaluateCondition)
			formula.POST("/substitute", formulaHandler.Substitute)
			formula.POST("/validate", formulaHandler.Validate)
			formula.GET("/functions", formulaHandler.GetFunctions)
			formula.DELETE("/cache", formulaHandler.ClearCache)
		}

		// Admin routes (system admin only)
		admin := api.Group("/admin")
		admin.Use(requireAuth, requireSystemAdmin)
		{
			admin.GET("/tables", adminHandler.GetTableRegistry)
			admin.POST("/validate-schema", adminHandler.ValidateSchema)
		}

		// Protected Metadata routes
		metadata := api.Group("/metadata")
		metadata.Use(requireAuth)
		{
			metadata.GET("/apps", uiHandler.GetApps)
			metadata.POST("/apps", requireSystemAdmin, uiHandler.CreateApp)
			metadata.PATCH("/apps/:id", requireSystemAdmin, uiHandler.UpdateApp)
			metadata.DELETE("/apps/:id", requireSystemAdmin, uiHandler.DeleteApp)

			metadata.GET("/themes/active", uiHandler.GetActiveTheme) // New standard endpoint

			metadata.POST("/themes", requireSystemAdmin, uiHandler.CreateTheme)
			metadata.PUT("/themes/:id/activate", requireSystemAdmin, uiHandler.ActivateTheme)

			metadata.GET("/objects", metadataHandler.GetSchemas)
			metadata.POST("/objects", requireSystemAdmin, metadataHandler.CreateSchema)
			metadata.GET("/objects/:apiName", metadataHandler.GetSchema)
			metadata.PATCH("/objects/:apiName", requireSystemAdmin, metadataHandler.UpdateSchema)
			metadata.DELETE("/objects/:apiName", requireSystemAdmin, metadataHandler.DeleteSchema)
			metadata.POST("/objects/:apiName/fields", requireSystemAdmin, metadataHandler.CreateField)
			metadata.PATCH("/objects/:apiName/fields/:fieldApiName", requireSystemAdmin, metadataHandler.UpdateField)
			metadata.DELETE("/objects/:apiName/fields/:fieldApiName", requireSystemAdmin, metadataHandler.DeleteField)
			metadata.GET("/layouts/:objectName", uiHandler.GetLayout)
			metadata.POST("/layouts", uiHandler.SaveLayout)
			metadata.DELETE("/layouts/:id", uiHandler.DeleteLayout)
			metadata.POST("/layouts/assign", uiHandler.AssignLayoutToProfile)
			metadata.GET("/actions/:objectName", actionHandler.GetActions)
			metadata.GET("/actions", actionHandler.GetAllActions)
			metadata.GET("/actions/id/:actionId", actionHandler.GetAction)
			metadata.POST("/actions", actionHandler.CreateAction)
			metadata.PATCH("/actions/:actionId", actionHandler.UpdateAction)
			metadata.DELETE("/actions/:actionId", actionHandler.DeleteAction)

			// Dashboards
			metadata.GET("/dashboards", uiHandler.GetDashboards)
			metadata.POST("/dashboards", uiHandler.CreateDashboard)
			metadata.GET("/dashboards/:id", uiHandler.GetDashboard)
			metadata.PATCH("/dashboards/:id", uiHandler.UpdateDashboard)
			metadata.DELETE("/dashboards/:id", uiHandler.DeleteDashboard)

			// List Views
			metadata.GET("/listviews", uiHandler.GetListViews)
			metadata.POST("/listviews", uiHandler.CreateListView)
			metadata.PATCH("/listviews/:id", uiHandler.UpdateListView)
			metadata.DELETE("/listviews/:id", uiHandler.DeleteListView)

			// Validation Rules
			metadata.GET("/validation-rules", metadataHandler.GetValidationRules)
			metadata.POST("/validation-rules", requireSystemAdmin, metadataHandler.CreateValidationRule)
			metadata.PATCH("/validation-rules/:id", requireSystemAdmin, metadataHandler.UpdateValidationRule)
			metadata.DELETE("/validation-rules/:id", requireSystemAdmin, metadataHandler.DeleteValidationRule)

			// Field Types (includes plugins)
			metadata.GET("/fieldtypes", metadataHandler.GetFieldTypes)

			// Flows
			metadata.GET("/flows", flowHandler.GetAllFlows)
			metadata.GET("/flows/:flowId", flowHandler.GetFlow)
			metadata.POST("/flows", flowHandler.CreateFlow)
			metadata.PATCH("/flows/:flowId", flowHandler.UpdateFlow)
			metadata.DELETE("/flows/:flowId", flowHandler.DeleteFlow)
		}

		// Protected Action routes
		actions := api.Group("/actions")
		actions.Use(requireAuth)
		{
			actions.POST("/execute/:actionId", actionHandler.ExecuteAction)
		}

		// Protected Flow Execution routes (Auto-Launched Flows)
		flows := api.Group("/flows")
		flows.Use(requireAuth)
		{
			flows.POST("/:flowId/execute", flowHandler.ExecuteFlow)
		}

		// Protected Data routes
		data := api.Group("/data")
		data.Use(requireAuth)
		{
			data.POST("/query", dataHandler.Query)
			data.POST("/analytics", dataHandler.RunAnalytics)
			data.POST("/search", dataHandler.Search)
			data.GET("/recyclebin/items", dataHandler.GetRecycleBinItems)
			data.POST("/recyclebin/restore/:id", dataHandler.RestoreFromRecycleBin)
			data.DELETE("/recyclebin/:id", dataHandler.PurgeFromRecycleBin)
			// Single object search - MUST be before /:objectApiName/:id to avoid conflict
			data.GET("/search/:objectApiName", dataHandler.SearchSingleObject)
			data.POST("/:objectApiName/calculate", dataHandler.Calculate)
			data.GET("/:objectApiName/:id", dataHandler.GetRecord)
			data.POST("/:objectApiName", dataHandler.CreateRecord)
			data.POST("/:objectApiName/bulk", dataHandler.BulkCreateRecords)
			data.PATCH("/:objectApiName/:id", dataHandler.UpdateRecord)
			data.DELETE("/:objectApiName/:id", dataHandler.DeleteRecord)
		}
		// Protected Analytics routes (System Admin Only)
		analytics := api.Group("/analytics")
		analytics.Use(requireAuth, requireSystemAdmin)
		{
			analytics.POST("/query", analyticsHandler.ExecuteAdminQuery)
		}

		// Protected File routes
		files := api.Group("/files")
		files.Use(requireAuth)
		{
			files.POST("/upload", fileHandler.Upload)
		}

		// Protected Approval routes
		approvals := api.Group("/approvals")
		approvals.Use(requireAuth)
		{
			approvals.POST("/submit", approvalHandler.Submit)
			approvals.POST("/:workItemId/approve", approvalHandler.Approve)
			approvals.POST("/:workItemId/reject", approvalHandler.Reject)
			approvals.GET("/pending", approvalHandler.GetPending)
			approvals.GET("/check/:objectApiName", approvalHandler.CheckProcess)
			approvals.GET("/history/:objectApiName/:recordId", approvalHandler.GetHistory)
			approvals.GET("/flow-progress/:instanceId", approvalHandler.GetFlowProgress)
		}

		// Protected Feed routes
		feed := api.Group("/feed")
		feed.Use(requireAuth)
		{
			feed.POST("/comments", feedHandler.CreateComment)
			feed.GET("/:recordId", feedHandler.GetComments)
		}

		// Protected Notification routes
		notifications := api.Group("/notifications")
		notifications.Use(requireAuth)
		{
			notifications.GET("/", notificationHandler.GetNotifications)
			notifications.POST("/:id/read", notificationHandler.MarkAsRead)
		}

		// Protected Setup routes
		setup := api.Group("/setup")
		setup.Use(requireAuth)
		{
			setup.GET("/pages", uiHandler.GetSetupPages)
		}

		// Protected Agent routes
		agent := api.Group("/agent")
		agent.Use(requireAuth)
		{
			agent.POST("/chat/stream", agentHandler.ChatStream)
			agent.GET("/context", agentHandler.GetContext)
			agent.POST("/compact", agentHandler.CompactContext)
			// Conversation persistence
			agent.GET("/conversation", agentHandler.GetConversation)
			agent.POST("/conversation", agentHandler.SaveConversation)
			agent.DELETE("/conversation", agentHandler.ClearConversation)
			// Multiple conversations
			agent.GET("/conversations", agentHandler.ListConversations)
			agent.DELETE("/conversations/:id", agentHandler.DeleteConversation)
		}
	}

	// Static Files
	router.Static("/uploads", "./uploads")

	// Start background workers
	svcMgr.StartOutboxWorker()
	log.Println("📤 Outbox event worker started (500ms polling)")

	// Start scheduled job executor
	svcMgr.StartScheduler()
	log.Println("⏰ Scheduler service started (60s polling)")

	// Start server
	log.Println("\n═══════════════════════════════════════════════════════════════════════════")
	log.Println("🚀 NexusCRM Golang Backend Started Successfully")
	log.Println("═══════════════════════════════════════════════════════════════════════════")
	log.Printf("\n📍 Server:         http://localhost:%s", port)
	log.Printf("🔐 Auth API:       http://localhost:%s/api/auth", port)
	log.Printf("📐 Formula API:    http://localhost:%s/api/formula", port)
	log.Printf("📊 Metadata API:   http://localhost:%s/api/metadata", port)
	log.Printf("🤖 MCP Endpoint:   http://localhost:%s/mcp", port)
	log.Printf("💾 Data API:       http://localhost:%s/api/data", port)
	log.Printf("💚 Health check:   http://localhost:%s/health\n", port)

	// Create HTTP Server
	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Stop background workers
	svcMgr.StopOutboxWorker()
	log.Println("🛑 Outbox worker stopped")
	svcMgr.StopScheduler()
	log.Println("🛑 Scheduler stopped")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
