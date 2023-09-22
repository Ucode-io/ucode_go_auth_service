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
	ServiceName string
	Environment string // debug, test, release
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

	DefaultOffset string
	DefaultLimit  string

	SecretKey string

	PasscodePool   string
	PasscodeLength int

	AuthServiceHost string
	AuthGRPCPort    string

	ObjectBuilderServiceHost string
	ObjectBuilderGRPCPort    string

	SmsServiceHost string
	SmsGRPCPort    string

	CompanyServiceHost string
	CompanyGRPCPort    string

	WebPageServiceHost string
	WebPageServicePort string

	PostgresObjectBuidlerServiceHost string
	PostgresObjectBuidlerServicePort string

	UcodeAppBaseUrl string
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
	// postgres://auth_service:IeX7ieso@95.217.155.57:30034/auth_service?sslmode=disable
	config.ServiceName = cast.ToString(getOrReturnDefaultValue("SERVICE_NAME", "ucode_go_auth_service"))
	config.Environment = cast.ToString(getOrReturnDefaultValue("ENVIRONMENT", DebugMode))
	config.Version = cast.ToString(getOrReturnDefaultValue("VERSION", "1.0"))

	config.HTTPPort = cast.ToString(getOrReturnDefaultValue("HTTP_PORT", ":9107"))
	config.HTTPScheme = cast.ToString(getOrReturnDefaultValue("HTTP_SCHEME", "http"))
	config.Email = cast.ToString(getOrReturnDefaultValue("EMAIL", "ucode.udevs.io@gmail.com"))
	config.Password = cast.ToString(getOrReturnDefaultValue("PASSWORD", "xkiaqodjfuielsug"))

	config.PostgresHost = cast.ToString(getOrReturnDefaultValue("POSTGRES_HOST", "65.109.239.69"))
	config.PostgresPort = cast.ToInt(getOrReturnDefaultValue("POSTGRES_PORT", 5432))
	config.PostgresUser = cast.ToString(getOrReturnDefaultValue("POSTGRES_USER", "auth_service"))
	config.PostgresPassword = cast.ToString(getOrReturnDefaultValue("POSTGRES_PASSWORD", "Iegfrte45eatr7ieso"))
	config.PostgresDatabase = cast.ToString(getOrReturnDefaultValue("POSTGRES_DATABASE", "auth_service"))

	config.PostgresMaxConnections = cast.ToInt32(getOrReturnDefaultValue("POSTGRES_MAX_CONNECTIONS", 30))

	config.DefaultOffset = cast.ToString(getOrReturnDefaultValue("DEFAULT_OFFSET", "0"))
	config.DefaultLimit = cast.ToString(getOrReturnDefaultValue("DEFAULT_LIMIT", "100"))

	config.SecretKey = cast.ToString(getOrReturnDefaultValue("SECRET_KEY", "Here$houldBe$ome$ecretKey"))

	config.PasscodePool = cast.ToString(getOrReturnDefaultValue("PASSCODE_POOL", "0123456789"))
	config.PasscodeLength = cast.ToInt(getOrReturnDefaultValue("PASSCODE_LENGTH", "6"))

	config.AuthServiceHost = cast.ToString(getOrReturnDefaultValue("AUTH_SERVICE_HOST", "localhost"))
	config.AuthGRPCPort = cast.ToString(getOrReturnDefaultValue("AUTH_GRPC_PORT", ":9103"))

	config.ObjectBuilderServiceHost = cast.ToString(getOrReturnDefaultValue("OBJECT_BUILDER_SERVICE_HOST", "localhost"))
	config.ObjectBuilderGRPCPort = cast.ToString(getOrReturnDefaultValue("OBJECT_BUILDER_GRPC_PORT", ":9102"))

	config.SmsServiceHost = cast.ToString(getOrReturnDefaultValue("SMS_SERVICE_HOST", "localhost"))
	config.SmsGRPCPort = cast.ToString(getOrReturnDefaultValue("SMS_GRPC_PORT", ":9105"))

	config.CompanyServiceHost = cast.ToString(getOrReturnDefaultValue("COMPANY_SERVICE_HOST", "localhost"))
	config.CompanyGRPCPort = cast.ToString(getOrReturnDefaultValue("COMPANY_GRPC_PORT", ":8092"))

	config.WebPageServiceHost = cast.ToString(getOrReturnDefaultValue("WEB_PAGE_SERVICE_HOST", "localhost"))
	config.WebPageServicePort = cast.ToString(getOrReturnDefaultValue("WEB_PAGE_GRPC_PORT", ":8098"))

	config.PostgresObjectBuidlerServiceHost = cast.ToString(getOrReturnDefaultValue("NODE_POSTGRES_SERVICE_HOST", "localhost"))
	config.PostgresObjectBuidlerServicePort = cast.ToString(getOrReturnDefaultValue("NODE_POSTGRES_SERVICE_PORT", ":9202"))
	config.UcodeAppBaseUrl = cast.ToString(getOrReturnDefaultValue("UCODE_APP_BASE_URL", "https://dev-app.ucode.run"))

	return config
}

func getOrReturnDefaultValue(key string, defaultValue interface{}) interface{} {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return defaultValue
}
