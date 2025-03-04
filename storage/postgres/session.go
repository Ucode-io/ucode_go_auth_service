package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/util"
	"ucode/ucode_go_auth_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

type sessionRepo struct {
	db *Pool
}

func NewSessionRepo(db *Pool) storage.SessionRepoI {
	return &sessionRepo{
		db: db,
	}
}

func (r *sessionRepo) Create(ctx context.Context, entity *pb.CreateSessionRequest) (pKey *pb.SessionPrimaryKey, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "session.Create")
	defer dbSpan.Finish()

	var envID *string
	if entity.EnvId != "" {
		envID = &entity.EnvId
	} else {
		envID = nil
	}

	countQuery := `
		SELECT COUNT(*) 
		FROM "session" 
		WHERE client_type_id = $1 
		  AND user_id_auth = $2 
		  AND env_id IS NOT DISTINCT FROM $3
	`
	var sessionCount int32
	err = r.db.QueryRow(ctx, countQuery, entity.ClientTypeId, entity.UserIdAuth, envID).Scan(&sessionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	if sessionCount >= entity.SessionLimit {
		sessionsToDelete := sessionCount - entity.SessionLimit + 1
		deleteQuery := `
			DELETE FROM "session"
			WHERE id IN (
				SELECT id
				FROM "session"
				WHERE client_type_id = $1 
				  AND user_id_auth = $2 
				  AND env_id IS NOT DISTINCT FROM $3
				ORDER BY created_at ASC
				LIMIT $4
			)`
		_, err = r.db.Exec(ctx, deleteQuery, entity.ClientTypeId, entity.UserIdAuth, envID, sessionsToDelete)
		if err != nil {
			return nil, fmt.Errorf("failed to delete sessions: %w", err)
		}
	}

	queryInitial := `INSERT INTO "session" (
		id,
		user_id,
		ip,
		data,
		user_id_auth,
		expires_at,
		client_id`

	queryValue := ` VALUES (
		$1, $2, $3, $4, $5, $6, $7`

	queryReturn := ` RETURNING id`

	random, err := uuid.NewRandom()
	if err != nil {
		return pKey, err
	}

	args := []any{
		random.String(),   // $1 - id
		entity.UserId,     // $2 - user_id
		entity.Ip,         // $3 - ip
		entity.Data,       // $4 - data
		entity.UserIdAuth, // $5 - user_id_auth
		entity.ExpiresAt,  // $6 - expires_at
		entity.ClientId,   // $7 - client_id
	}

	argIndex := 8

	if util.IsValidUUID(entity.ProjectId) {
		queryInitial += `, project_id`
		queryValue += `, $` + strconv.Itoa(argIndex)
		args = append(args, entity.ProjectId)
		argIndex++
	}

	if util.IsValidUUID(entity.EnvId) {
		queryInitial += `, env_id`
		queryValue += `, $` + strconv.Itoa(argIndex)
		args = append(args, entity.EnvId)
		argIndex++
	}

	if util.IsValidUUID(entity.ClientTypeId) {
		queryInitial += `, client_type_id`
		queryValue += `, $` + strconv.Itoa(argIndex)
		args = append(args, entity.ClientTypeId)
		argIndex++
	}

	if util.IsValidUUID(entity.RoleId) {
		queryInitial += `, role_id`
		queryValue += `, $` + strconv.Itoa(argIndex)
		args = append(args, entity.RoleId)
	}

	query := queryInitial + ")" + queryValue + ")" + queryReturn

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	pKey = &pb.SessionPrimaryKey{
		Id: random.String(),
	}

	return pKey, err
}

func (r *sessionRepo) GetByPK(ctx context.Context, pKey *pb.SessionPrimaryKey) (res *pb.Session, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "session.GetByPK")
	defer dbSpan.Finish()

	res = &pb.Session{}

	query := `SELECT
		id,
		COALESCE(project_id::text, ''),
		COALESCE(client_type_id::text, ''),
		user_id,
		COALESCE(role_id::text, ''),
		TEXT(ip) AS ip,
		data,
		COALESCE(user_id_auth::text, ''),
		COALESCE(is_changed, FALSE),
		COALESCE(env_id::text, ''),
		COALESCE(client_id, ''),
		COALESCE(TO_CHAR(expires_at, ` + config.DatabaseQueryTimeLayout + `)::TEXT, '') AS expires_at,
		COALESCE(TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `)::TEXT, '') AS created_at,
		COALESCE(TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `)::TEXT, '') AS updated_at
	FROM
		"session"
	WHERE
		id = $1`

	err = r.db.QueryRow(ctx, query, pKey.Id).Scan(
		&res.Id,
		&res.ProjectId,
		&res.ClientTypeId,
		&res.UserId,
		&res.RoleId,
		&res.Ip,
		&res.Data,
		&res.UserIdAuth,
		&res.IsChanged,
		&res.EnvId,
		&res.ClientId,
		&res.ExpiresAt,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return res, errors.Wrap(err, "error while getting session by id: "+err.Error())
	}

	return res, nil
}

