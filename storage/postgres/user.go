package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/saidamir98/udevs_pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
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
		-- name,
		-- photo_url,
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
		$6
		-- $7,
		-- $8
	)`

	id, err := uuid.NewRandom()
	if err != nil {
		return pKey, err
	}

	_, err = r.db.Exec(ctx, query,
		id.String(),
		// entity.GetName(),
		// entity.GetPhotoUrl(),
		entity.GetPhone(),
		entity.GetEmail(),
		entity.GetLogin(),
		entity.GetPassword(),
		entity.GetCompanyId(),
	)

	pKey = &pb.UserPrimaryKey{
		Id: id.String(),
	}

	return pKey, err
}

func (r *userRepo) GetByPK(ctx context.Context, pKey *pb.UserPrimaryKey) (res *pb.User, err error) {
	res = &pb.User{}
	var (
		lan  pb.Language
		time pb.Timezone
	)
	query := `SELECT
		u.id,
		-- coalesce(u.name, ''),
		-- coalesce(u.photo_url, ''),
		u.phone,
		u.email,
		u.login,
		-- u.password,
		u.company_id
		-- coalesce(t.id::VARCHAR, ''),
		-- coalesce(t.name, ''),
		-- coalesce(t.text, ''),
		-- coalesce(l.id::VARCHAR, ''),
		-- coalesce(l.name, ''),
		-- coalesce(l.short_name, ''),
		-- coalesce(l.native_name, '')
		-- TO_CHAR(u.expires_at, ` + config.DatabaseQueryTimeLayout + `) AS expires_at
		-- TO_CHAR(u.created_at, ` + config.DatabaseQueryTimeLayout + `) AS created_at,
		-- TO_CHAR(u.updated_at, ` + config.DatabaseQueryTimeLayout + `) AS updated_at
	FROM
		"user" u
		-- LEFT JOIN "language" l on u.language_id = l.id
		-- LEFT JOIN "timezone" t on u.timezone_id = t.id
	WHERE
		u.id = $1`
	err = r.db.QueryRow(ctx, query, pKey.Id).Scan(
		&res.Id,
		// &res.Name,
		// &res.PhotoUrl,
		&res.Phone,
		&res.Email,
		&res.Login,
		// &res.Password,
		&res.CompanyId,
		// &time.Id,
		// &time.Name,
		// &time.Text,
		// &lan.Id,
		// &lan.Name,
		// &lan.ShortName,
		// &lan.NativeName,
		// &res.ExpiresAt,
		// &res.CreatedAt,
		// &res.UpdatedAt,
	)
	if err != nil {
		return res, err
	}
	res.Language = &lan
	res.Timezone = &time

	return res, nil
}

func (r *userRepo) GetListByPKs(ctx context.Context, pKeys *pb.UserPrimaryKeyList) (res *pb.GetUserListResponse, err error) {

	res = &pb.GetUserListResponse{}
	query := `SELECT
		id,
		-- name,
		-- photo_url,
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
			// &user.Name,
			// &user.PhotoUrl,
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
		-- name,
		company_id,
		-- photo_url,
		phone,
		email,
		login,
		-- password,
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
		filter += " AND ((phone || email || login) ILIKE ('%' || :search || '%'))"
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
			// &obj.Name,
			&companyID,
			// &obj.PhotoUrl,
			&obj.Phone,
			&obj.Email,
			&obj.Login,
			// &obj.Password,
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
		-- name = :name,
		company_id = :company_id,
		-- photo_url = :photo_url,
		phone = :phone,
		email = :email,
		login = :login,
		updated_at = now()
    	-- language_id = :language_id,
        -- timezone_id = :timezone_id
	WHERE
		id = :id`

	params := map[string]interface{}{
		"id": entity.GetId(),
		// "name":        entity.GetName(),
		// "photo_url":   entity.GetPhotoUrl(),
		"phone":      entity.GetPhone(),
		"email":      entity.GetEmail(),
		"login":      entity.GetLogin(),
		"company_id": entity.GetCompanyId(),
		// "language_id": entity.GetLanguageId(),
		// "timezone_id": entity.GetTimezoneId(),
	}
	q, arr := helper.ReplaceQueryParams(query, params)
	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, nil
}

func (r *userRepo) Delete(ctx context.Context, pKey *pb.UserPrimaryKey) (int64, error) {

	// return 0, nil
	if pKey.GetIsTest() {
		queryDeleteFromUserProject := `DELETE FROM user_project WHERE user_id = $1`

		result, err := r.db.Exec(ctx, queryDeleteFromUserProject, pKey.Id)
		if err != nil {
			return 0, err
		}
		rowsAffected := result.RowsAffected()
		if rowsAffected == 0 {
			return 0, errors.New("user not found")
		}

	}
	// result, err := r.db.Exec(ctx, queryDeleteFromUserProject, pKey.Id)
	// if err != nil {
	// 	return 0, err
	// }
	// rowsAffected = result.RowsAffected()
	// if rowsAffected == 0 {
	// 	return 0, errors.New("user not found")
	// }

	result, err := r.db.Exec(ctx, `DELETE FROM "user" WHERE id = $1`, pKey.GetId())
	if err != nil {
		return 0, errors.Wrap(err, "delete user error")
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return 0, errors.New("user not found")
	}

	return 0, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (res *pb.User, err error) {
	res = &pb.User{}

	query := `SELECT
		id,
		-- name,
		-- coalesce(photo_url, ''),
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

	lowercasedUsername := strings.ToLower(username)

	if IsValidEmailNew(username) {
		query = query + ` LOWER(email) = $1`
	} else if util.IsValidPhone(username) {
		query = query + ` phone = $1`
	} else {
		query = query + ` LOWER(login) = $1`
	}

	err = r.db.QueryRow(ctx, query, lowercasedUsername).Scan(
		&res.Id,
		// &res.Name,
		// &res.PhotoUrl,
		&res.Phone,
		&res.Email,
		&res.Login,
		&res.Password,
		// &res.Active,
		// &res.ExpiresAt,
		// &res.CreatedAt,
		// &res.UpdatedAt,
	)
	if err == pgx.ErrNoRows && IsValidEmailNew(username) {
		queryIf := `
					SELECT
						id,
						phone,
						email,
						login,
						password
					FROM
						"user"
					WHERE
				`

		queryIf = queryIf + ` LOWER(login) = $1`

		err = r.db.QueryRow(ctx, queryIf, lowercasedUsername).Scan(
			&res.Id,
			&res.Phone,
			&res.Email,
			&res.Login,
			&res.Password,
		)
		if err == pgx.ErrNoRows {
			return res, nil
		}
		return res, nil
	}

	if err == pgx.ErrNoRows {
		return res, nil
	}

	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *userRepo) ResetPassword(ctx context.Context, user *pb.ResetPasswordRequest) (rowsAffected int64, err error) {
	query := `UPDATE "user" SET
		login = :login,
		email = :email,
		password = :password,
		phone = :phone,
		updated_at = now()
	WHERE
		id = :id`

	params := map[string]interface{}{
		"id":       user.UserId,
		"login":    user.Login,
		"email":    user.Email,
		"password": user.Password,
		"phone":    user.Phone,
	}

	q, arr := helper.ReplaceQueryParams(query, params)
	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *userRepo) GetUserProjectClientTypes(ctx context.Context, req *models.UserProjectClientTypeRequest) (res *models.UserProjectClientTypeResponse, err error) {
	res = &models.UserProjectClientTypeResponse{}

	query := `SELECT 
				array_agg(client_type_id) as client_type_ids
			FROM user_project 
			WHERE user_id = $1
			AND project_id = $2
			AND client_type_id IS NOT NULL
			GROUP BY  user_id`

	err = r.db.QueryRow(ctx, query, req.UserId, req.ProjectId).Scan(
		&res.ClientTypeIds,
	)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *userRepo) GetUserProjects(ctx context.Context, userId string) (*pb.GetUserProjectsRes, error) {
	res := pb.GetUserProjectsRes{}

	query := `SELECT company_id,
      			array_agg( DISTINCT project_id)
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

		res.Companies = append(res.Companies, &pb.UserCompany{
			Id:         company,
			ProjectIds: projects,
		})
	}

	return &res, nil
}

func (r *userRepo) AddUserToProject(ctx context.Context, req *pb.AddUserToProjectReq) (*pb.AddUserToProjectRes, error) {
	res := pb.AddUserToProjectRes{}

	var (
		clientTypeId, roleId, envId pgtype.UUID
	)
	if req.GetClientTypeId() != "" {
		err := clientTypeId.Set(req.GetClientTypeId())
		if err != nil {
			return nil, err
		}
	} else {
		clientTypeId.Status = pgtype.Null
	}
	if req.GetRoleId() != "" {
		err := roleId.Set(req.GetRoleId())
		if err != nil {
			return nil, err
		}
	} else {
		roleId.Status = pgtype.Null
	}
	if req.GetEnvId() != "" {
		err := envId.Set(req.GetEnvId())
		if err != nil {
			return nil, err
		}
	} else {
		envId.Status = pgtype.Null
	}

	query := `INSERT INTO
			user_project(user_id, company_id, project_id, client_type_id, role_id, env_id)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING user_id, company_id, project_id, client_type_id, role_id, env_id`

	err := r.db.QueryRow(ctx,
		query,
		req.GetUserId(),
		req.GetCompanyId(),
		req.GetProjectId(),
		clientTypeId,
		roleId,
		envId,
	).Scan(
		&res.UserId,
		&res.CompanyId,
		&res.ProjectId,
		&clientTypeId,
		&roleId,
		&envId,
	)
	if err != nil {
		return nil, err
	}
	if roleId.Status != pgtype.Null {
		req.RoleId = fmt.Sprintf("%v", roleId.Status)
	}
	if clientTypeId.Status != pgtype.Null {
		req.ClientTypeId = fmt.Sprintf("%v", clientTypeId.Status)
	}
	if envId.Status != pgtype.Null {
		req.EnvId = fmt.Sprintf("%v", envId.Status)
	}

	return &res, nil
}

func (r *userRepo) UpdateUserToProject(ctx context.Context, req *pb.AddUserToProjectReq) (*pb.AddUserToProjectRes, error) {
	res := pb.AddUserToProjectRes{}

	var (
		clientTypeId, roleId, envId pgtype.UUID
	)
	if req.GetClientTypeId() != "" {
		err := clientTypeId.Set(req.GetClientTypeId())
		if err != nil {
			return nil, err
		}
	} else {
		clientTypeId.Status = pgtype.Null
	}
	if req.GetRoleId() != "" {
		err := roleId.Set(req.GetRoleId())
		if err != nil {
			return nil, err
		}
	} else {
		roleId.Status = pgtype.Null
	}
	if req.GetEnvId() != "" {
		err := envId.Set(req.GetRoleId())
		if err != nil {
			return nil, err
		}
	} else {
		envId.Status = pgtype.Null
	}
	query := `UPDATE user_project 
			  SET client_type_id = $4,
			  role_id = $5
			  WHERE user_id = $1 AND project_id = $2 AND env_id = $3
			  RETURNING user_id, company_id, project_id, client_type_id, role_id, env_id`

	err := r.db.QueryRow(ctx,
		query,
		req.UserId,
		req.ProjectId,
		envId,
		clientTypeId,
		roleId,
	).Scan(
		&res.UserId,
		&res.CompanyId,
		&res.ProjectId,
		&clientTypeId,
		&roleId,
		&envId,
	)
	if err != nil {
		if err.Error() != "no rows in result set" {
			return nil, err
		}
	}
	if roleId.Status != pgtype.Null {
		req.RoleId = fmt.Sprintf("%v", roleId.Status)
	}
	if clientTypeId.Status != pgtype.Null {
		req.ClientTypeId = fmt.Sprintf("%v", clientTypeId.Status)
	}

	return &res, nil
}

func (r *userRepo) GetProjectsByUserId(ctx context.Context, req *pb.GetProjectsByUserIdReq) (*pb.GetProjectsByUserIdRes, error) {
	res := pb.GetProjectsByUserIdRes{}

	query := `SELECT
				project_id,
				env_id
			from user_project
			where user_id = $1`

	rows, err := r.db.Query(ctx, query, req.GetUserId())
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			projectID sql.NullString
			envID     sql.NullString
		)

		err = rows.Scan(&projectID, &envID)
		if err != nil {
			return nil, err
		}

		res.UserProjects = append(res.UserProjects, &pb.UserProject{
			ProjectId: projectID.String,
			EnvId:     envID.String,
		})
	}

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
			filter += " OR login = :login"
		} else {
			filter = "login = :login"
		}
		params["login"] = req.Login
	}
	if req.Phone != "" {
		if filter != "" {
			filter += " OR phone = :phone"
		} else {
			filter = "phone = :phone"
		}
		params["phone"] = req.Phone
	}

	lastQuery, args := helper.ReplaceQueryParams(query+filter, params)
	var userId string
	err := r.db.QueryRow(ctx, lastQuery, args...).Scan(&userId)
	if err == pgx.ErrNoRows {
		return nil, errors.New("not found")
	} else if err != nil {
		return nil, err
	}
	return &pb.GetUserByLoginTypesResponse{
		UserId: userId,
	}, nil
}

func (c *userRepo) GetListLanguage(ctx context.Context, in *pb.GetListSettingReq) (*models.ListLanguage, error) {
	var (
		res models.ListLanguage
	)
	params := make(map[string]interface{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var arr []interface{}
	query := `SELECT
			id,
			name,
			short_name,
			native_name
		FROM
			"language"`
	filter := " "
	offset := " OFFSET 0"
	limit := " LIMIT 10"

	if len(in.GetSearch()) > 0 {
		params["search"] = in.GetSearch()
		filter = " WHERE (((name) ILIKE ('%' || :search || '%'))" +
			" OR ((short_name) ILIKE ('%' || :search || '%'))" +
			" OR ((native_name) ILIKE ('%' || :search || '%')))"
	}

	if in.Offset > 0 {
		params["offset"] = in.Offset
		offset = " OFFSET :offset"
	}

	if in.Limit > 0 {
		params["limit"] = in.Limit
		limit = " LIMIT :limit"
	}

	cQ := `SELECT count(1) FROM "language"` + filter
	cQ, arr = helper.ReplaceQueryParams(cQ, params)
	err := c.db.QueryRow(ctx, cQ, arr...).Scan(
		&res.Count,
	)
	if err != nil {
		return &res, err
	}

	q := query + filter + offset + limit

	q, arr = helper.ReplaceQueryParams(q, params)
	rows, err := c.db.Query(ctx, q, arr...)
	if err != nil {
		return &res, err
	}
	defer rows.Close()

	for rows.Next() {
		var obj models.Language

		err = rows.Scan(
			&obj.Id,
			&obj.Name,
			&obj.ShortName,
			&obj.NativeName,
		)

		if err != nil {
			return &res, err
		}

		res.Language = append(res.Language, &obj)
	}

	return &res, nil
}

func (c *userRepo) GetListTimezone(ctx context.Context, in *pb.GetListSettingReq) (*models.ListTimezone, error) {
	var (
		res models.ListTimezone
	)
	params := make(map[string]interface{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var arr []interface{}
	query := `SELECT
			id,
			"name",
			"text"
		FROM
			"timezone"`
	filter := " "
	offset := " OFFSET 0"
	limit := " LIMIT 10"

	if len(in.GetSearch()) > 0 {
		params["search"] = in.GetSearch()
		filter = " WHERE (((name) ILIKE ('%' || :search || '%'))" +
			"OR ((text) ILIKE ('%' || :search || '%')))"
	}

	if in.Offset > 0 {
		params["offset"] = in.Offset
		offset = " OFFSET :offset"
	}

	if in.Limit > 0 {
		params["limit"] = in.Limit
		limit = " LIMIT :limit"
	}

	cQ := `SELECT count(1) FROM "timezone"` + filter
	cQ, arr = helper.ReplaceQueryParams(cQ, params)
	err := c.db.QueryRow(ctx, cQ, arr...).Scan(
		&res.Count,
	)
	if err != nil {
		return &res, err
	}

	q := query + filter + offset + limit

	q, arr = helper.ReplaceQueryParams(q, params)
	rows, err := c.db.Query(ctx, q, arr...)
	if err != nil {
		return &res, err
	}
	defer rows.Close()

	for rows.Next() {
		var obj models.Timezone

		err = rows.Scan(
			&obj.Id,
			&obj.Name,
			&obj.Text,
		)

		if err != nil {
			return &res, err
		}

		res.Timezone = append(res.Timezone, &obj)
	}

	return &res, nil
}

func (r *userRepo) V2ResetPassword(ctx context.Context, req *pb.V2ResetPasswordRequest) (int64, error) {
	var (
		params                      = make(map[string]interface{})
		subQueryEmail, subQueryPass string
	)
	if req.GetPassword() != "" {
		subQueryPass = "password = :password, "
		params["password"] = req.GetPassword()
	}
	if req.GetEmail() != "" {
		subQueryEmail = "email = :email, "
		params["email"] = req.GetEmail()
	}

	query := `UPDATE "user" SET ` + subQueryPass + subQueryEmail + `
		updated_at = now()
	WHERE
		id = :id`
	params["id"] = req.GetUserId()

	q, arr := helper.ReplaceQueryParams(query, params)

	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected := result.RowsAffected()

	return rowsAffected, err
}

func (c *userRepo) GetUserProjectByAllFields(ctx context.Context, req models.GetUserProjectByAllFieldsReq) (bool, error) {

	var (
		isExists bool
		count    int
	)
	query := `SELECT
			COUNT(1)
		FROM
			"user_project"
		WHERE user_id = $1 AND project_id = $2 AND company_id = $3
		AND client_type_id = $4 AND role_id = $5`
	err := c.db.QueryRow(
		ctx,
		query,
		req.UserId,
		req.ProjectId,
		req.CompanyId,
		req.ClientTypeId,
		req.RoleId,
	).Scan(&count)
	if err != nil {
		return isExists, nil
	}
	if count > 0 {
		isExists = true
	}
	return isExists, nil
}

func (r *userRepo) DeleteUserFromProject(ctx context.Context, req *pb.DeleteSyncUserRequest) (*empty.Empty, error) {

	params := make(map[string]interface{})

	query := `DELETE FROM "user_project" 
	WHERE  
	user_id = :user_id and 
	client_type_id = :client_type_id
	`

	// `DELETE FROM "user_project"
	// 			WHERE
	// 			project_id = :project_id
	// 			AND
	// 			user_id = :user_id
	// 			AND
	// 			company_id = :company_id`

	// params["project_id"] = req.ProjectId
	params["user_id"] = req.UserId
	params["client_type_id"] = req.ClientTypeId

	// params["company_id"] = req.CompanyId
	// if req.GetRoleId() != "" {
	// 	query += " AND role_id = :role_id"
	// 	params["role_id"] = req.GetRoleId()
	// }
	// if req.GetClientTypeId() != "" {
	// 	query += " AND client_type_id = :client_type_id"
	// 	params["client_type_id"] = req.GetClientTypeId()
	// }
	// if req.GetEnvironmentId() != "" {
	// 	query += " AND env_id = :env_id"
	// 	params["env_id"] = req.GetEnvironmentId()
	// }

	q, args := helper.ReplaceQueryParams(query, params)
	_, err := r.db.Exec(ctx,
		q,
		args...,
	)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}
