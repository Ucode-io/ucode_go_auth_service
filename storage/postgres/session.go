package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type sessionRepo struct {
	db *pgxpool.Pool
}

func NewSessionRepo(db *pgxpool.Pool) storage.SessionRepoI {
	return &sessionRepo{
		db: db,
	}
}

func (r *sessionRepo) Create(ctx context.Context, entity *pb.CreateSessionRequest) (pKey *pb.SessionPrimaryKey, err error) {
	log.Printf("--->STRG: CreateSessionRequest: %+v", entity)

	query := `INSERT INTO "session" (
		id,
		project_id,
		client_platform_id,
		client_type_id,
		user_id,
		role_id,
		ip,
		data,
		expires_at
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9
	)`

	uuid, err := uuid.NewRandom()
	if err != nil {
		return pKey, err
	}

	_, err = r.db.Exec(ctx, query,
		uuid.String(),
		entity.ProjectId,
		entity.ClientPlatformId,
		entity.ClientTypeId,
		entity.UserId,
		entity.RoleId,
		entity.Ip,
		entity.Data,
		entity.ExpiresAt,
	)

	pKey = &pb.SessionPrimaryKey{
		Id: uuid.String(),
	}

	return pKey, err
}

func (r *sessionRepo) CreateSuperAdmin(ctx context.Context, entity *pb.CreateSessionRequest) (pKey *pb.SessionPrimaryKey, err error) {
	query := `INSERT INTO "session" (
		id,
		user_id,
		ip,
		data,
		expires_at,
		project_id
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6
	)`

	uuid, err := uuid.NewRandom()
	if err != nil {
		return pKey, err
	}

	_, err = r.db.Exec(ctx, query,
		uuid.String(),
		entity.GetUserId(),
		entity.GetIp(),
		entity.GetData(),
		entity.GetExpiresAt(),
		entity.GetProjectId(),
	)

	pKey = &pb.SessionPrimaryKey{
		Id: uuid.String(),
	}

	return pKey, err
}

func (r *sessionRepo) GetByPK(ctx context.Context, pKey *pb.SessionPrimaryKey) (res *pb.Session, err error) {
	res = &pb.Session{}
	query := `SELECT
		id,
		coalesce(project_id::text, ''),
		coalesce(client_platform_id::text, ''),
		coalesce(client_type_id::text, ''),
		user_id,
		coalesce(role_id::text, ''),
		TEXT(ip) AS ip,
		data,
		is_changed,
		TO_CHAR(expires_at, ` + config.DatabaseQueryTimeLayout + `) AS expires_at,
		TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM
		"session"
	WHERE
		id = $1`

	var (
		expiresAt sql.NullString
		createdAt sql.NullString
		updatedAt sql.NullString
		userID    sql.NullString
		isChanged sql.NullBool
	)

	err = r.db.QueryRow(ctx, query, pKey.Id).Scan(
		&res.Id,
		&res.ProjectId,
		&res.ClientPlatformId,
		&res.ClientTypeId,
		&userID,
		&res.RoleId,
		&res.Ip,
		&res.Data,
		&isChanged,
		// &res.ExpiresAt,
		&expiresAt,
		// &res.CreatedAt,
		&createdAt,
		// &res.UpdatedAt,
		&updatedAt,
	)
	if err != nil {
		return res, errors.Wrap(err, "error while getting session by id: "+err.Error())
	}

	if userID.Valid {
		res.UserId = userID.String
	}

	if expiresAt.Valid {
		res.ExpiresAt = expiresAt.String
	}

	if createdAt.Valid {
		res.CreatedAt = createdAt.String
	}

	if updatedAt.Valid {
		res.UpdatedAt = updatedAt.String
	}
	if isChanged.Valid {
		res.IsChanged = isChanged.Bool
	}

	return res, nil
}

func (r *sessionRepo) GetList(ctx context.Context, queryParam *pb.GetSessionListRequest) (res *pb.GetSessionListResponse, err error) {
	res = &pb.GetSessionListResponse{}
	params := make(map[string]interface{})
	var arr []interface{}
	query := `SELECT
		id,
		project_id,
		client_platform_id,
		client_type_id,
		user_id,
		role_id,
		TEXT(ip) AS ip,
		data,
		is_changed,
		TO_CHAR(expires_at, ` + config.DatabaseQueryTimeLayout + `) AS expires_at,
		TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM
		"session"`
	filter := " WHERE 1=1"
	order := " ORDER BY created_at"
	arrangement := " DESC"
	offset := " OFFSET 0"
	limit := " LIMIT 10"

	if len(queryParam.Search) > 0 {
		params["search"] = queryParam.Search
		filter += " AND ((ip) ILIKE ('%' || :search || '%'))"
	}

	if queryParam.Offset > 0 {
		params["offset"] = queryParam.Offset
		offset = " OFFSET :offset"
	}

	if queryParam.Limit > 0 {
		params["limit"] = queryParam.Limit
		limit = " LIMIT :limit"
	}

	cQ := `SELECT count(1) FROM "session"` + filter
	cQ, arr = helper.ReplaceQueryParams(cQ, params)
	err = r.db.QueryRow(ctx, cQ, arr...).Scan(
		&res.Count,
	)
	if err != nil {
		return res, err
	}

	q := query + filter + order + arrangement + offset + limit
	q, arr = helper.ReplaceQueryParams(q, params)
	rows, err := r.db.Query(ctx, q, arr...)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		obj := &pb.Session{}
		err = rows.Scan(
			&obj.Id,
			&obj.ProjectId,
			&obj.ClientPlatformId,
			&obj.ClientTypeId,
			&obj.UserId,
			&obj.RoleId,
			&obj.Ip,
			&obj.Data,
			&obj.ExpiresAt,
			&obj.CreatedAt,
			&obj.UpdatedAt,
			obj.IsChanged,
		)

		if err != nil {
			return res, err
		}

		res.Sessions = append(res.Sessions, obj)
	}

	return res, nil
}

