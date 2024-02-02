package postgres

import (
	"context"
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
			rps_limit-(SELECT COUNT(*) FROM api_key_usage WHERE api_key = $1 AND created_at >= now() - interval '1 second') as rps_count,
			monthly_request_limit-(SELECT COUNT(*) FROM api_key_usage WHERE api_key = $1 AND created_at >= date_trunc('month', current_date)) as monthly_request_count
		FROM api_keys 
		WHERE app_id = $1;
	`

	err := r.db.QueryRow(ctx, query, req.GetApiKey()).Scan(
		&res.RpsCount,
		&res.MonthlyCount,
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
			api_key, 
			request_count
		) VALUES(
			$1,
			$2
		)
	`

	_, err := r.db.Exec(context.Background(), query,
		req.GetApiKey(),
		req.GetRequestCount(),
	)

	return err
}
