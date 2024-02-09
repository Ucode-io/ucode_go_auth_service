package postgres

import (
	"context"
	"database/sql"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/util"
	"ucode/ucode_go_auth_service/storage"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type apiKeysRepo struct {
	db *pgxpool.Pool
}

func NewApiKeysRepo(db *pgxpool.Pool) storage.ApiKeysRepoI {
	return &apiKeysRepo{
		db: db,
	}
}

func (r *apiKeysRepo) Create(ctx context.Context, req *pb.CreateReq, appSecret, appId, id string) (*pb.CreateRes, error) {
	var (
		res pb.CreateRes

		createdAt sql.NullString
		updatedAt sql.NullString
	)

	query := `
		INSERT INTO api_keys(
			id, 
			name, 
			app_id, 
			app_secret, 
			role_id, 
			environment_id, 
			project_id, 
			client_type_id,
			created_at, 
			updated_at
		)
		VALUES (
			$1, 
			$2, 
			$3, 
			$4, 
			$5, 
			$6, 
			$7, 
			$8,
			now(), 
			now()) 
		RETURNING id, status, name, app_id, app_secret, role_id, created_at, updated_at, environment_id, project_id, client_type_id, rps_limit, monthly_request_limit`

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		req.GetName(),
		appId,
		appSecret,
		req.GetRoleId(),
		req.GetEnvironmentId(),
		req.GetProjectId(),
		req.GetClientTypeId(),
	).Scan(&res.Id, &res.Status, &res.Name, &res.AppId, &res.AppSecret, &res.RoleId, &createdAt, &updatedAt, &res.EnvironmentId, &res.ProjectId, &res.ClientTypeId, &res.RpsLimit, &res.MonthlyRequestLimit)

	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		res.CreatedAt = createdAt.String
	}

	if updatedAt.Valid {
		res.UpdatedAt = updatedAt.String
	}

	return &res, nil
}
func (r *apiKeysRepo) GetList(ctx context.Context, req *pb.GetListReq) (*pb.GetListRes, error) {
	var (
		res = pb.GetListRes{Count: 0}
	)

	query := `SELECT
				id,
  				status,
  				name,
  				app_id,
				app_secret,
  				role_id,
  				created_at,
  				updated_at,
  				environment_id,
				project_id,
				client_type_id,
				rps_limit,
				monthly_request_limit,
				(SELECT count(1) FROM api_key_usage where api_key=app_id AND created_at >= date_trunc('month', current_date))
			FROM
			    api_keys`

	filter := ` WHERE project_id = :project_id`
	params := make(map[string]interface{})
	offset := " OFFSET 0"
	limit := " LIMIT 10"
	order := " ORDER BY created_at"
	arrangement := " DESC"
	params["project_id"] = req.GetProjectId()

	if req.Offset > 0 {
		params["offset"] = req.Offset
		offset = " OFFSET :offset"
	}

	if req.Limit > 0 {
		params["limit"] = req.Limit
		limit = " LIMIT :limit"
	}

	if len(req.Search) > 0 {
		params["search"] = req.Search
		filter += " AND (name ILIKE '%' || :search || '%')"
	}

	if util.IsValidUUID(req.EnvironmentId) {
		filter += ` AND environment_id = :environment_id`
		params["environment_id"] = req.GetEnvironmentId()
	}

	if util.IsValidUUID(req.ClientTypeId) {
		filter += ` AND client_type_id = :client_type_id`
		params["client_type_id"] = req.ClientTypeId
	}

	if util.IsValidUUID(req.RoleId) {
		filter += ` AND role_id = :role_id`
		params["role_id"] = req.RoleId
	}

	countQuery := `SELECT count(*) from api_keys` + filter
	countStmt, countArgs := helper.ReplaceQueryParams(countQuery, params)

	err := r.db.QueryRow(ctx, countStmt, countArgs...).Scan(
		&res.Count,
	)
	if err != nil {
		return &res, err
	}

	q := query + filter + order + arrangement + offset + limit
	stmt, args := helper.ReplaceQueryParams(q, params)

	rows, err := r.db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		row := pb.GetRes{}
		var (
			createdAt    sql.NullString
			updatedAt    sql.NullString
			clientTypeId sql.NullString
		)

		err = rows.Scan(
			&row.Id,
			&row.Status,
			&row.Name,
			&row.AppId,
			&row.AppSecret,
			&row.RoleId,
			&createdAt,
			&updatedAt,
			&row.EnvironmentId,
			&row.ProjectId,
			&clientTypeId,
			&row.RpsLimit,
			&row.MonthlyRequestLimit,
			&row.UsedCount,
		)

		if err != nil {
			return nil, err
		}

		if createdAt.Valid {
			row.CreatedAt = createdAt.String
		}

		if clientTypeId.Valid {
			row.ClientTypeId = clientTypeId.String
		}

		if updatedAt.Valid {
			row.UpdatedAt = updatedAt.String
		}

		res.Data = append(res.Data, &row)
	}

	return &res, nil
}
func (r *apiKeysRepo) Get(ctx context.Context, req *pb.GetReq) (*pb.GetRes, error) {
	var (
		res pb.GetRes

		createdAt sql.NullString
		updatedAt sql.NullString
	)

	query := `SELECT
				id,
  				status,
  				name,
  				app_id,
  				app_secret,
  				role_id,
  				environment_id,
				project_id,
				client_type_id,
				rps_limit,
				monthly_request_limit,
				(SELECT count(1) FROM api_key_usage where api_key=app_id AND created_at >= date_trunc('month', current_date)),
  				created_at,
  				updated_at
			FROM
			    api_keys
			WHERE
			    id = $1`

	err := r.db.QueryRow(ctx, query, req.GetId()).Scan(
		&res.Id,
		&res.Status,
		&res.Name,
		&res.AppId,
		&res.AppSecret,
		&res.RoleId,
		&res.EnvironmentId,
		&res.ProjectId,
		&res.ClientTypeId,
		&res.RpsLimit,
		&res.MonthlyRequestLimit,
		&res.UsedCount,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		res.CreatedAt = createdAt.String
	}

	if updatedAt.Valid {
		res.UpdatedAt = updatedAt.String
	}
	return &res, nil
}
func (r *apiKeysRepo) Update(ctx context.Context, req *pb.UpdateReq) (rowsAffected int64, err error) {
	query := `UPDATE "api_keys" SET
				status = $1,
				name = $2,
				role_id = $3,
				client_type_id = $4,
				updated_at = now()
			WHERE
			    id = $5`

	res, err := r.db.Exec(
		ctx,
		query,
		req.GetStatus(),
		req.GetName(),
		req.GetRoleId(),
		req.GetClientTypeId(),
		req.GetId(),
	)

	if err != nil {
		return 0, err
	}

	rowsAffected = res.RowsAffected()
	return
}
func (r *apiKeysRepo) Delete(ctx context.Context, req *pb.DeleteReq) (rowsAffected int64, err error) {
	query := `DELETE FROM "api_keys"
				WHERE id = $1`

	result, err := r.db.Exec(ctx, query, req.Id)
	if err != nil {
		return rowsAffected, err
	}

	return result.RowsAffected(), nil
}

