package client

import (
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/genproto/sms_service"
	"ucode/ucode_go_auth_service/genproto/web_page_service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceManagerI interface {
	IntegrationService() auth_service.IntegrationServiceClient
	ClientService() auth_service.ClientServiceClient
	PermissionService() auth_service.PermissionServiceClient
	UserService() auth_service.UserServiceClient
	SessionService() auth_service.SessionServiceClient
	ObjectBuilderService() object_builder_service.ObjectBuilderServiceClient
	SmsService() sms_service.SmsServiceClient
	LoginService() object_builder_service.LoginServiceClient
	EmailService() auth_service.EmailOtpServiceClient
	CompanyService() auth_service.CompanyServiceClient
	ProjectService() auth_service.ProjectServiceClient
	CompanyServiceClient() company_service.CompanyServiceClient
	ProjectServiceClient() company_service.ProjectServiceClient
	BuilderPermissionService() object_builder_service.PermissionServiceClient
	ApiKeysService() auth_service.ApiKeysClient
	ResourceService() company_service.ResourceServiceClient
	EnvironmentService() company_service.EnvironmentServiceClient
	MicroServiceResourceService() company_service.MicroserviceResourceClient
	WebPageAppService() web_page_service.AppServiceClient
	ServiceResource() company_service.MicroserviceResourceClient
	AppleIdService()  auth_service.AppleIdLoginServiceClient
}

type grpcClients struct {
	integrationService          auth_service.IntegrationServiceClient
	clientService               auth_service.ClientServiceClient
	permissionService           auth_service.PermissionServiceClient
	userService                 auth_service.UserServiceClient
	sessionService              auth_service.SessionServiceClient
	objectBuilderService        object_builder_service.ObjectBuilderServiceClient
	smsService                  sms_service.SmsServiceClient
	loginService                object_builder_service.LoginServiceClient
	emailService                auth_service.EmailOtpServiceClient
	companyService              auth_service.CompanyServiceClient
	projectService              auth_service.ProjectServiceClient
	companyServiceClient        company_service.CompanyServiceClient
	projectServiceClient        company_service.ProjectServiceClient
	builderPermissionService    object_builder_service.PermissionServiceClient
	apiKeysClients              auth_service.ApiKeysClient
	resourceService             company_service.ResourceServiceClient
	environmentService          company_service.EnvironmentServiceClient
	microServiceResourceService company_service.MicroserviceResourceClient
	webPageAppService           web_page_service.AppServiceClient
	serviceResource             company_service.MicroserviceResourceClient
	appleIdService              auth_service.AppleIdLoginServiceClient
}

func NewGrpcClients(cfg config.Config) (ServiceManagerI, error) {
	connAuthService, err := grpc.Dial(
		cfg.AuthServiceHost+cfg.AuthGRPCPort,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	connObjectBuilderService, err := grpc.Dial(
		cfg.ObjectBuilderServiceHost+cfg.ObjectBuilderGRPCPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(52428800), grpc.MaxCallSendMsgSize(52428800)),
	)
	if err != nil {
		return nil, err
	}

	connSmsService, err := grpc.Dial(
		cfg.SmsServiceHost+cfg.SmsGRPCPort,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	connCompanyService, err := grpc.Dial(
		cfg.CompanyServiceHost+cfg.CompanyGRPCPort,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	connWebPageService, err := grpc.Dial(
		cfg.WebPageServiceHost+cfg.WebPageServicePort,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return &grpcClients{
		clientService:               auth_service.NewClientServiceClient(connAuthService),
		permissionService:           auth_service.NewPermissionServiceClient(connAuthService),
		userService:                 auth_service.NewUserServiceClient(connAuthService),
		sessionService:              auth_service.NewSessionServiceClient(connAuthService),
		integrationService:          auth_service.NewIntegrationServiceClient(connAuthService),
		objectBuilderService:        object_builder_service.NewObjectBuilderServiceClient(connObjectBuilderService),
		smsService:                  sms_service.NewSmsServiceClient(connSmsService),
		loginService:                object_builder_service.NewLoginServiceClient(connObjectBuilderService),
		emailService:                auth_service.NewEmailOtpServiceClient(connAuthService),
		companyService:              auth_service.NewCompanyServiceClient(connAuthService),
		projectService:              auth_service.NewProjectServiceClient(connAuthService),
		companyServiceClient:        company_service.NewCompanyServiceClient(connCompanyService),
		projectServiceClient:        company_service.NewProjectServiceClient(connCompanyService),
		builderPermissionService:    object_builder_service.NewPermissionServiceClient(connObjectBuilderService),
		apiKeysClients:              auth_service.NewApiKeysClient(connAuthService),
		resourceService:             company_service.NewResourceServiceClient(connCompanyService),
		environmentService:          company_service.NewEnvironmentServiceClient(connCompanyService),
		microServiceResourceService: company_service.NewMicroserviceResourceClient(connCompanyService),
		webPageAppService:           web_page_service.NewAppServiceClient(connWebPageService),
		serviceResource:             company_service.NewMicroserviceResourceClient(connCompanyService),
		appleIdService:              auth_service.NewAppleIdLoginServiceClient(connAuthService),
	}, nil
}

func (g *grpcClients) ClientService() auth_service.ClientServiceClient {
	return g.clientService
}

func (g *grpcClients) EmailService() auth_service.EmailOtpServiceClient {
	return g.emailService
}

func (g *grpcClients) PermissionService() auth_service.PermissionServiceClient {
	return g.permissionService
}

func (g *grpcClients) UserService() auth_service.UserServiceClient {
	return g.userService
}

func (g *grpcClients) SessionService() auth_service.SessionServiceClient {
	return g.sessionService
}

func (g *grpcClients) IntegrationService() auth_service.IntegrationServiceClient {
	return g.integrationService
}

func (g *grpcClients) ObjectBuilderService() object_builder_service.ObjectBuilderServiceClient {
	return g.objectBuilderService
}

func (g *grpcClients) SmsService() sms_service.SmsServiceClient {
	return g.smsService
}

func (g *grpcClients) LoginService() object_builder_service.LoginServiceClient {
	return g.loginService
}

func (g *grpcClients) CompanyService() auth_service.CompanyServiceClient {
	return g.companyService
}

func (g *grpcClients) ProjectService() auth_service.ProjectServiceClient {
	return g.projectService
}

func (g *grpcClients) CompanyServiceClient() company_service.CompanyServiceClient {
	return g.companyServiceClient
}

func (g *grpcClients) ProjectServiceClient() company_service.ProjectServiceClient {
	return g.projectServiceClient
}

func (g *grpcClients) BuilderPermissionService() object_builder_service.PermissionServiceClient {
	return g.builderPermissionService
}

func (g *grpcClients) ApiKeysService() auth_service.ApiKeysClient {
	return g.apiKeysClients
}

func (g *grpcClients) ResourceService() company_service.ResourceServiceClient {
	return g.resourceService
}

func (g *grpcClients) EnvironmentService() company_service.EnvironmentServiceClient {
	return g.environmentService
}

func (g *grpcClients) MicroServiceResourceService() company_service.MicroserviceResourceClient {
	return g.microServiceResourceService
}

func (g *grpcClients) WebPageAppService() web_page_service.AppServiceClient {
	return g.webPageAppService
}

func (g *grpcClients) ServiceResource() company_service.MicroserviceResourceClient {
	return g.serviceResource
}

func (g *grpcClients) AppleIdService() auth_service.AppleIdLoginServiceClient {
	return g.appleIdService
}
