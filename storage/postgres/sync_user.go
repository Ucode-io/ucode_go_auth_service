package postgres

import (
	"context"
	"strings"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
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
	defer func() {
		_ = tx.Rollback(ctx)
	}()

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
	case "tin":
		query = query + ` tin = $1`
		loginValue = req.GetTin()
		resetPasswordReq = &pb.ResetPasswordRequest{
			Tin:    req.GetTin(),
			UserId: req.GetGuid(),
		}
	}

	err = tx.QueryRow(ctx, query, loginValue).Scan(&userId)
	if err == pgx.ErrNoRows {
		pKey, err := r.UpdateLoginStrategy(ctx, req, resetPasswordReq, tx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to reset password")
		}

		resp.UserId = pKey
	} else if err == nil {
		if loginType == "login" {
			return nil, errors.New("login already exists")
		}
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
	} else {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return resp, nil
}

func (r *userRepo) UpdateLoginStrategy(ctx context.Context, req *pb.UpdateSyncUserRequest, user *pb.ResetPasswordRequest, tx pgx.Tx) (string, error) {
	var (
		count  int
		userId string
	)

	query := `
		SELECT
			COUNT(*)
		FROM
			"user_project"
		WHERE
			user_id = $1`

	err := tx.QueryRow(ctx, query, req.GetGuid()).Scan(&count)
	if err != nil {
		return "", errors.Wrap(err, "failed to get user_project count")
	}

	switch {
	case count == 0:
		var (
			clientTypeId, roleId, envId pgtype.UUID
		)

		pKey, err := r.CreateWithTx(ctx, &pb.CreateUserRequest{
			Login:     req.GetLogin(),
			Password:  req.GetPassword(),
			Email:     req.GetEmail(),
			Phone:     req.GetPhone(),
			CompanyId: req.GetCompanyId(),
			Tin:       req.GetTin(),
		}, tx)
		if err != nil {
			return "", errors.Wrap(err, "failed to create user")
		}

		userId = pKey.GetId()

		if req.GetClientTypeId() != "" {
			err := clientTypeId.Set(req.GetClientTypeId())
			if err != nil {
				return "", errors.Wrap(err, "failed to set client type id")
			}
		} else {
			clientTypeId.Status = pgtype.Null
		}
		if req.GetRoleId() != "" {
			err := roleId.Set(req.GetRoleId())
			if err != nil {
				return "", errors.Wrap(err, "failed to set role id")
			}
		} else {
			roleId.Status = pgtype.Null
		}
		if req.GetEnvId() != "" {
			err := envId.Set(req.GetEnvId())
			if err != nil {
				return "", errors.Wrap(err, "failed to set env id")
			}
		} else {
			envId.Status = pgtype.Null
		}

		query = `INSERT INTO
				user_project(user_id, company_id, project_id, client_type_id, role_id, env_id)
				VALUES ($1, $2, $3, $4, $5, $6)`

		_, err = tx.Exec(ctx,
			query,
			userId,
			req.GetCompanyId(),
			req.GetProjectId(),
			clientTypeId,
			roleId,
			envId,
		)
		if err != nil {
			return "", errors.Wrap(err, "failed to insert user to project")
		}
	case count == 1:
		_, err = r.ResetPassword(ctx, user, tx)
		if err != nil {
			return "", errors.Wrap(err, "failed to reset password")
		}
		userId = req.GetGuid()
	case count > 1:
		pKey, err := r.CreateWithTx(ctx, &pb.CreateUserRequest{
			Login:     req.GetLogin(),
			Password:  req.GetPassword(),
			Email:     req.GetEmail(),
			Phone:     req.GetPhone(),
			CompanyId: req.GetCompanyId(),
			Tin:       req.GetTin(),
		}, tx)
		if err != nil {
			return "", errors.Wrap(err, "failed to create user")
		}

		userId = pKey.GetId()

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
			return "", errors.Wrap(err, "failed to update user_project")
		}
	}

	return userId, nil
}

func (r *userRepo) CreateWithTx(ctx context.Context, entity *pb.CreateUserRequest, tx pgx.Tx) (pKey *pb.UserPrimaryKey, err error) {
	query := `INSERT INTO "user" (
		id,
		phone,
		email,
		login,
		password,
		company_id,
		hash_type
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		'bcrypt'
	)`

	id, err := uuid.NewRandom()
	if err != nil {
		return pKey, errors.Wrap(err, "failed to generate uuid")
	}

	_, err = tx.Exec(ctx, query,
		id.String(),
		entity.GetPhone(),
		entity.GetEmail(),
		entity.GetLogin(),
		entity.GetPassword(),
		entity.GetCompanyId(),
	)
	if err != nil {
		return pKey, errors.Wrap(err, "failed to create user")
	}

	pKey = &pb.UserPrimaryKey{
		Id: id.String(),
	}

	return pKey, nil
}
