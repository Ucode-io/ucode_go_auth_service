package client

import (
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/new_object_builder_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"

	"ucode/ucode_go_auth_service/genproto/sms_service"
	"ucode/ucode_go_auth_service/genproto/web_page_service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SharedServiceManagerI interface {
	ObjectBuilderService() object_builder_service.ObjectBuilderServiceClient
	LoginService() object_builder_service.LoginServiceClient
	GoLoginService() new_object_builder_service.LoginServiceClient
	BuilderPermissionService() object_builder_service.PermissionServiceClient
	PostgresObjectBuilderService() object_builder_service.ObjectBuilderServiceClient
	PostgresLoginService() object_builder_service.LoginServiceClient
	PostgresBuilderPermissionService() object_builder_service.PermissionServiceClient
	VersionHistoryService() object_builder_service.VersionHistoryServiceClient

	GoObjectBuilderService() new_object_builder_service.ObjectBuilderServiceClient
	GoItemService() new_object_builder_service.ItemsServiceClient
	GoObjectBuilderPermissionService() new_object_builder_service.PermissionServiceClient
	GoObjectBuilderLoginService() new_object_builder_service.LoginServiceClient

	HighObjectBuilderService() object_builder_service.ObjectBuilderServiceClient
	HighLoginService() object_builder_service.LoginServiceClient
	HighBuilderPermissionService() object_builder_service.PermissionServiceClient

	GetObjectBuilderServiceByType(nodeType string) object_builder_service.ObjectBuilderServiceClient
	GetLoginServiceByType(nodeType string) object_builder_service.LoginServiceClient
	GetBuilderPermissionServiceByType(nodeType string) object_builder_service.PermissionServiceClient

	SmsService() sms_service.SmsServiceClient
}

type sharedGrpcClients struct {
	objectBuilderService             object_builder_service.ObjectBuilderServiceClient
	loginService                     object_builder_service.LoginServiceClient
	goLoginService                   new_object_builder_service.LoginServiceClient
	builderPermissionService         object_builder_service.PermissionServiceClient
	postgresObjectBuilderService     object_builder_service.ObjectBuilderServiceClient
	postgresLoginService             object_builder_service.LoginServiceClient
	postgresBuilderPermissionService object_builder_service.PermissionServiceClient
	versionHisotryService            object_builder_service.VersionHistoryServiceClient

	goObjectBuilderService           new_object_builder_service.ObjectBuilderServiceClient
	goItemsService                   new_object_builder_service.ItemsServiceClient
	goObjectBuilderPermissionService new_object_builder_service.PermissionServiceClient
	goItemService                    new_object_builder_service.ItemsServiceClient
	goObjectBuilderLoginService      new_object_builder_service.LoginServiceClient

	highObjectBuilderService     object_builder_service.ObjectBuilderServiceClient
	highLoginService             object_builder_service.LoginServiceClient
	highBuilderPermissionService object_builder_service.PermissionServiceClient

	webPageAppService web_page_service.AppServiceClient
	smsService        sms_service.SmsServiceClient
}

