package postgres

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v4/pgxpool"
	"strings"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/storage"
)

type apiKeysRepo struct {
	db *pgxpool.Pool
}

func NewApiKeysRepo(db *pgxpool.Pool) storage.ApiKeysRepoI {
	return &apiKeysRepo{
		db: db,
	}
}

func (r *apiKeysRepo) Create(ctx context.Context, req *pb.CreateReq) (*pb.CreateRes, error) {
	var (
		res pb.CreateRes

		createdAt sql.NullString
		updatedAt sql.NullString
	)

	query := `INSERT INTO api_keys(id, name, app_id, app_secret, role_id, resource_environment_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, now(), now()) RETURNING id, status, name, app_id, app_secret, role_id, created_at, updated_at, resource_environment_id`

	err := r.db.QueryRow(
		ctx,
		query,
		req.GetId(),
		req.GetName(),
		req.GetAppId(),
		req.GetAppSecret(),
		req.GetRoleId(),
		req.GetResourceEnvironmentId(),
	).Scan(&res.Id, &res.Status, &res.Name, &res.AppId, &res.AppSecret, &res.RoleId, &createdAt, &updatedAt, &res.ResourceEnvironmentId)

	if err != nil {
		return nil, err
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
  				updated_at
			FROM
			    api_keys
			WHERE
			    resource_environment_id = $1`

	rows, err := r.db.Query(ctx, query, req.GetResourceEnvironmentId())
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
			&row.AppSecret,
			&row.RoleId,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, err
		}
		res.Data = append(res.Data, &row)
		res.Count++
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
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
func (r *apiKeysRepo) Update(ctx context.Context, req *pb.UpdateReq) (rowsAffected int64, err error) {
	query := `UPDATE "api_keys" SET
				status = $1,
				name = $2,
				app_id = $3,
				app_secret = $4,
				role_id = $5,
				updated_at = now()
			WHERE
			    id = $6`

	res, err := r.db.Exec(
		ctx,
		query,
		req.GetStatus(),
		req.GetName(),
		req.GetAppId(),
		req.GetAppSecret(),
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

	keys := strings.Split(req.GetIds(), ",")

	for _, key := range keys {
		result, err := r.db.Exec(ctx, query, &key)
		if err != nil {
			return rowsAffected, err
		}
		rowsAffected += result.RowsAffected()
	}
	return
}
