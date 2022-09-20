package postgres

import (
	"context"
	"errors"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/storage"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type emailRepo struct {
	db *pgxpool.Pool
}

func NewEmailRepo(db *pgxpool.Pool) storage.EmailRepoI {
	return &emailRepo{
		db: db,
	}
}

func (e *emailRepo) Create(ctx context.Context, input *pb.Email) (*pb.Email, error) {
	query := `INSERT INTO "email_sms" (
		id,
		email,
		otp,
		expires_at
	) VALUES (
		$1,
		$2,
		$3,
		$4
	)`

	_, err := e.db.Exec(ctx, query,
		input.Id,
		input.Email,
		input.Otp,
		input.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return input, nil
}

func (e *emailRepo) GetByPK(ctx context.Context, pKey *pb.EmailOtpPrimaryKey) (res *pb.Email, err error) {
	res = &pb.Email{}
	query := `SELECT
					id,
					email,
					otp
	FROM
		"email_sms"
	WHERE
		id = $1 AND expires_at > NOW()`

	err = e.db.QueryRow(ctx, query, pKey.Id).Scan(
		&res.Id,
		&res.Email,
		&res.Otp,
	)
	if err == pgx.ErrNoRows {
		err := errors.New("Otp has been expired")
		return nil, err
	} else if err != nil {
		return res, err
	}

	return res, nil
}
