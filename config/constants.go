package config

import (
	"errors"
	"time"
)

const (
	DatabaseQueryTimeLayout = `'YYYY-MM-DD"T"HH24:MI:SS"."MS"Z"TZ'`
	// DatabaseTimeLayout
	DatabaseTimeLayout string = time.RFC3339
	// AccessTokenExpiresInTime ... 1 * 24 *
	AccessTokenExpiresInTime time.Duration = 1 * 1 * 60 * time.Minute
	// RefreshTokenExpiresInTime ... 30 * 24 * 60
	RefreshTokenExpiresInTime time.Duration = 30 * 24 * 60 * time.Minute

	// ProjectID
	ProjectID string = "f5955c82-f264-4655-aeb4-86fd1c642cb6"
	// CustomerClientPlatformID
	CustomerClientPlatformID string = "7d4a4c38-dd84-4902-b744-0488b80a4c02"
	// CustomerClientTypeID
	CustomerClientTypeID string = "5a3818a9-90f0-44e9-a053-3be0ba1e2c04"
	// CustomerDefaultRoleID
	CustomerDefaultRoleID string = "a1ca1301-4da9-424d-a9e2-578ae6dcde04"
	// AdminClientPlatformID
	AdminClientPlatformID string = "7d4a4c38-dd84-4902-b744-0488b80a4c01"
	// DeveloperClientTypeID
	DeveloperClientTypeID string = "5a3818a9-90f0-44e9-a053-3be0ba1e2c02"

	AdminClientTypeID    string = "142e9d0b-d9d3-4f71-bde1-5f1dbd70e83d"
	AdminClientName      string = "ADMIN"
	UcodeTestAdminDomain string = "test.admin.u-code.io"
	// UcodeDefaultProjectID string = "ucode_default_project_id"
	UcodeDefaultProjectID string = "39f1b0cc-8dc3-42df-b2bf-813310c007a4"

	ObjectBuilderService = "BUILDER_SERVICE"

	LOW_NODE_TYPE  string = "LOW"
	HIGH_NODE_TYPE string = "HIGH"
)

var (
	// these apis also manage by app's permission

	WithGoogle    = "google"
	Default       = "default"
	WithPhone     = "phone"
	WithApple     = "apple"
	WithEmail     = "email"
	RegisterTypes = map[string]int{
		"google":  1,
		"default": 1,
		"phone":   1,
		"apple":   1,
		"email":   1,
	}
	ErrUserNotFound = errors.New("user not found")

	LoginStrategyTypes = map[string]int{
		"EMAIL":       1,
		"PHONE":       1,
		"EMAIL_OTP":   1,
		"PHONE_OTP":   1,
		"LOGIN":       1,
		"LOGIN_PWD":   1,
		"GOOGLE_AUTH": 1,
		"APPLE_AUTH":  1,
	}

	ObjectBuilderTableSlugs = map[string]int{
		"field":               1,
		"view":                1,
		"table":               1,
		"relation":            1,
		"section":             1,
		"view_relation":       1,
		"html-template":       1,
		"variable":            1,
		"dashboard":           1,
		"panel":               1,
		"html-to-pdf":         1,
		"document":            1,
		"template-to-html":    1,
		"many-to-many":        1,
		"upload":              1,
		"upload-file":         1,
		"close-cashbox":       1,
		"open-cashbox":        1,
		"cashbox_transaction": 1,
		"query":               1,
		"event":               1,
		"event-log":           1,
		"permission-upsert":   1,
		"custom-event":        1,
		"excel":               1,
		"field-permission":    1,
		"function":            1,
		"invoke_function":     1,
	}
)

var (
	ErrUserAlradyMember = errors.New("user is already member")
)