func (r *userRepo) DeleteUsersFromProject(ctx context.Context, req *pb.DeleteManyUserRequest) (*empty.Empty, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	// call function to commit or rollback transaction at the end
	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	query := `DELETE FROM "user_project" 
				WHERE 
					project_id = :project_id AND  
					company_id = :company_id AND
					env_id = :env_id`
	for _, user := range req.GetUsers() {
		params := map[string]interface{}{}
		params["project_id"] = req.GetProjectId()
		params["company_id"] = req.GetCompanyId()
		params["env_id"] = req.GetEnvironmentId()

		if user.UserId != "" {
			query = query + " AND user_id = :user_id"
			params["user_id"] = user.GetUserId()
		} else {
			return nil, errors.New("user id is required")
		}

		if user.GetClientTypeId() != "" {
			query = query + " AND client_type_id = :client_type_id"
			params["client_type_id"] = user.GetClientTypeId()
		}

		if user.GetRoleId() != "" {
			query = query + " AND role_id = :role_id"
			params["role_id"] = user.GetRoleId()
		}

		q, args := helper.ReplaceQueryParams(query, params)
		_, err = r.db.Exec(ctx,
			q,
			args...,
		)
		if err != nil {
			return nil, err
		}
	}

	return &empty.Empty{}, nil
}

