package postgres

import (
	"context"
	"ucode/ucode_go_auth_service/storage"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
)

type loginStrategyRepoI struct {
	db *Pool
}

func NewLoginStrategy(db *Pool) storage.LoginStrategyI {
	return &loginStrategyRepoI{
		db: db,
	}
}

func (ls *loginStrategyRepoI) GetList(ctx context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "login_strategy.GetList")
	defer dbSpan.Finish()

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
		)
		err = rows.Scan(
			&row.Id,
			&row.Type,
			&row.ProjectId,
			&row.EnvId,
		)
		if err != nil {
			return nil, err
		}
		res.LoginStrategies = append(res.LoginStrategies, &row)
	}
	return &pb.GetListResponse{LoginStrategies: res.LoginStrategies}, nil
}

func (ls *loginStrategyRepoI) GetByID(ctx context.Context, req *pb.LoginStrategyPrimaryKey) (*pb.LoginStrategy, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "login_strategy.GetByID")
	defer dbSpan.Finish()

	var (
		res = pb.LoginStrategy{}
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
		&res.Type,
		&res.ProjectId,
		&res.EnvId,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (ls *loginStrategyRepoI) Upsert(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "login_strategy.Upsert")
	defer dbSpan.Finish()

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
			entity.Type,
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
	resp.RowsAffected = int32(len(req.LoginStrategies))
	return &resp, nil
}
