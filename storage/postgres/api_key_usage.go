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
		INSERT INTO api_key_usage (api_key, request_count)
	`

	_, err := r.db.Exec(ctx, query,
		req.ApiKey,
		req.RequestCount,
	)

	return err
}
