package postgres

import (
	"context"
	"time"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/storage"

	"github.com/jackc/pgx/v4/pgxpool"
)

type apiKeyUsageRepo struct {
	db *pgxpool.Pool
}

func NewApiKeyUsageRepo(db *pgxpool.Pool) storage.ApiKeyUsageRepoI {
	return &apiKeyUsageRepo{
		db: db,
	}
}

func (r *apiKeyUsageRepo) CheckLimit(ctx context.Context, req *pb.CheckLimitRequest) (*pb.CheckLimitResponse, error) {
	var (
		res pb.CheckLimitResponse
	)

	query := `
		SELECT
			is_monthly_request_limit_reached,
			rps_limit-(SELECT COUNT(1) FROM api_key_usage WHERE api_key = $1 AND creation_time = $2) as rps_count
		FROM api_keys 
		WHERE app_id = $1;
	`

	err := r.db.QueryRow(
		ctx,
		query,
		req.GetApiKey(),
		time.Now().Format(time.DateTime),
	).Scan(
		&res.IsLimitReached,
		&res.RpsCount,
	)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *apiKeyUsageRepo) Get(ctx context.Context, req *pb.GetApiKeyUsageReq) (*pb.ApiKeyUsage, error) {
	var (
		res pb.ApiKeyUsage
	)

	return &res, nil
}

func (r *apiKeyUsageRepo) Create(ctx context.Context, req *pb.ApiKeyUsage) error {

	query := `
		INSERT INTO api_key_usage (
			api_key
		) VALUES(
			$1
		)
	`

	_, err := r.db.Exec(context.Background(), query,
		req.GetApiKey(),
	)

	return err
}
