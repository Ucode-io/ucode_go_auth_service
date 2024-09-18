package postgres

import (
	"context"
	"strings"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func (r *userRepo) UpdateSyncUser(ctx context.Context, req *pb.UpdateSyncUserRequest, loginType string) (*pb.SyncUserResponse, error) {
	var (
		loginValue, userId string
		resp               = &pb.SyncUserResponse{}
		resetPasswordReq   *pb.ResetPasswordRequest
	)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	query := `SELECT
		id
	FROM
		"user"
	WHERE`

	switch loginType {
	case "email":
		query = query + ` LOWER(email) = $1`
		loginValue = strings.ToLower(req.GetEmail())
		resetPasswordReq = &pb.ResetPasswordRequest{
			Email:    req.GetEmail(),
			UserId:   req.GetGuid(),
			Password: req.GetPassword(),
		}
	case "phone":
		query = query + ` phone = $1`
		loginValue = strings.ToLower(req.GetPhone())
		resetPasswordReq = &pb.ResetPasswordRequest{
			Phone:    req.GetPhone(),
			UserId:   req.GetGuid(),
			Password: req.GetPassword(),
		}
	case "login":
		query = query + ` LOWER(login) = $1`
		loginValue = strings.ToLower(req.GetLogin())
		resetPasswordReq = &pb.ResetPasswordRequest{
			Login:    req.GetLogin(),
			UserId:   req.GetGuid(),
			Password: req.GetPassword(),
		}
	}

	err = tx.QueryRow(ctx, query, loginValue).Scan(&userId)
	if err == pgx.ErrNoRows {
		_, err = r.ResetPassword(ctx, resetPasswordReq, tx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to reset password")
		}

		resp.UserId = req.GetGuid()
	} else if err == nil {
		if _, err = uuid.Parse(userId); err == nil {
			query = `
				UPDATE 
					user_project
				SET user_id = $1
				WHERE user_id = $2
		  		AND project_id = $3
		  		AND client_type_id = $4
		  		AND role_id = $5
		  		AND env_id = $6
				AND company_id = $7`

			_, err = tx.Exec(ctx,
				query,
				userId,
				req.GetGuid(),
				req.GetProjectId(),
				req.GetClientTypeId(),
				req.GetRoleId(),
				req.GetEnvId(),
				req.GetCompanyId(),
			)
			if err != nil {
				return nil, errors.Wrap(err, "failed to update user_project")
			}

			resp.UserId = userId
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return resp, nil
}
