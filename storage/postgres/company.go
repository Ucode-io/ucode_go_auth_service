package postgres

import (
	"context"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type companyRepo struct {
	db *pgxpool.Pool
}

func NewCompanyRepo(db *pgxpool.Pool) storage.CompanyRepoI {
	return &companyRepo{
		db: db,
	}
}

func (r *companyRepo) Register(ctx context.Context, entity *pb.RegisterCompanyRequest) (pKey *pb.CompanyPrimaryKey, err error) {
	query := `INSERT INTO "company" (
		id,
		name
	) VALUES (
		$1,
		$2
	)`

	uuid, err := uuid.NewRandom()
	if err != nil {
		return pKey, err
	}

	_, err = r.db.Exec(ctx, query,
		uuid.String(),
		entity.Name,
	)

	pKey = &pb.CompanyPrimaryKey{
		Id: uuid.String(),
	}

	return pKey, err
}

func (r *companyRepo) Update(ctx context.Context, entity *pb.UpdateCompanyRequest) (rowsAffected int64, err error) {
	query := `UPDATE "company" SET
		name = :name,
		updated_at = now()
	WHERE
		id = :id`

	params := map[string]interface{}{
		"id":   entity.Id,
		"name": entity.Name,
	}

	q, arr := helper.ReplaceQueryParams(query, params)
	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *companyRepo) Remove(ctx context.Context, pKey *pb.CompanyPrimaryKey) (rowsAffected int64, err error) {
	query := `DELETE FROM "company" WHERE id = $1`

	result, err := r.db.Exec(ctx, query, pKey.Id)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}
