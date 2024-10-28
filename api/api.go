package api

import (
	"ucode/ucode_go_auth_service/api/docs"
	"ucode/ucode_go_auth_service/api/handlers"
	"ucode/ucode_go_auth_service/config"

	"github.com/gin-gonic/gin"
	"github.com/golanguzb70/ratelimiter"
	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing/opentracing-go"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// SetUpRouter godoc
// @description This is a api gateway
// @termsOfService https://udevs.io
func SetUpRouter(h handlers.Handler, cfg config.BaseConfig, tracer opentracing.Tracer, limiter ratelimiter.RateLimiterI) (r *gin.Engine) {
	r = gin.New()

	r.Use(gin.Logger(), gin.Recovery())
	r.Use(ginhttp.Middleware(tracer))

	docs.SwaggerInfo.Title = cfg.ServiceName
	docs.SwaggerInfo.Version = cfg.Version
	docs.SwaggerInfo.Schemes = []string{cfg.HTTPScheme}

	// @securityDefinitions.apikey ApiKeyAuth
	// @in header
	// @name Authorization
	r.Use(customCORSMiddleware())

	v2 := r.Group("/v2")
	v2.POST("/login/superadmin", h.V2LoginSuperAdmin)
	v2.PUT("/refresh", h.V2RefreshToken)

	v2.Use(h.LoginMiddleware())
	{
		v2.GET("/connection", h.V2GetConnectionList)
		v2.POST("/multi-company/one-login", h.V2MultiCompanyOneLogin)
		v2.POST("/login", h.V2Login)
	}

	v2.Use(h.AuthMiddleware())
	{

		// register
		v2.POST("/register", h.V2Register)

		v2.POST("/login/with-option", h.V2LoginWithOption)
		v2.POST("/send/message", h.SendMessage)
		v2.POST("/send-code-app", h.V2SendCodeApp)
		v2.POST("/forgot-password", h.ForgotPassword)
		v2.POST("/forgot-password-with-environment-email", h.ForgotPasswordWithEnvironmentEmail)
		v2.PUT("/reset-password", h.V2ResetPassword)
		v2.PUT("set-email/send-code", h.EmailEnter)
		v2.PUT("/expire-sessions", h.ExpireSessions)

		v2.PUT("/refresh-superadmin", h.V2RefreshTokenSuperAdmin)
		v2.POST("/multi-company/login", h.V2MultiCompanyLogin) // @TODO

		//connection
		v2.POST("/connection", h.V2CreateConnection)
		v2.GET("/connection/:connection_id", h.V2GetConnectionByID)
		v2.PUT("/connection", h.V2UpdateConnection)
		v2.DELETE("/connection/:connection_id", h.V2DeleteConnection)
		v2.GET("/get-connection-options/:connection_id/:user_id", h.GetConnectionOptions)

		// admin, dev, hr, ceo
		v2.POST("/client-type", h.V2CreateClientType)
		v2.GET("/client-type", h.V2GetClientTypeList)
		v2.GET("/client-type/:client-type-id", h.V2GetClientTypeByID)
		v2.PUT("/client-type", h.V2UpdateClientType)
		v2.DELETE("/client-type/:client-type-id", h.V2DeleteClientType)

		// ROLE SERVICE
		v2.GET("/role/:role-id", h.V2GetRoleByID)
		v2.GET("/role", h.V2GetRolesList)
		v2.POST("/role", h.V2AddRole)
		v2.DELETE("/role/:role-id", h.V2RemoveRole)

		// role-permission
		v2.GET("/role-permission/detailed/:project-id/:role-id", h.GetListWithRoleAppTablePermissions)
		v2.PUT("/role-permission/detailed", h.UpdateRoleAppTablePermissions)

		// menu-permission
		v2.GET("/menu-permission/detailed/:project-id/:role-id/:parent-id", h.GetListMenuPermissions)
		v2.PUT("/menu-permission/detailed", h.UpdateMenuPermissions)

		// user
		v2.POST("/user", h.V2CreateUser)
		v2.GET("/user", h.V2GetUserList)
		v2.GET("/user/:user-id", h.V2GetUserByID)
		v2.PUT("/user", h.V2UpdateUser)
		v2.DELETE("/user/:user-id", h.V2DeleteUser)
		v2.PUT("/user/reset-password", h.V2UserResetPassword)
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

		// objects
		v2.POST("/object/get-list/:table_slug", h.V2GetListObjects)

		// login strategy
		v2.GET("/login-strategy", h.GetLoginStrategy)
		v2.GET("/login-strategy/:login-strategy-id", h.GetLoginStrategyById)
		v2.POST("/upsert-login-strategy", h.UpsertLoginStrategy)
	}

	auth := v2.Group("/auth")
	{
		auth.POST("/register/:provider", h.V2RegisterProvider)
		auth.POST("/verify/:verify_id", h.V2VerifyOtp)
		auth.POST("/login/:provider", h.V2LoginProvider)
		auth.POST("/refresh", h.V2RefreshToken)
		auth.POST("/send-code", h.V2SendCode)
		auth.POST("/logout", h.V2Logout)
		auth.POST("/password/reset", h.V2UserResetPassword)
	}

	v2Sms := r.Group("/v2")
	v2Sms.Use(h.AuthMiddleware(), limiter.GinMiddleware())
	{
		v2Sms.POST("/send-code", h.V2SendCode)
	}

	//COMPANY
	r.POST("/company", h.RegisterCompany)
	r.PUT("/company", h.UpdateCompany)
	r.DELETE("/company/:company-id", h.RemoveCompany)

	//PROJECT
	r.POST("/project", h.CreateProject)
	r.PUT("/project", h.UpdateProject)
	r.PUT("/project/:project-id/user-update", h.UpdateProjectUserData)
	r.GET("project/:project-id", h.GetProjectByID)
	r.DELETE("/project/:project-id", h.DeleteProject)

	// With API-KEY authentication
	v2.POST("/send-message", h.SendMessageToEmail)
	v2.POST("/verify-email/:sms_id/:otp", h.VerifyEmail)
	v2.POST("/verify-only-email", h.VerifyOnlyEmailOtp)
	v2.POST("/register-email-otp/:table_slug", h.RegisterEmailOtp)

	v2.POST("/email-settings", h.CreateEmailSettings)
	v2.PUT("/email-settings", h.UpdateEmailSettings)
	v2.GET("/email-settings", h.GetEmailSettings)
	v2.DELETE("/email-settings/:id", h.DeleteEmailSettings)

	// sms-otp-settings
	v2.POST("/sms-otp-settings", h.CreateSmsOtpSettings)
	v2.GET("/sms-otp-settings", h.GetListSmsOtpSettings)
	v2.GET("/sms-otp-settings/:id", h.GetByIdSmsOtpSettings)
	v2.PUT("/sms-otp-settings", h.UpdateSmsOtpSettings)
	v2.DELETE("/sms-otp-settings/:id", h.DeleteSmsOtpSettings)

	v2.POST("/apple-id-settings", h.CreateAppleIdSettings)
	v2.PUT("/apple-id-settings", h.UpdateAppleIdSettings)
	v2.GET("/apple-id-settings", h.GetAppleIdSettings)
	v2.DELETE("/apple-id-settings/:id", h.DeleteAppleIdSettings)

	v2.POST("/login-platform-type", h.CreateLoginPlatformType)
	v2.PUT("/login-platform-type", h.UpdateLoginPlatformType)
	v2.GET("/login-platform-type", h.GetLoginPlatformType)
	v2.GET("/login-platform-type/:id", h.LoginPlatformTypePrimaryKey)
	v2.DELETE("/login-platform-type/:id", h.DeleteLoginPlatformType)

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

//
