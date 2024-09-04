package postgres

import (
	"context"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/storage"

	"github.com/opentracing/opentracing-go"
)

type userRelationRepo struct {
	db *Pool
}

func NewUserRelationRepo(db *Pool) storage.UserRelationRepoI {
	return &userRelationRepo{
		db: db,
	}
}

func (r *userRelationRepo) Add(ctx context.Context, entity *pb.AddUserRelationRequest) (pKey *pb.UserRelationPrimaryKey, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "user_relation.Add")
	defer dbSpan.Finish()

	query := `INSERT INTO "user_relation" (
		user_id,
		relation_id
	) VALUES (
		$1,
		$2
	)`

	_, err = r.db.Exec(ctx, query,
		entity.UserId,
		entity.RelationId,
	)

	pKey = &pb.UserRelationPrimaryKey{
		UserId:     entity.UserId,
		RelationId: entity.RelationId,
	}

	return pKey, err
}

func (r *userRelationRepo) GetByPK(ctx context.Context, pKey *pb.UserRelationPrimaryKey) (res *pb.UserRelation, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "user_relation.GetByPK")
	defer dbSpan.Finish()

	res = &pb.UserRelation{}
	query := `SELECT
		user_id,
		relation_id
	FROM
		"user_relation"
	WHERE
		user_id = $1 AND relation_id = $2`

	err = r.db.QueryRow(ctx, query, pKey.UserId, pKey.RelationId).Scan(
		&res.UserId,
		&res.RelationId,
	)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *userRelationRepo) Remove(ctx context.Context, pKey *pb.UserRelationPrimaryKey) (rowsAffected int64, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "user_relation.Remove")
	defer dbSpan.Finish()

	query := `DELETE FROM
		"user_relation"
	WHERE
		user_id = $1 AND relation_id = $2`

	result, err := r.db.Exec(ctx, query, pKey.UserId, pKey.RelationId)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}
