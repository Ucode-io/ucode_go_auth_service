package postgres

import (
	"context"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type loginStrategyRepoI struct {
	db *pgxpool.Pool
}

func NewLoginStrategy(db *pgxpool.Pool) storage.LoginStrategyI {
	return &loginStrategyRepoI{
		db: db,
	}
}

func (ls *loginStrategyRepoI) GetList(ctx context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	var (
		res = pb.GetListResponse{}
	)
	query := `SELECT
				id,
				type,
				project_id,
				env_id
  				FROM
				login_strategy WHERE project_id = $1`

	rows, err := ls.db.Query(ctx, query, req.GetProjectId())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			row = pb.LoginStrategy{}
			t   string
		)
		err = rows.Scan(
			&row.Id,
			&t,
			&row.ProjectId,
			&row.EnvId,
		)
		if err != nil {
			return nil, err
		}
		row.Type = pb.LoginStrategyType(helper.ParsePsqlTypeToEnum(t))
		res.LoginStrategies = append(res.LoginStrategies, &row)
	}
	return &pb.GetListResponse{LoginStrategies: res.LoginStrategies}, nil
}

func (ls *loginStrategyRepoI) GetByID(ctx context.Context, req *pb.LoginStrategyPrimaryKey) (*pb.LoginStrategy, error) {
	var (
		res = pb.LoginStrategy{}
		t   string
	)
	query := `SELECT
				id,
				type,
				project_id,
				env_id
  				FROM
				login_strategy WHERE id = $1`
	err := ls.db.QueryRow(ctx, query, req.GetId()).Scan(
		&res.Id,
		&t,
		&res.ProjectId,
		&res.EnvId,
	)
	if err != nil {
		return nil, err
	}
	res.Type = pb.LoginStrategyType(helper.ParsePsqlTypeToEnum(t))
	return &res, nil
}

func (ls *loginStrategyRepoI) Upsert(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	var (
		resp = pb.UpdateResponse{}
	)
	const stmt = `INSERT INTO "login_strategy" (
		id,
		type,
		project_id,
		env_id
	) VALUES (
		$1,
		$2,
		$3,
		$4
	) ON CONFLICT ("id") DO UPDATE SET "type" = $2, "project_id" = $3, "env_id" = $4
	RETURNING *`

	batch := &pgx.Batch{}
	for _, entity := range req.GetLoginStrategies() {
		batch.Queue(
			stmt,
			entity.Id,
			entity.Type.String(),
			entity.ProjectId,
			entity.EnvId,
		)
	}
	res := ls.db.SendBatch(ctx, batch)
	defer res.Close()
	_, err := res.Exec()
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
