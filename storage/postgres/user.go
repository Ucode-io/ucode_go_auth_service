package postgres

import (
	"context"
	"database/sql"
	"errors"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/util"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
)

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) storage.UserRepoI {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) Create(ctx context.Context, entity *pb.CreateUserRequest) (pKey *pb.UserPrimaryKey, err error) {

	query := `INSERT INTO "user" (
		id,
		name,
		photo_url,
		phone,
		email,
		login,
		password,
		company_id
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8
	)`

	uuid, err := uuid.NewRandom()
	if err != nil {
		return pKey, err
	}

	_, err = r.db.Exec(ctx, query,
		uuid.String(),
		entity.Name,
		entity.PhotoUrl,
		entity.Phone,
		entity.Email,
		entity.Login,
		entity.Password,
		entity.CompanyId,
	)

	pKey = &pb.UserPrimaryKey{
		Id: uuid.String(),
	}

	return pKey, err
}

func (r *userRepo) GetByPK(ctx context.Context, pKey *pb.UserPrimaryKey) (res *pb.User, err error) {
	res = &pb.User{}
	query := `SELECT
		id,
		name,
		photo_url,
		phone,
		email,
		login,
		password,
		company_id
		-- TO_CHAR(expires_at, ` + config.DatabaseQueryTimeLayout + `) AS expires_at
		-- TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		-- TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM
		"user"
	WHERE
		id = $1`

	err = r.db.QueryRow(ctx, query, pKey.Id).Scan(
		&res.Id,
		&res.Name,
		&res.PhotoUrl,
		&res.Phone,
		&res.Email,
		&res.Login,
		&res.Password,
		&res.CompanyId,
		// &res.ExpiresAt,
		// &res.CreatedAt,
		// &res.UpdatedAt,
	)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *userRepo) GetListByPKs(ctx context.Context, pKeys *pb.UserPrimaryKeyList) (res *pb.GetUserListResponse, err error) {

	res = &pb.GetUserListResponse{}
	query := `SELECT
		id,
		name,
		photo_url,
		phone,
		email,
		login,
		password,
		created_at,
		updated_at,
		company_id
	FROM
		"user"
	WHERE
		id = ANY($1)`

	rows, err := r.db.Query(ctx, query, pq.Array(pKeys.Ids))
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			createdAt sql.NullString
			updatedAt sql.NullString
		)

		user := &pb.User{}
		err = rows.Scan(
			&user.Id,
			&user.Name,
			&user.PhotoUrl,
			&user.Phone,
			&user.Email,
			&user.Login,
			&user.Password,
			&createdAt,
			&updatedAt,
			&user.CompanyId,
		)

		if err != nil {
			return res, err
		}

		// if active.Valid {
		// 	user.Active = active.Int32
		// }

		// if expiresAt.Valid {
		// 	user.ExpiresAt = expiresAt.String
		// }

		// if createdAt.Valid {
		// 	user.CreatedAt = createdAt.String
		// }

		// if updatedAt.Valid {
		// 	user.UpdatedAt = updatedAt.String
		// }

		res.Users = append(res.Users, user)
	}

	return res, nil
}

