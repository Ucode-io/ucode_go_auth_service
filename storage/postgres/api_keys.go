package postgres

import (
	"context"
	"database/sql"
	"fmt"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/util"
	"ucode/ucode_go_auth_service/storage"

	"github.com/jackc/pgx/v4/pgxpool"
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

	query := `INSERT INTO api_keys(id, name, app_id, app_secret, role_id, resource_environment_id, project_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now()) RETURNING id, status, name, app_id, app_secret, role_id, created_at, updated_at, resource_environment_id, project_id`

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		req.GetName(),
		appId,
		appSecret,
		req.GetRoleId(),
		req.GetResourceEnvironmentId(),
		req.GetProjectId(),
	).Scan(&res.Id, &res.Status, &res.Name, &res.AppId, &res.AppSecret, &res.RoleId, &createdAt, &updatedAt, &res.ResourceEnvironmentId, &res.ProjectId)

	fmt.Println("err::", err)
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
  				role_id,
  				created_at,
  				updated_at,
  				resource_environment_id,
				project_id
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

	if util.IsValidUUID(req.ResourceEnvironmentId) {
		filter += ` AND resource_environment_id = :resource_environment_id`
		params["resource_environment_id"] = req.GetResourceEnvironmentId()
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
			createdAt sql.NullString
			updatedAt sql.NullString
		)

		err = rows.Scan(
			&row.Id,
			&row.Status,
			&row.Name,
			&row.AppId,
			&row.RoleId,
			&createdAt,
			&updatedAt,
			&row.ResourceEnvironmentId,
			&row.ProjectId,
		)
		if err != nil {
			return nil, err
		}

		if createdAt.Valid {
			row.CreatedAt = createdAt.String
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
  				resource_environment_id,
				project_id,
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
		&res.ResourceEnvironmentId,
		&res.ProjectId,
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
				updated_at = now()
			WHERE
			    id = $4`

	res, err := r.db.Exec(
		ctx,
		query,
		req.GetStatus(),
		req.GetName(),
		req.GetRoleId(),
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
