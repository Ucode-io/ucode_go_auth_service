package postgres

import (
	"context"
	"database/sql"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (r *companyRepo) GetList(ctx context.Context, queryParam *pb.GetComapnyListRequest) (*pb.GetListCompanyResponse, error) {

	res := &pb.GetListCompanyResponse{}
	params := make(map[string]interface{})
	var arr []interface{}
	query := `SELECT
		id, 
		name,
		TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM
		"company"`
	filter := " WHERE 1=1"
	order := " ORDER BY created_at"
	arrangement := " DESC"
	offset := " OFFSET 0"
	limit := ""

	if len(queryParam.Search) > 0 {
		params["search"] = queryParam.Search
		filter += " AND (name ILIKE ('%' || :search || '%'))"
	}

	if queryParam.Offset > 0 {
		params["offset"] = queryParam.Offset
		offset = " OFFSET :offset"
	}

	if queryParam.Limit > 0 {
		params["limit"] = queryParam.Limit
		limit = " LIMIT :limit"
	}

	cQ := `SELECT count(1) FROM "company"` + filter
	cQ, arr = helper.ReplaceQueryParams(cQ, params)
	err := r.db.QueryRow(ctx, cQ, arr...).Scan(
		&res.Count,
	)
	if err != nil {
		return res, err
	}

	q := query + filter + order + arrangement + offset + limit

	q, arr = helper.ReplaceQueryParams(q, params)
	rows, err := r.db.Query(ctx, q, arr...)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		obj := &pb.Company{}
		var (
			createdAt sql.NullString
			updatedAt sql.NullString
		)

		err = rows.Scan(
			&obj.Id,
			&obj.Name,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return res, err
		}

		if createdAt.Valid {
			obj.CreatedAt = createdAt.String
		}

		if updatedAt.Valid {
			obj.UpdatedAt = updatedAt.String
		}

		res.Companies = append(res.Companies, obj)
	}

	return res, nil
}

func (r *companyRepo) GetByID(ctx context.Context, pKey *pb.CompanyPrimaryKey) (*pb.Company, error) {

	res := &pb.Company{}
	query := `SELECT
		id,
		name,
		TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM
		"company"
	WHERE
		id = $1`

	err := r.db.QueryRow(ctx, query, pKey.Id).Scan(
		&res.Id,
		&res.Name,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *companyRepo) TransferOwnership(ctx context.Context, companyID, ownerID string) (rowsAffected int64, err error) {

	query := `UPDATE "company" SET
		owner_id = :owner_id,
		updated_at = now()
	WHERE
		id = :id`

	params := map[string]interface{}{
		"id":       companyID,
		"owner_id": ownerID,
	}

	q, arr := helper.ReplaceQueryParams(query, params)
	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}