func (r *apiKeysRepo) GetByAppId(ctx context.Context, appId string) (*pb.GetRes, error) {
	var (
		res pb.GetRes

		createdAt sql.NullString
		updatedAt sql.NullString
	)

	query := `SELECT
				id,
  				status,
  				name,
  				app_id,
  				app_secret,
  				role_id,
  				environment_id,
				project_id,
				client_type_id,
				rps_limit,
				monthly_request_limit,
  				created_at,
  				updated_at
			FROM
			    api_keys
			WHERE
			    app_id = $1`

	err := r.db.QueryRow(ctx, query, appId).Scan(
		&res.Id,
		&res.Status,
		&res.Name,
		&res.AppId,
		&res.AppSecret,
		&res.RoleId,
		&res.EnvironmentId,
		&res.ProjectId,
		&res.ClientTypeId,
		&res.RpsLimit,
		&res.MonthlyRequestLimit,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		res.CreatedAt = createdAt.String
	}

	if updatedAt.Valid {
		res.UpdatedAt = updatedAt.String
	}
	return &res, nil
}

func (r *apiKeysRepo) GetEnvID(ctx context.Context, req *pb.GetReq) (*pb.GetRes, error) {

	res := &pb.GetRes{}

	query := `
		SELECT
  			environment_id,
			project_id,
			id,
			client_type_id,
			role_id,
			status,
			name,
			app_id,
			app_secret,
			rps_limit,
			monthly_request_limit
		FROM
			api_keys
		WHERE
			app_id = $1`

	err := r.db.QueryRow(ctx, query, req.GetId()).Scan(
		&res.EnvironmentId,
		&res.ProjectId,
		&res.Id,
		&res.ClientTypeId,
		&res.RoleId,
		&res.Status,
		&res.Name,
		&res.AppId,
		&res.AppSecret,
		&res.RpsLimit,
		&res.MonthlyRequestLimit,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while scanning")
	}

	return res, nil
}

func (r *apiKeysRepo) UpdateIsMonthlyLimitReached(ctx context.Context) error {
	query := `
		UPDATE api_keys SET 
			is_monthly_request_limit_reached = false
	`

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return errors.Wrap(err, "error while executing UpdateIsMonthlyLimitReached query")
	}

	return nil
}