func (r *userRepo) GetAllUserProjects(ctx context.Context) ([]string, error) {
	count := 0
	query := `SELECT count(distinct project_id)
	FROM user_project`

	err := r.db.QueryRow(ctx, query).Scan(&count)
	res := make([]string, 0, count)

	query = `SELECT distinct project_id
				FROM user_project`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			project string
		)

		err = rows.Scan(&project)
		if err != nil {
			return nil, err
		}

		res = append(res, project)
	}

	return res, nil
}

func (r *userRepo) UpdateUserProjects(ctx context.Context, envId, projectId string) (*emptypb.Empty, error) {

	query := `UPDATE user_project SET env_id = $1
	  WHERE project_id = $2`

	_, err := r.db.Exec(ctx, query, envId, projectId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (r *userRepo) GetUserEnvProjects(ctx context.Context, userId string) (*models.GetUserEnvProjectRes, error) {
	res := models.GetUserEnvProjectRes{
		EnvProjects: map[string][]string{},
	}

	query := `SELECT project_id,
      			array_agg( DISTINCT env_id)
				FROM user_project
				WHERE user_id = $1
				GROUP BY project_id`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			envIds    []string
			projectId string
		)

		err = rows.Scan(&projectId, pq.Array(&envIds))
		if err != nil {
			return nil, err
		}

		res.EnvProjects[projectId] = envIds
	}

	return &res, nil
}

func IsValidEmailNew(email string) bool {
	// Define the regular expression pattern for a valid email address
	// This is a basic pattern and may not cover all edge cases
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Compile the regular expression
	re := regexp.MustCompile(emailRegex)

	// Use the MatchString method to check if the email matches the pattern
	return re.MatchString(email)
}