func (r *userRepo) GetList(ctx context.Context, queryParam *pb.GetUserListRequest) (res *pb.GetUserListResponse, err error) {
	res = &pb.GetUserListResponse{}
	params := make(map[string]interface{})
	var arr []interface{}
	query := `SELECT
		id,
		name,
		company_id,
		photo_url,
		phone,
		email,
		login,
		password,
		created_at,
		updated_at
	FROM
		"user"`
	filter := " WHERE 1=1"
	order := " ORDER BY created_at"
	arrangement := " DESC"
	offset := " OFFSET 0"
	limit := " LIMIT 10"

	if len(queryParam.Search) > 0 {
		params["search"] = queryParam.Search
		filter += " AND ((name || phone || email || login) ILIKE ('%' || :search || '%'))"
	}

	//if len(queryParam.ClientPlatformId) > 0 {
	//	params["client_platform_id"] = queryParam.ClientPlatformId
	//	filter += " AND client_platform_id = :client_platform_id"
	//}
	// if len(queryParam.ProjectId) > 0 {
	// 	params["project_id"] = queryParam.ProjectId
	// 	filter += " AND project_id = :project_id"
	// }

	//if len(queryParam.ClientTypeId) > 0 {
	//	params["client_type_id"] = queryParam.ClientTypeId
	//	filter += " AND client_type_id = :client_type_id"
	//}

	if queryParam.Offset > 0 {
		params["offset"] = queryParam.Offset
		offset = " OFFSET :offset"
	}

	if queryParam.Limit > 0 {
		params["limit"] = queryParam.Limit
		limit = " LIMIT :limit"
	}

	cQ := `SELECT count(1) FROM "user"` + filter
	cQ, arr = helper.ReplaceQueryParams(cQ, params)
	err = r.db.QueryRow(ctx, cQ, arr...).Scan(
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
		obj := &pb.User{}
		var (
			active    sql.NullInt32
			expiresAt sql.NullString
			createdAt sql.NullString
			updatedAt sql.NullString
			companyID sql.NullString
		)

		err = rows.Scan(
			&obj.Id,
			&obj.Name,
			&companyID,
			&obj.PhotoUrl,
			&obj.Phone,
			&obj.Email,
			&obj.Login,
			&obj.Password,
			&active,
			&expiresAt,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return res, err
		}

		// if active.Valid {
		// 	obj.Active = active.Int32
		// }

		// if expiresAt.Valid {
		// 	obj.ExpiresAt = expiresAt.String
		// }

		// if createdAt.Valid {
		// 	obj.CreatedAt = createdAt.String
		// }

		// if updatedAt.Valid {
		// 	obj.UpdatedAt = updatedAt.String
		// }

		res.Users = append(res.Users, obj)
	}

	return res, nil
}

func (r *userRepo) Update(ctx context.Context, entity *pb.UpdateUserRequest) (rowsAffected int64, err error) {
	query := `UPDATE "user" SET
		name = :name,
		company_id = :company_id,
		photo_url = :photo_url,
		phone = :phone,
		email = :email,
		login = :login,
		updated_at = now()
	WHERE
		id = :id`

	params := map[string]interface{}{
		"id":         entity.GetId(),
		"name":       entity.GetName(),
		"photo_url":  entity.GetPhotoUrl(),
		"phone":      entity.GetPhone(),
		"email":      entity.GetEmail(),
		"login":      entity.GetLogin(),
		"company_id": entity.GetCompanyId(),
	}

	q, arr := helper.ReplaceQueryParams(query, params)
	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *userRepo) Delete(ctx context.Context, pKey *pb.UserPrimaryKey) (rowsAffected int64, err error) {
	queryDeleteFromUserProject := `DELETE FROM user_project WHERE user_id = $1`

	result, err := r.db.Exec(ctx, queryDeleteFromUserProject, pKey.Id)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (res *pb.User, err error) {
	res = &pb.User{}

	query := `SELECT
		id,
		name,
		photo_url,
		phone,
		email,
		login,
		password
		-- TO_CHAR(expires_at, ` + config.DatabaseQueryTimeLayout + `) AS expires_at,
		-- TO_CHAR(created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		-- TO_CHAR(updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM
		"user"
	WHERE`

	if util.IsValidEmail(username) {
		query = query + ` email = $1`
	} else if util.IsValidPhone(username) {
		query = query + ` phone = $1`
	} else {
		query = query + ` login = $1`
	}

	err = r.db.QueryRow(ctx, query, username).Scan(
		&res.Id,
		&res.Name,
		&res.PhotoUrl,
		&res.Phone,
		&res.Email,
		&res.Login,
		&res.Password,
		// &res.Active,
		// &res.ExpiresAt,
		// &res.CreatedAt,
		// &res.UpdatedAt,
	)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *userRepo) ResetPassword(ctx context.Context, user *pb.ResetPasswordRequest) (rowsAffected int64, err error) {
	query := `UPDATE "user" SET
		password = :password,
		updated_at = now()
	WHERE
		id = :id`

	params := map[string]interface{}{
		"id":       user.UserId,
		"password": user.Password,
	}

	q, arr := helper.ReplaceQueryParams(query, params)
	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *userRepo) GetUserProjects(ctx context.Context, userId string) (*models.GetUserProjects, error) {
	res := models.GetUserProjects{}

	query := `SELECT company_id,
      			array_agg(project_id)
				FROM user_project
				WHERE user_id = $1
				GROUP BY company_id`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			projects []string
			company  string
		)

		err = rows.Scan(&company, pq.Array(&projects))
		if err != nil {
			return nil, err
		}

		res.Companies = append(res.Companies, models.Companie{
			Id:       company,
			Projects: projects,
		})
	}

	return &res, nil
}

