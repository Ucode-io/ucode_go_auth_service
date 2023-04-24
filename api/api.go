package api

import (
	"ucode/ucode_go_auth_service/api/docs"
	"ucode/ucode_go_auth_service/api/handlers"
	"ucode/ucode_go_auth_service/config"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// SetUpRouter godoc
// @description This is a api gateway
// @termsOfService https://udevs.io
func SetUpRouter(h handlers.Handler, cfg config.Config) (r *gin.Engine) {
	r = gin.New()

	r.Use(gin.Logger(), gin.Recovery())

	docs.SwaggerInfo.Title = cfg.ServiceName
	docs.SwaggerInfo.Version = cfg.Version
	// docs.SwaggerInfo.Host = cfg.ServiceHost + cfg.HTTPPort
	docs.SwaggerInfo.Schemes = []string{cfg.HTTPScheme}

	r.Use(customCORSMiddleware())

	r.GET("/ping", h.Ping)
	r.GET("/config", h.GetConfig)

	// CLIENT SERVICE
	// (admin, bot, mobile ext)
	r.POST("/client-platform", h.CreateClientPlatform)
	r.GET("/client-platform", h.GetClientPlatformList)
	r.GET("/client-platform/:client-platform-id", h.GetClientPlatformByID)
	r.GET("/client-platform-detailed/:client-platform-id", h.GetClientPlatformByIDDetailed)
	r.PUT("/client-platform", h.UpdateClientPlatform)
	r.DELETE("/client-platform/:client-platform-id", h.DeleteClientPlatform)

	// admin, dev, hr, ceo
	r.POST("/client-type", h.CreateClientType)
	r.GET("/client-type", h.GetClientTypeList)
	r.GET("/client-type/:client-type-id", h.GetClientTypeByID)
	r.PUT("/client-type", h.UpdateClientType)
	r.DELETE("/client-type/:client-type-id", h.DeleteClientType)

	r.POST("/client", h.AddClient)
	r.GET("/client/:project-id", h.GetClientMatrix)
	r.PUT("/client", h.UpdateClient)
	r.DELETE("/client", h.RemoveClient)

	r.POST("/relation", h.AddRelation)
	r.PUT("/relation", h.UpdateRelation)
	r.DELETE("/relation/:relation-id", h.RemoveRelation)

	r.POST("/user-info-field", h.AddUserInfoField)
	r.PUT("/user-info-field", h.UpdateUserInfoField)
	r.DELETE("/user-info-field/:user-info-field-id", h.RemoveUserInfoField)

	// PERMISSION SERVICE
	r.GET("/role/:role-id", h.GetRoleByID)
	r.GET("/role", h.GetRolesList)
	r.POST("/role", h.AddRole)
	r.PUT("/role", h.UpdateRole)
	r.DELETE("/role/:role-id", h.RemoveRole)

	r.POST("/permission", h.CreatePermission)
	r.GET("/permission", h.GetPermissionList)
	r.GET("/permission/:permission-id", h.GetPermissionByID)
	r.PUT("/permission", h.UpdatePermission)
	r.DELETE("/permission/:permission-id", h.DeletePermission)

	r.POST("/upsert-scope", h.UpsertScope)

	r.POST("/permission-scope", h.AddPermissionScope)
	r.DELETE("/permission-scope", h.RemovePermissionScope)

	r.POST("/role-permission", h.AddRolePermission)
	r.POST("/role-permission/many", h.AddRolePermissions)
	r.DELETE("/role-permission", h.RemoveRolePermission)

	r.POST("/user", h.CreateUser)
	r.GET("/user", h.GetUserList)
	r.GET("/user/:user-id", h.GetUserByID)
	r.PUT("/user", h.UpdateUser)
	r.DELETE("/user/:user-id", h.DeleteUser)
	r.PUT("/user/reset-password", h.ResetPassword)
	r.POST("/user/send-message", h.SendMessageToUserEmail)

	r.POST("/integration", h.CreateIntegration)
	r.GET("/integration", h.GetIntegrationList)
	r.GET("/integration/:integration-id", h.GetIntegrationByID)
	r.DELETE("/integration/:integration-id", h.DeleteIntegration)
	r.GET("/integration/:integration-id/session", h.GetIntegrationSessions)
	r.POST("/integration/:integration-id/session", h.AddSessionToIntegration)
	r.GET("/integration/:integration-id/session/:session-id", h.GetIntegrationToken)
	r.DELETE("/integration/:integration-id/session/:session-id", h.RemoveSessionFromIntegration)

	r.POST("/user-relation", h.AddUserRelation)
	r.DELETE("/user-relation", h.RemoveUserRelation)

	r.POST("/upsert-user-info/:user-id", h.UpsertUserInfo)

	r.POST("/login", h.Login)
	r.DELETE("/logout", h.Logout)
	r.PUT("/refresh", h.RefreshToken)
	r.POST("/has-acess", h.HasAccess)
	r.POST("/has-access-super-admin", h.HasAccessSuperAdmin)

	v2 := r.Group("/v2")
	v2.Use(h.AuthMiddleware())
	{
		v2.POST("/client-platform", h.V2CreateClientPlatform)
		v2.GET("/client-platform", h.V2GetClientPlatformList) //project_id
		v2.GET("/client-platform/:client-platform-id", h.V2GetClientPlatformByID)
		v2.GET("/client-platform-detailed/:client-platform-id", h.V2GetClientPlatformByIDDetailed)
		v2.PUT("/client-platform", h.V2UpdateClientPlatform)
		v2.DELETE("/client-platform/:client-platform-id", h.V2DeleteClientPlatform)

		// admin, dev, hr, ceo
		v2.POST("/client-type", h.V2CreateClientType)
		v2.GET("/client-type", h.V2GetClientTypeList) //
		v2.GET("/client-type/:client-type-id", h.V2GetClientTypeByID)
		v2.PUT("/client-type", h.V2UpdateClientType)
		v2.DELETE("/client-type/:client-type-id", h.V2DeleteClientType)

		v2.POST("/client", h.V2AddClient)
		v2.GET("/client/:project-id", h.V2GetClientMatrix)
		v2.PUT("/client", h.V2UpdateClient)
		v2.DELETE("/client", h.V2RemoveClient)

		v2.POST("/user-info-field", h.V2AddUserInfoField)
		v2.PUT("/user-info-field", h.V2UpdateUserInfoField)
		v2.DELETE("/user-info-field/:user-info-field-id", h.V2RemoveUserInfoField)

		// PERMISSION SERVICE
		v2.GET("/role/:role-id", h.V2GetRoleByID)
		v2.GET("/role", h.V2GetRolesList)
		v2.POST("/role", h.V2AddRole)
		v2.PUT("/role", h.V2UpdateRole)
		v2.DELETE("/role/:role-id", h.V2RemoveRole)

		v2.POST("/permission", h.V2CreatePermission)
		v2.GET("/permission", h.V2GetPermissionList)
		v2.GET("/permission/:permission-id", h.V2GetPermissionByID)
		v2.PUT("/permission", h.V2UpdatePermission)
		v2.DELETE("/permission/:permission-id", h.V2DeletePermission)

		v2.POST("/permission-scope", h.V2AddPermissionScope)
		v2.DELETE("/permission-scope", h.V2RemovePermissionScope)

		v2.POST("/role-permission", h.V2AddRolePermission)
		v2.DELETE("/role-permission", h.V2RemoveRolePermission)

		v2.GET("/role-permission/detailed/:project-id/:role-id", h.GetListWithRoleAppTablePermissions)
		v2.PUT("/role-permission/detailed", h.UpdateRoleAppTablePermissions)

		v2.POST("/user", h.V2CreateUser)
		v2.GET("/user", h.V2GetUserList)
		v2.GET("/user/:user-id", h.V2GetUserByID)
		v2.PUT("/user", h.V2UpdateUser)
		v2.DELETE("/user/:user-id", h.V2DeleteUser)
		v2.POST("/login", h.V2Login)
		v2.PUT("/refresh", h.V2RefreshToken)
		v2.PUT("/refresh-superadmin", h.V2RefreshTokenSuperAdmin)
		v2.POST("/login/superadmin", h.V2LoginSuperAdmin)
		v2.POST("/multi-company/login", h.V2MultiCompanyLogin)
		v2.POST("/user/invite", h.AddUserToProject)
		v2.POST("/user/check", h.V2GetUserByLoginType)

		// api keys
		v2.POST("/api-key/:project-id", h.CreateApiKey)
		v2.PUT("/api-key/:project-id/:id", h.UpdateApiKey)
		v2.GET("/api-key/:project-id/:id", h.GetApiKey)
		v2.GET("/api-key/:project-id", h.GetListApiKeys)
		v2.DELETE("/api-key/:project-id/:id", h.DeleteApiKeys)
		v2.POST("/api-key/generate-token", h.GenerateApiKeyToken)
		v2.POST("/api-key/refresh-token", h.RefreshApiKeyToken)

		// environment
		v2.GET("/resource-environment", h.GetAllResourceEnvironments)
		v2.GET("/webpage-app", h.GetListWebPageApp)

		// connection
		v2.POST("/connection", h.V2CreateConnection)
		v2.GET("/connection", h.V2GetConnectionList)
		v2.GET("/connection/:connection_id", h.V2GetConnectionByID)
		v2.PUT("/connection", h.V2UpdateConnection)
		v2.DELETE("/connection/:connection_id", h.V2DeleteConnection)

		// objects
		v2.POST("/object/get-list/:table_slug", h.V2GetListObjects)
	}

	//COMPANY
	r.POST("/company", h.RegisterCompany)
	r.PUT("/company", h.UpdateCompany)
	r.DELETE("/company/:company-id", h.RemoveCompany)

	//PROJECT
	r.POST("/project", h.CreateProject)
	r.PUT("/project", h.UpdateProject)
	r.PUT("/project/:project-id/user-update", h.UpdateProjectUserData)
	r.GET("/project", h.GetProjectList)
	r.GET("project/:project-id", h.GetProjectByID)
	r.DELETE("/project/:project-id", h.DeleteProject)

	r.POST("/send-code", h.SendCode)
	r.POST("/verify/:sms_id/:otp", h.Verify)
	r.POST("/register-otp/:table_slug", h.RegisterOtp)

	// With API-KEY authentication
	v2.POST("/send-message", h.SendMessageToEmail)
	v2.POST("/verify-email/:sms_id/:otp", h.VerifyEmail)
	v2.POST("/register-email-otp/:table_slug", h.RegisterEmailOtp)

	v2.POST("/email-settings", h.CreateEmailSettings)
	v2.PUT("/email-settings", h.UpdateEmailSettings)
	v2.GET("/email-settings", h.GetEmailSettings)
	v2.DELETE("/email-settings/:id", h.DeleteEmailSettings)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return
}

func customCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Resource-Id, Environment-Id, X-API-KEY, Platform-Type")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