func NewSharedGrpcClients(cfg config.Config) (SharedServiceManagerI, error) {

	connObjectBuilderService, err := grpc.Dial(
		cfg.ObjectBuilderServiceHost+cfg.ObjectBuilderGRPCPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(52428800), grpc.MaxCallSendMsgSize(52428800)),
	)
	if err != nil {
		return nil, err
	}
	connHighObjectBuilderService, err := grpc.Dial(
		cfg.HighObjectBuilderServiceHost+cfg.HighObjectBuilderGRPCPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(52428800), grpc.MaxCallSendMsgSize(52428800)),
	)
	if err != nil {
		return nil, err
	}
	connSmsService, err := grpc.Dial(
		cfg.SmsServiceHost+cfg.SmsGRPCPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	connWebPageService, err := grpc.Dial(
		cfg.WebPageServiceHost+cfg.WebPageServicePort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	connPostgresObjectBuilderService, err := grpc.Dial(
		cfg.PostgresObjectBuidlerServiceHost+cfg.PostgresObjectBuidlerServicePort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	connGoObjectBuilderService, err := grpc.Dial(
		cfg.GoObjectBuilderServiceHost+cfg.GoObjectBuilderServicePort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &sharedGrpcClients{
		objectBuilderService:             object_builder_service.NewObjectBuilderServiceClient(connObjectBuilderService),
		loginService:                     object_builder_service.NewLoginServiceClient(connObjectBuilderService),
		goLoginService:                   new_object_builder_service.NewLoginServiceClient(connGoObjectBuilderService),
		builderPermissionService:         object_builder_service.NewPermissionServiceClient(connObjectBuilderService),
		postgresLoginService:             object_builder_service.NewLoginServiceClient(connPostgresObjectBuilderService),
		postgresObjectBuilderService:     object_builder_service.NewObjectBuilderServiceClient(connPostgresObjectBuilderService),
		postgresBuilderPermissionService: object_builder_service.NewPermissionServiceClient(connPostgresObjectBuilderService),
		versionHisotryService:            object_builder_service.NewVersionHistoryServiceClient(connObjectBuilderService),

		highObjectBuilderService:     object_builder_service.NewObjectBuilderServiceClient(connHighObjectBuilderService),
		highLoginService:             object_builder_service.NewLoginServiceClient(connHighObjectBuilderService),
		highBuilderPermissionService: object_builder_service.NewPermissionServiceClient(connHighObjectBuilderService),

		webPageAppService: web_page_service.NewAppServiceClient(connWebPageService),
		smsService:        sms_service.NewSmsServiceClient(connSmsService),

		goObjectBuilderService:           new_object_builder_service.NewObjectBuilderServiceClient(connGoObjectBuilderService),
		goItemsService:                   new_object_builder_service.NewItemsServiceClient(connGoObjectBuilderService),
		goObjectBuilderPermissionService: new_object_builder_service.NewPermissionServiceClient(connGoObjectBuilderService),
		goItemService:                    new_object_builder_service.NewItemsServiceClient(connGoObjectBuilderService),
		goObjectBuilderLoginService:      new_object_builder_service.NewLoginServiceClient(connGoObjectBuilderService),
	}, nil
}

func (g *sharedGrpcClients) GetObjectBuilderServiceByType(nodeType string) object_builder_service.ObjectBuilderServiceClient {
	switch nodeType {
	case config.LOW_NODE_TYPE:
		return g.objectBuilderService
	case config.HIGH_NODE_TYPE:
		return g.highObjectBuilderService
	}

	return g.objectBuilderService
}

func (g *sharedGrpcClients) GetLoginServiceByType(nodeType string) object_builder_service.LoginServiceClient {
	switch nodeType {
	case config.LOW_NODE_TYPE:
		return g.loginService
	case config.HIGH_NODE_TYPE:
		return g.highLoginService
	}

	return g.loginService
}

func (g *sharedGrpcClients) GetBuilderPermissionServiceByType(nodeType string) object_builder_service.PermissionServiceClient {
	switch nodeType {
	case config.LOW_NODE_TYPE:
		return g.builderPermissionService
	case config.HIGH_NODE_TYPE:
		return g.highBuilderPermissionService
	}

	return g.builderPermissionService
}

func (g *sharedGrpcClients) BuilderPermissionService() object_builder_service.PermissionServiceClient {
	return g.builderPermissionService
}

func (g *sharedGrpcClients) PostgresObjectBuilderService() object_builder_service.ObjectBuilderServiceClient {
	return g.postgresObjectBuilderService
}

func (g *sharedGrpcClients) PostgresLoginService() object_builder_service.LoginServiceClient {
	return g.postgresLoginService
}

func (g *sharedGrpcClients) ObjectBuilderService() object_builder_service.ObjectBuilderServiceClient {
	return g.objectBuilderService
}

func (g *sharedGrpcClients) SmsService() sms_service.SmsServiceClient {
	return g.smsService
}

func (g *sharedGrpcClients) PostgresBuilderPermissionService() object_builder_service.PermissionServiceClient {
	return g.postgresBuilderPermissionService
}

func (g *sharedGrpcClients) HighObjectBuilderService() object_builder_service.ObjectBuilderServiceClient {
	return g.highObjectBuilderService
}

func (g *sharedGrpcClients) HighLoginService() object_builder_service.LoginServiceClient {
	return g.highLoginService
}

func (g *sharedGrpcClients) HighBuilderPermissionService() object_builder_service.PermissionServiceClient {
	return g.highBuilderPermissionService
}

func (g *sharedGrpcClients) WebPageAppService() web_page_service.AppServiceClient {
	return g.webPageAppService
}

func (g *sharedGrpcClients) LoginService() object_builder_service.LoginServiceClient {
	return g.loginService
}

func (g *sharedGrpcClients) VersionHistoryService() object_builder_service.VersionHistoryServiceClient {
	return g.versionHisotryService
}

func (g *sharedGrpcClients) GoObjectBuilderService() new_object_builder_service.ObjectBuilderServiceClient {
	return g.goObjectBuilderService
}

func (g *sharedGrpcClients) GoLoginService() new_object_builder_service.LoginServiceClient {
	return g.goLoginService
}

func (g *sharedGrpcClients) GoObjectBuilderPermissionService() new_object_builder_service.PermissionServiceClient {
	return g.goObjectBuilderPermissionService
}

func (g *sharedGrpcClients) GoItemService() new_object_builder_service.ItemsServiceClient {
	return g.goItemService
}

func (g *sharedGrpcClients) GoObjectBuilderLoginService() new_object_builder_service.LoginServiceClient {
	return g.goLoginService
}
