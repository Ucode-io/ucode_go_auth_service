package config

import (
	"errors"
	"time"

	"github.com/golanguzb70/ratelimiter"
)

const (
	// Default Configs
	DatabaseQueryTimeLayout   string        = `'YYYY-MM-DD"T"HH24:MI:SS"."MS"Z"TZ'`
	DatabaseTimeLayout        string        = time.RFC3339
	AccessTokenExpiresInTime  time.Duration = 1 * 60 * 24 * time.Minute
	RefreshTokenExpiresInTime time.Duration = 30 * 24 * 60 * time.Minute

	ProjectID             string = "f5955c82-f264-4655-aeb4-86fd1c642cb6"
	AdminClientPlatformID string = "7d4a4c38-dd84-4902-b744-0488b80a4c01"
	AdminClientName       string = "ADMIN"
	OpenFaaSPlatformID    string = "7d4a4c38-dd84-4902-b744-0488b80a4c04"

	DefaultOtp string = "208071"

	// Service Configs
	LOW_NODE_TYPE        string = "LOW"
	HIGH_NODE_TYPE       string = "HIGH"
	ENTER_PRICE_TYPE     string = "ENTER_PRICE"
	ObjectBuilderService string = "BUILDER_SERVICE"

	// Login Strategy
	WithGoogle string = "google"
	Default    string = "default"
	WithPhone  string = "phone"
	WithApple  string = "apple"
	WithEmail  string = "email"
	WithLogin  string = "login"

	K8SNamespace string = "cp-region-type-id"
	LanguageId   string = "e2d68f08-8587-4136-8cd4-c26bf1b9cda1"
	NativeName   string = "English"
	ShortName    string = "en"

	// Errors
	UserProjectIdConstraint   string = "user_project_idx_unique"
	DuplicateUserProjectError string = "user with this client_type already exists in the project"

	UserStatusBlocked  string = "BLOCKED"
	UserStatusInactive string = "INACTIVE"
	UserStatusActive   string = "ACTIVE"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailRequired    = errors.New("email required for register company")
	ErrUserAlradyMember = errors.New("user is already member")
	ErrInactiveClientId = errors.New("client id is an inactive")

	RegisterTypes = map[string]bool{
		"google":  true,
		"default": true,
		"phone":   true,
		"apple":   true,
		"email":   true,
	}

	ObjectBuilderTableSlugs = map[string]bool{
		"field":               true,
		"view":                true,
		"table":               true,
		"relation":            true,
		"section":             true,
		"view_relation":       true,
		"html-template":       true,
		"variable":            true,
		"dashboard":           true,
		"panel":               true,
		"html-to-pdf":         true,
		"document":            true,
		"template-to-html":    true,
		"many-to-many":        true,
		"upload":              true,
		"upload-file":         true,
		"close-cashbox":       true,
		"open-cashbox":        true,
		"cashbox_transaction": true,
		"query":               true,
		"event":               true,
		"event-log":           true,
		"permission-upsert":   true,
		"custom-event":        true,
		"excel":               true,
		"field-permission":    true,
		"function":            true,
		"invoke_function":     true,
	}

	HashTypes = map[string]int{
		"argon":  1,
		"bcrypt": 2,
	}

	Path = map[string]bool{
		"object":      true,
		"object-slim": true,
		// "items":       true,
	}

	RateLimitCfg = []*ratelimiter.LeakyBucket{
		{
			Method:         "POST",
			Path:           "/v2/send-code",
			RequestLimit:   5,
			Interval:       "minute",
			Type:           "body",
			KeyField:       "recipient",
			AllowOnFailure: true,
			NotAllowMsg:    "send-code request limit exceeded",
			NotAllowCode:   "TOO_MANY_REQUESTS",
		},
	}
)
