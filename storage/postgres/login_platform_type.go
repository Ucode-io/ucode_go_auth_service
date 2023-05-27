package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/storage"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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

func (e *loginPlatformTypeRepo) CreateLogin(ctx context.Context, input *pb.LoginPlatform) (*pb.LoginPlatform, error) {
	data := map[string]string{
		"team_id":   input.Data.TeamId,
		"client_id": input.Data.ClientId,
		"key_id":    input.Data.KeyId,
		"secret":    input.Data.Secret,
		"email":     input.Data.Email,
		"password":  input.Data.Password,
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
		fmt.Println("Error executing query ", err)
		return nil, err
	}

	return input, nil
}

func (e *loginPlatformTypeRepo) GetLoginBysPK(ctx context.Context, pKey *pb.LoginPlatformTypePrimaryKey) (res *pb.LoginPlatform, err error) {
	res = &pb.LoginPlatform{}
	query := `SELECT id, data
	FROM "login_platform_setting"
	WHERE id=$1;`

	var resp []byte
	data := map[string]string{}

	err = e.db.QueryRow(ctx, query, pKey.Id).Scan(
		&res.Id,
		&resp,
	)

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(resp, &data)
	if err != nil {
		log.Fatal(err)
	}

	res.Data.TeamId = data["team_id"]
	res.Data.ClientId = data["client_id"]
	res.Data.KeyId = data["key_id"]
	res.Data.Secret = data["secret"]
	res.Data.Email = data["email"]
	res.Data.Password = data["password"]

	if err == pgx.ErrNoRows {
		err := errors.New("login settings not found")
		return nil, err
	} else if err != nil {
		return res, err
	}

	return res, nil
}

func (e *loginPlatformTypeRepo) UpdateLoginPlatformType(ctx context.Context, input *pb.LoginPlatform) (string, error) {

	var resp = &pb.LoginPlatform{}

	data := map[string]string{
		"team_id":   input.Data.TeamId,
		"client_id": input.Data.ClientId,
		"key_id":    input.Data.KeyId,
		"secret":    input.Data.Secret,
		"email":     input.Data.Email,
		"password":  input.Data.Password,
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
		project_id = $1`

	rows, err := e.db.Query(ctx, query, input.ProjectId)
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

		obj.Data.TeamId = data["team_id"]
		obj.Data.ClientId = data["client_id"]
		obj.Data.KeyId = data["key_id"]
		obj.Data.Secret = data["secret"]
		obj.Data.Email = data["email"]
		obj.Data.Password = data["password"]

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