func (r *userRepo) AddUserToProject(ctx context.Context, req *pb.AddUserToProjectReq) (*pb.AddUserToProjectRes, error) {
	res := pb.AddUserToProjectRes{}

	query := `INSERT INTO
			user_project(user_id, company_id, project_id)
			VALUES ($1, $2, $3)
			RETURNING user_id, company_id, project_id`

	err := r.db.QueryRow(ctx, query, req.GetUserId(), req.GetCompanyId(), req.GetProjectId()).Scan(&res.UserId, &res.CompanyId, &res.ProjectId)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *userRepo) GetProjectsByUserId(ctx context.Context, req *pb.GetProjectsByUserIdReq) (*pb.GetProjectsByUserIdRes, error) {
	res := pb.GetProjectsByUserIdRes{}

	query := `SELECT
				array_agg(project_id)
			from user_project
			where user_id = $1`

	tmp := make([]string, 0, 20)
	err := r.db.QueryRow(ctx, query, req.GetUserId()).Scan(pq.Array(&tmp))
	if err != nil {
		return nil, err
	}
	res.ProjectIds = tmp
	return &res, nil
}

func (r *userRepo) GetUserIds(ctx context.Context, req *pb.GetUserListRequest) (*[]string, error) {

	query := `SELECT
				array_agg(user_id)
			from user_project
			where true=true`

	filter := ` and project_id = :project_id`
	params := map[string]interface{}{}
	params["project_id"] = req.ProjectId

	if len(req.Search) > 0 {
		params["search"] = req.Search
		filter += " AND ((name || phone || email || login) ILIKE ('%' || :search || '%'))"
	}

	query, args := helper.ReplaceQueryParams(query+filter, params)

	tmp := make([]string, 0, 20)
	err := r.db.QueryRow(ctx, query, args...).Scan(pq.Array(&tmp))
	if err != nil {
		return nil, err
	}

	return &tmp, nil
}

func (r *userRepo) GetUserByLoginType(ctx context.Context, req *pb.GetUserByLoginTypesRequest) (*pb.GetUserByLoginTypesResponse, error) {

	query := `SELECT
				id
			from "user" WHERE `
	var filter string
	params := map[string]interface{}{}
	if req.Email != "" {
		filter = "email = :email"
		params["email"] = req.Email
	}
	if req.Login != "" {
		if filter != "" {
			filter += "OR login = :login"
		} else {
			filter = "login = :login"
		}
		params["login"] = req.Login
	}
	if req.Phone != "" {
		if filter != "" {
			filter += " OR phone = :login"
		} else {
			filter = "phone = :" + req.Phone
		}
		params["phone"] = req.Phone
	}

	query, args := helper.ReplaceQueryParams(query+filter, params)

	var userId string
	err := r.db.QueryRow(ctx, query, args...).Scan(&userId)
	if err == pgx.ErrNoRows {
		return nil, errors.New("not found")
	} else if err != nil {
		return nil, err
	}

	return &pb.GetUserByLoginTypesResponse{
		UserId: userId,
	}, nil
}