func (r *sessionRepo) GetList(ctx context.Context, queryParam *pb.GetSessionListRequest) (*pb.GetSessionListResponse, error) {
	// Start tracing span
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "session.GetList")
	defer dbSpan.Finish()

	var (
		res   = &pb.GetSessionListResponse{}
		args  []any
		argID = 1
	)

	// Base query components
	baseQuery := `SELECT
		id,
		project_id,
		client_type_id,
		user_id,
		role_id,
		TEXT(ip) AS ip,
		data,
		is_changed,
		TO_CHAR(expires_at, ` + config.DatabaseQueryTimeLayout + `) AS expires_at,
		TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM "session"`

	filter := " WHERE 1=1"

	// Filtering conditions
	if len(queryParam.Search) > 0 {
		filter += " AND ip ILIKE $" + strconv.Itoa(argID)
		args = append(args, "%"+queryParam.Search+"%")
		argID++
	}

	if len(queryParam.UserId) > 0 {
		filter += " AND user_id = $" + strconv.Itoa(argID)
		args = append(args, queryParam.UserId)
		argID++
	}

	if len(queryParam.ClientTypeId) > 0 {
		filter += " AND client_type_id = $" + strconv.Itoa(argID)
		args = append(args, queryParam.ClientTypeId)
		argID++
	}

	// Count query
	countQuery := `SELECT count(1) FROM "session" ` + filter
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&res.Count)
	if err != nil {
		return res, err
	}

	// Pagination and ordering
	order := " ORDER BY created_at DESC"
	offset := " OFFSET 0"
	limit := " LIMIT 10"

	if queryParam.Offset > 0 {
		offset = " OFFSET $" + strconv.Itoa(argID)
		args = append(args, queryParam.Offset)
		argID++
	}

	if queryParam.Limit > 0 {
		limit = " LIMIT $" + strconv.Itoa(argID)
		args = append(args, queryParam.Limit)
		argID++
	}

	// Final query
	finalQuery := baseQuery + filter + order + offset + limit
	rows, err := r.db.Query(ctx, finalQuery, args...)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	// Process rows
	for rows.Next() {
		obj := &pb.Session{}
		err = rows.Scan(
			&obj.Id,
			&obj.ProjectId,
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

func (r *sessionRepo) Update(ctx context.Context, entity *pb.UpdateSessionRequest) (rowsAffected int64, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "session.Update")
	defer dbSpan.Finish()

	params := make(map[string]any)
	queryInitial := `UPDATE "session" SET
        ip = :ip,
		expires_at = :expires_at,
        is_changed = TRUE,
		updated_at = now()`

	filter := ` WHERE id = :id`
	params["id"] = entity.Id
	params["ip"] = entity.Ip
	params["expires_at"] = entity.ExpiresAt

	if util.IsValidUUID(entity.ProjectId) {
		params["project_id"] = entity.ProjectId
		queryInitial += `, project_id = :project_id`
	}

	if util.IsValidUUID(entity.EnvId) {
		params["env_id"] = entity.EnvId
		queryInitial += `, env_id = :env_id`
	}

	if util.IsValidUUID(entity.ClientTypeId) {
		params["client_type_id"] = entity.ClientTypeId
		queryInitial += `, client_type_id = :client_type_id`
	}

	if util.IsValidUUID(entity.RoleId) {
		params["role_id"] = entity.RoleId
		queryInitial += `, role_id = :role_id`
	}

	if entity.Data != "" {
		params["data"] = entity.Data
		queryInitial += `, data = :data`
	}

	query := queryInitial + filter

	cQuery, arr := helper.ReplaceQueryParams(query, params)
	result, err := r.db.Exec(ctx, cQuery, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *sessionRepo) Delete(ctx context.Context, pKey *pb.SessionPrimaryKey) (rowsAffected int64, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "session.Delete")
	defer dbSpan.Finish()

	query := `DELETE FROM "session" WHERE id = $1`

	result, err := r.db.Exec(ctx, query, pKey.Id)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *sessionRepo) DeleteExpiredUserSessions(ctx context.Context, userID string) (rowsAffected int64, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "session.DeleteExpiredUserSessions")
	defer dbSpan.Finish()

	query := `DELETE FROM "session" WHERE user_id = $1 AND expires_at < $2`

	result, err := r.db.Exec(ctx, query, userID, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *sessionRepo) GetSessionListByUserID(ctx context.Context, userID string) (res *pb.GetSessionListResponse, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "session.GetSessionListByUserID")
	defer dbSpan.Finish()

	res = &pb.GetSessionListResponse{}

	query := `SELECT
		id,
		coalesce(project_id::text, ''),
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
	WHERE user_id = $1
	ORDER BY created_at DESC`

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

func (r *sessionRepo) ExpireSessions(ctx context.Context, entity *pb.ExpireSessionsRequest) (err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "session.ExpireSessions")
	defer dbSpan.Finish()

	queryInitial := `DELETE FROM "session" WHERE id::varchar = ANY($1)`

	result, err := r.db.Exec(ctx, queryInitial, entity.SessionIds)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *sessionRepo) DeleteByParams(ctx context.Context, entity *pb.DeleteByParamsRequest) (err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "session.ExpireSessions")
	defer dbSpan.Finish()

	if entity.UserId == "" || entity.ProjectId == "" || entity.ClientTypeId == "" || entity.SessionId == "" {
		return errors.New("user_id, project_id, session_id and client_type_id are required")
	}

	query := `
		DELETE FROM session
		WHERE client_type_id = $1 AND user_id = $2 AND project_id = $3 AND id != $4
	`
	result, err := r.db.Exec(ctx, query, entity.ClientTypeId, entity.UserId, entity.ProjectId, entity.SessionId)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