func (r *sessionRepo) Update(ctx context.Context, entity *pb.UpdateSessionRequest) (rowsAffected int64, err error) {
	query := `UPDATE "session" SET
		project_id = :project_id,
		client_platform_id = :client_platform_id,
		client_type_id = :client_type_id,
		user_id = :user_id,
		role_id = :role_id,
		ip = :ip,
		data = :data,
		is_changed = :is_changed,
		expires_at = :expires_at,
		updated_at = now()
	WHERE
		id = :id`

	params := map[string]interface{}{
		"id":                 entity.Id,
		"project_id":         entity.ProjectId,
		"client_platform_id": entity.ClientPlatformId,
		"client_type_id":     entity.ClientTypeId,
		"user_id":            entity.UserId,
		"role_id":            entity.RoleId,
		"ip":                 entity.Ip,
		"data":               entity.Data,
		"is_changed":         entity.IsChanged,
		"expires_at":         entity.ExpiresAt,
	}

	q, arr := helper.ReplaceQueryParams(query, params)
	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *sessionRepo) Delete(ctx context.Context, pKey *pb.SessionPrimaryKey) (rowsAffected int64, err error) {
	query := `DELETE FROM "session" WHERE id = $1`

	result, err := r.db.Exec(ctx, query, pKey.Id)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *sessionRepo) DeleteExpiredUserSessions(ctx context.Context, userID string) (rowsAffected int64, err error) {
	log.Printf("---STRG->DeleteExpiredUserSessions---> %s", userID)
	query := `DELETE FROM "session" WHERE user_id = $1 AND expires_at < $2`

	result, err := r.db.Exec(ctx, query, userID, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *sessionRepo) DeleteExpiredIntegrationSessions(ctx context.Context, integrationId string) (rowsAffected int64, err error) {
	query := `DELETE FROM "session" WHERE integration_id = $1 AND expires_at < $2`

	result, err := r.db.Exec(ctx, query, integrationId, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *sessionRepo) GetSessionListByUserID(ctx context.Context, userID string) (res *pb.GetSessionListResponse, err error) {
	res = &pb.GetSessionListResponse{}

	query := `SELECT
		id,
		coalesce(project_id::text, ''),
		coalesce(client_platform_id::text, ''),
		coalesce(client_type_id::text, ''),
		user_id,
		coalesce(role_id::text, ''),
		TEXT(ip) AS ip,
		data,
		is_changed,
		TO_CHAR(expires_at, ` + config.DatabaseQueryTimeLayout + `) AS expires_at,
		TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM
		"session"
	WHERE user_id = $1`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		obj := &pb.Session{}
		err = rows.Scan(
			&obj.Id,
			&obj.ProjectId,
			&obj.ClientPlatformId,
			&obj.ClientTypeId,
			&obj.UserId,
			&obj.RoleId,
			&obj.Ip,
			&obj.Data,
			&obj.IsChanged,
			&obj.ExpiresAt,
			&obj.CreatedAt,
			&obj.UpdatedAt,
		)

		if err != nil {
			return res, err
		}

		res.Sessions = append(res.Sessions, obj)
	}

	return res, nil
}
func (r *sessionRepo) GetSessionListByIntegrationID(ctx context.Context, integrationId string) (res *pb.GetSessionListResponse, err error) {
	res = &pb.GetSessionListResponse{}

	query := `SELECT
		id,
		project_id,
		client_platform_id,
		client_type_id,
		integration_id,
		role_id,
		TEXT(ip) AS ip,
		data,
		is_changed,
		TO_CHAR(expires_at, ` + config.DatabaseQueryTimeLayout + `) AS expires_at,
		TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM
		"session"
	WHERE integration_id = $1`

	rows, err := r.db.Query(ctx, query, integrationId)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		obj := &pb.Session{}
		err = rows.Scan(
			&obj.Id,
			&obj.ProjectId,
			&obj.ClientPlatformId,
			&obj.ClientTypeId,
			&obj.IntegrationId,
			&obj.RoleId,
			&obj.Ip,
			&obj.Data,
			&obj.IsChanged,
			&obj.ExpiresAt,
			&obj.CreatedAt,
			&obj.UpdatedAt,
		)

		if err != nil {
			return res, err
		}

		res.Sessions = append(res.Sessions, obj)
	}

	return res, nil
}

func (r *sessionRepo) UpdateByRoleId(ctx context.Context, entity *pb.UpdateSessionByRoleIdRequest) (rowsAffected int64, err error) {
	query := `UPDATE "session" SET
		is_changed = $2
	WHERE
		role_id = $1`

	result, err := r.db.Exec(ctx, query, entity.RoleId, entity.IsChanged)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *sessionRepo) UpdateBySessionId(ctx context.Context, entity *pb.UpdateSessionBySessionIdRequest) (rowsAffected int64, err error) {
	query := `UPDATE "session" SET
		is_changed = $2
	WHERE
		id = $1`

	result, err := r.db.Exec(ctx, query, entity.Id, entity.IsChanged)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}
