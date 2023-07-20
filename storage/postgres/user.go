package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/saidamir98/udevs_pkg/util"

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
		u.password,
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
		&res.Password,
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
		-- name = :name,
		company_id = :company_id,
		-- photo_url = :photo_url,
		phone = :phone,
		email = :email,
		login = :login,
		updated_at = now(),
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
	log.Println("language_id", entity.LanguageId, "timezone_id", entity.TimezoneId)
	q, arr := helper.ReplaceQueryParams(query, params)
	log.Println("query", q, "arr", arr)
	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected = result.RowsAffected()

	return rowsAffected, err
}

func (r *userRepo) Delete(ctx context.Context, pKey *pb.UserPrimaryKey) (int64, error) {

	// return 0, nil

	// queryDeleteFromUserProject := `DELETE FROM user_project WHERE user_id = $1`

	// result, err := r.db.Exec(ctx, queryDeleteFromUserProject, pKey.Id)
	// if err != nil {
	// 	return 0, err
	// }
	// rowsAffected = result.RowsAffected()
	// if rowsAffected == 0 {
	// 	return 0, errors.New("user not found")
	// }

	// result, err = r.db.Exec(ctx, `DELETE FROM "user" WHERE id = $1`, pKey.GetId())
	// if err != nil {
	// 	return 0, errors.Wrap(err, "delete user error")
	// }
	// rowsAffected = result.RowsAffected()
	// if rowsAffected == 0 {
	// 	return 0, errors.New("user not found")
	// }

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

	if util.IsValidEmail(username) {
		query = query + ` email = $1`
	} else if util.IsValidPhone(username) {
		query = query + ` phone = $1`
	} else {
		query = query + ` login = $1`
	}

	err = r.db.QueryRow(ctx, query, username).Scan(
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

	var (
		clientTypeId, roleId pgtype.UUID
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

	query := `INSERT INTO
			user_project(user_id, company_id, project_id, client_type_id, role_id)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING user_id, company_id, project_id, client_type_id, role_id`

	err := r.db.QueryRow(ctx,
		query,
		req.GetUserId(),
		req.GetCompanyId(),
		req.GetProjectId(),
		clientTypeId,
		roleId,
	).Scan(
		&res.UserId,
		&res.CompanyId,
		&res.ProjectId,
		&clientTypeId,
		&roleId,
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
			filter += " OR login = :login"
		} else {
			filter = "login = :login"
		}
		params["login"] = req.Login
	}
	if req.Phone != "" {
		if filter != "" {
			filter += " OR phone = :login"
		} else {
			filter = "phone = :phone"
		}
		params["phone"] = req.Phone
	}
	fmt.Println("params: ", params)

	lastQuery, args := helper.ReplaceQueryParams(query+filter, params)
	fmt.Println("query: ", lastQuery, args)
	var userId string
	err := r.db.QueryRow(ctx, lastQuery, args...).Scan(&userId)
	if err == pgx.ErrNoRows {
		return nil, errors.New("not found")
	} else if err != nil {
		return nil, err
	}
	fmt.Println("user_id: ", userId)
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

	query := `DELETE FROM "user_project" 
				WHERE 
				project_id = $1 AND 
				user_id = $2 AND 
				role_id = $3 AND 
				client_type_id = $4 AND company_id = $5`

	_, err := r.db.Exec(ctx,
		query,
		req.GetProjectId(),
		req.GetUserId(),
		req.GetRoleId(),
		req.GetClientTypeId(),
		req.GetCompanyId(),
	)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (r *userRepo) V2ResetPassword(ctx context.Context, req *pb.V2ResetPasswordRequest) (int64, error) {
	var (
		subQuery string
		params   = make(map[string]interface{})
	)
	if req.GetEmail() == "" {
		subQuery = ""
	} else {
		subQuery = ` email = :email, `
	}
	query := `UPDATE "user" SET
		password = :password,` + subQuery + `
		updated_at = now()
	WHERE
		id = :id`
	if subQuery != "" {
		params["email"] = req.GetEmail()
	}
	params["id"] = req.GetUserId()
	params["password"] = req.GetPassword()

	fmt.Println("query: ", query)
	fmt.Print("\n\nParams: ", params)

	q, arr := helper.ReplaceQueryParams(query, params)
	fmt.Println("q: ", q)
	result, err := r.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	rowsAffected := result.RowsAffected()

	return rowsAffected, err
}
