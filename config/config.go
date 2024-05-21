package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

const (
	// DebugMode indicates service mode is debug.
	DebugMode = "debug"
	// TestMode indicates service mode is test.
	TestMode = "test"
	// ReleaseMode indicates service mode is release.
	ReleaseMode = "release"
)

var CreadentialsForTest = map[string]map[string]string{
	DebugMode: {
		"projectId":             "62d6f9d4-dd9c-425b-84f6-cb90860967a8",
		"companyId":             "61bd72ca-f847-40f2-85b3-a337873862c3",
		"resourceEnvironmentId": "ecb08c73-3b52-42e9-970b-56be9b7c4e81",
		"clientTypeId":          "921743b1-9315-4eb9-b180-244bcbeb67cb",
		"roleId":                "3306fd21-ee1a-4c68-8843-6d0699b6f9ce",
	},
	TestMode: {
		"projectId": "",
		"companyId": "",
	},
	ReleaseMode: {
		"projectId": "",
		"companyId": "",
	},
}

type Config struct {
	DefaultOffset string
	DefaultLimit  string

	SecretKey string

	PasscodePool   string
	PasscodeLength int

	ObjectBuilderServiceHost string
	ObjectBuilderGRPCPort    string

	HighObjectBuilderServiceHost string
	HighObjectBuilderGRPCPort    string

	SmsServiceHost string
	SmsGRPCPort    string

	WebPageServiceHost string
	WebPageServicePort string

	PostgresObjectBuidlerServiceHost string
	PostgresObjectBuidlerServicePort string

	GoObjectBuilderServiceHost string
	GoObjectBuilderServicePort string

	UcodeAppBaseUrl string
}

type BaseConfig struct {
	ServiceName string
	Environment string
	Version     string

	HTTPPort   string
	HTTPScheme string

	Email    string
	Password string

	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDatabase string

	PostgresMaxConnections int32
	DefaultOffset          string
	DefaultLimit           string

	SmsServiceHost string
	SmsGRPCPort    string

	SecretKey string

	AuthServiceHost string
	AuthGRPCPort    string

	UcodeNamespace string

	CompanyServiceHost string
	CompanyGRPCPort    string
}

func BaseLoad() BaseConfig {

	if err := godotenv.Load("/app/.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			fmt.Println("No .env file found")
		}
		fmt.Println("No .env file found")
	}

	config := BaseConfig{}

	config.DefaultOffset = cast.ToString(getOrReturnDefaultValue("DEFAULT_OFFSET", "0"))
	config.DefaultLimit = cast.ToString(getOrReturnDefaultValue("DEFAULT_LIMIT", "100"))

	config.ServiceName = cast.ToString(getOrReturnDefaultValue("SERVICE_NAME", "auth_service"))
	config.Environment = cast.ToString(getOrReturnDefaultValue("ENVIRONMENT", DebugMode))
	config.Version = cast.ToString(getOrReturnDefaultValue("VERSION", "1.0"))

	config.HTTPPort = cast.ToString(getOrReturnDefaultValue("HTTP_PORT", ""))
	config.HTTPScheme = cast.ToString(getOrReturnDefaultValue("HTTP_SCHEME", ""))
	config.Email = cast.ToString(getOrReturnDefaultValue("EMAIL", ""))
	config.Password = cast.ToString(getOrReturnDefaultValue("PASSWORD", ""))

	config.PostgresHost = cast.ToString(getOrReturnDefaultValue("POSTGRES_HOST", ""))
	config.PostgresPort = cast.ToInt(getOrReturnDefaultValue("POSTGRES_PORT", 0))
	config.PostgresUser = cast.ToString(getOrReturnDefaultValue("POSTGRES_USER", ""))
	config.PostgresPassword = cast.ToString(getOrReturnDefaultValue("POSTGRES_PASSWORD", ""))
	config.PostgresDatabase = cast.ToString(getOrReturnDefaultValue("POSTGRES_DATABASE", ""))
	config.PostgresMaxConnections = cast.ToInt32(getOrReturnDefaultValue("POSTGRES_MAX_CONNECTIONS", 30))

	config.AuthServiceHost = cast.ToString(getOrReturnDefaultValue("AUTH_SERVICE_HOST", ""))
	config.AuthGRPCPort = cast.ToString(getOrReturnDefaultValue("AUTH_GRPC_PORT", ""))

	config.UcodeNamespace = "u-code"

	config.CompanyServiceHost = cast.ToString(getOrReturnDefaultValue("COMPANY_SERVICE_HOST", ""))
	config.CompanyGRPCPort = cast.ToString(getOrReturnDefaultValue("COMPANY_GRPC_PORT", ""))

	config.SmsServiceHost = cast.ToString(getOrReturnDefaultValue("SMS_SERVICE_HOST", ""))
	config.SmsGRPCPort = cast.ToString(getOrReturnDefaultValue("SMS_GRPC_PORT", ""))

	config.SecretKey = cast.ToString(getOrReturnDefaultValue("SECRET_KEY", "snZV9XNmvf"))

	return config
}

