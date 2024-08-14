package postgres

import (
	"context"
	"encoding/json"
	"log"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/storage"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type loginPlatformTypeRepo struct {
	db *pgxpool.Pool
}

func NewLoginPlatformTypeRepo(db *pgxpool.Pool) storage.LoginPlatformType {
	return &loginPlatformTypeRepo{
		db: db,
	}
}

func (e *loginPlatformTypeRepo) CreateLoginPlatformType(ctx context.Context, input *pb.LoginPlatform) (*pb.LoginPlatform, error) {

	var data map[string]string
	if input.Type == "APPLE" {
		data = map[string]string{
			"team_id":   input.Data.TeamId,
			"client_id": input.Data.ClientId,
			"key_id":    input.Data.KeyId,
			"secret":    input.Data.Secret,
		}
	} else if input.Type == "GOOGLE" {
		data = map[string]string{
			"email":    input.Data.Email,
			"password": input.Data.Password,
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	query := `INSERT INTO "login_platform_setting" (
		id,
		project_id,
		env_id,
		type,
		data
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5::JSONB
	)`

	_, err = e.db.Exec(ctx, query,
		input.Id,
		input.ProjectId,
		input.EnvironmentId,
		input.Type,
		jsonData,
	)

	if err != nil {

		return nil, err
	}

	return input, nil
}

func (e *loginPlatformTypeRepo) GetLoginPlatformType(ctx context.Context, pKey *pb.LoginPlatformTypePrimaryKey) (res *pb.LoginPlatform, err error) {

	query := `SELECT id, project_id, env_id, type, data
	FROM "login_platform_setting"
	WHERE id=$1;`

	res = &pb.LoginPlatform{}

	var resp []byte
	data := map[string]string{}

	err = e.db.QueryRow(ctx, query, pKey.Id).Scan(
		&res.Id,
		&res.ProjectId,
		&res.EnvironmentId,
		&res.Type,
		&resp,
	)
	if err == pgx.ErrNoRows {
		err := errors.New("login platform type not found")
		return nil, err
	} else if err != nil {
		return res, err
	}

	err = json.Unmarshal(resp, &data)
	if err != nil {
		return res, err
	}

	res.Data = &pb.LoginPlatformType{
		TeamId:   data["team_id"],
		ClientId: data["client_id"],
		KeyId:    data["key_id"],
		Secret:   data["secret"],
		Email:    data["email"],
		Password: data["password"],
	}

	return res, nil
}

func (e *loginPlatformTypeRepo) UpdateLoginPlatformType(ctx context.Context, input *pb.UpdateLoginPlatformTypeRequest, types string) (string, error) {

	var resp = &pb.LoginPlatform{}

	var data map[string]string
	if types == "APPLE" {
		data = map[string]string{
			"team_id":   input.Data.TeamId,
			"client_id": input.Data.ClientId,
			"key_id":    input.Data.KeyId,
			"secret":    input.Data.Secret,
		}
	} else if types == "GOOGLE" {
		data = map[string]string{
			"email":    input.Data.Email,
			"password": input.Data.Password,
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	query := `UPDATE "login_platform_setting" 
	SET	data = $1
	WHERE id = $2
	RETURNING id`

	err = e.db.QueryRow(ctx, query,
		jsonData,
		input.Id,
	).Scan(
		&resp.Id,
	)
	if err != nil {
		return "", err
	}

	return resp.Id, nil
}

func (e *loginPlatformTypeRepo) GetListLoginPlatformType(ctx context.Context, input *pb.GetListLoginPlatformTypeRequest) (*pb.GetListLoginPlatformTypeResponse, error) {

	arr := &pb.GetListLoginPlatformTypeResponse{}

	query := `SELECT
		id,
		project_id,
		type,
		data
	FROM
		"login_platform_setting"
	WHERE
	env_id = $1`

	rows, err := e.db.Query(ctx, query, input.EnvironmentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		obj := &pb.LoginPlatform{}
		var resp []byte
		data := map[string]string{}

		err = rows.Scan(
			&obj.Id,
			&obj.ProjectId,
			&obj.Type,
			&resp,
		)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(resp, &data)
		if err != nil {
			log.Fatal(err)
		}

		objdata := pb.LoginPlatformType{
			TeamId:   data["team_id"],
			ClientId: data["client_id"],
			KeyId:    data["key_id"],
			Secret:   data["secret"],
			Email:    data["email"],
			Password: data["password"],
		}
		obj.Data = &objdata

		if err != nil {
			return nil, err
		}

		arr.Items = append(arr.Items, obj)
	}

	return arr, nil
}

func (e *loginPlatformTypeRepo) DeleteLoginSettings(ctx context.Context, input *pb.LoginPlatformTypePrimaryKey) (*emptypb.Empty, error) {

	var resp = &emptypb.Empty{}

	query := `DELETE FROM "login_platform_setting" WHERE id = $1`

	_, err := e.db.Query(ctx, query, input.Id)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
