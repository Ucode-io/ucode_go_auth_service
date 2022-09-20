package config

import "time"

const (
	DatabaseQueryTimeLayout = `'YYYY-MM-DD"T"HH24:MI:SS"."MS"Z"TZ'`
	// DatabaseTimeLayout
	DatabaseTimeLayout string = time.RFC3339
	// AccessTokenExpiresInTime ...
	AccessTokenExpiresInTime time.Duration = 1 * 24 * 60 * time.Minute
	// RefreshTokenExpiresInTime ...
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
)