// Load ...
func Load() Config {
	if err := godotenv.Load("/app/.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			fmt.Println("No .env file found")
		}
		fmt.Println("No .env file found")
	}

	config := Config{}

	config.DefaultOffset = cast.ToString(getOrReturnDefaultValue("DEFAULT_OFFSET", "0"))
	config.DefaultLimit = cast.ToString(getOrReturnDefaultValue("DEFAULT_LIMIT", "100"))

	config.SecretKey = cast.ToString(getOrReturnDefaultValue("SECRET_KEY", ""))

	config.PasscodePool = cast.ToString(getOrReturnDefaultValue("PASSCODE_POOL", ""))
	config.PasscodeLength = cast.ToInt(getOrReturnDefaultValue("PASSCODE_LENGTH", ""))

	config.ObjectBuilderServiceHost = cast.ToString(getOrReturnDefaultValue("OBJECT_BUILDER_SERVICE_LOW_HOST", ""))
	config.ObjectBuilderGRPCPort = cast.ToString(getOrReturnDefaultValue("OBJECT_BUILDER_LOW_GRPC_PORT", ""))

	config.HighObjectBuilderServiceHost = cast.ToString(getOrReturnDefaultValue("OBJECT_BUILDER_SERVICE_HIGHT_HOST", ""))
	config.HighObjectBuilderGRPCPort = cast.ToString(getOrReturnDefaultValue("OBJECT_BUILDER_HIGH_GRPC_PORT", ""))

	config.SmsServiceHost = cast.ToString(getOrReturnDefaultValue("SMS_SERVICE_HOST", ""))
	config.SmsGRPCPort = cast.ToString(getOrReturnDefaultValue("SMS_GRPC_PORT", ""))

	config.WebPageServiceHost = cast.ToString(getOrReturnDefaultValue("WEB_PAGE_SERVICE_HOST", ""))
	config.WebPageServicePort = cast.ToString(getOrReturnDefaultValue("WEB_PAGE_GRPC_PORT", ""))

	config.PostgresObjectBuidlerServiceHost = cast.ToString(getOrReturnDefaultValue("NODE_POSTGRES_SERVICE_HOST", ""))
	config.PostgresObjectBuidlerServicePort = cast.ToString(getOrReturnDefaultValue("NODE_POSTGRES_SERVICE_PORT", ""))

	config.GoObjectBuilderServiceHost = cast.ToString(getOrReturnDefaultValue("GO_OBJECT_BUILDER_SERVICE_HOST", ""))
	config.GoObjectBuilderServicePort = cast.ToString(getOrReturnDefaultValue("GO_OBJECT_BUILDER_GRPC_PORT", ""))

	config.PostgresObjectBuidlerServiceHost = cast.ToString(getOrReturnDefaultValue("NODE_POSTGRES_SERVICE_HOST", ""))
	config.PostgresObjectBuidlerServicePort = cast.ToString(getOrReturnDefaultValue("NODE_POSTGRES_SERVICE_PORT", ""))
	config.UcodeAppBaseUrl = cast.ToString(getOrReturnDefaultValue("UCODE_APP_BASE_URL", ""))

	return config
}

func getOrReturnDefaultValue(key string, defaultValue interface{}) interface{} {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return defaultValue
}
