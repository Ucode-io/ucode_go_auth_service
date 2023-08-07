package postgres

import (
	"context"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type smsOtpSettingsRepo struct {
	db *pgxpool.Pool
}

func NewSmsOtpSettingsRepo(db *pgxpool.Pool) storage.SmsOtpSettingsRepoI {
	return &smsOtpSettingsRepo{
		db: db,
	}
}

func (s *smsOtpSettingsRepo) Create(ctx context.Context, req *pb.CreateSmsOtpSettingsRequest) (*pb.SmsOtpSettings, error) {

	id := uuid.New().String()
	_, err := s.db.Exec(ctx, `
		INSERT INTO 
			"sms_otp_settings" (
				"id",
				"login",
				"password",
				"project_id",
				"environment_id",
				"number_of_otp",
				"default_otp"
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7
			)`,
		id,
		req.Login,
		req.Password,
		req.ProjectId,
		req.EnvironmentId,
		req.NumberOfOtp,
		req.GetDefaultOtp(),
	)
	if err != nil {
		return nil, err
	}
	response := &pb.SmsOtpSettings{
		Id:            id,
		Login:         req.Login,
		Password:      req.Password,
		DefaultOtp:    req.DefaultOtp,
		NumberOfOtp:   req.NumberOfOtp,
		ProjectId:     req.ProjectId,
		EnvironmentId: req.EnvironmentId,
	}
	return response, nil
}
func (s *smsOtpSettingsRepo) Update(ctx context.Context, req *pb.SmsOtpSettings) (int64, error) {

	query := `UPDATE "sms_otp_settings" SET
		updated_at = now(),
		login = :login,
		password = :password,
		number_of_otp = :number_of_otp,
		default_otp = :default_otp
	WHERE
		id = :id`
	params := map[string]interface{}{
		"number_of_otp": req.NumberOfOtp,
		"default_otp":   req.GetDefaultOtp(),
		"login":         req.GetLogin(),
		"password":      req.GetPassword(),
		"id":            req.Id,
	}
	q, arr := helper.ReplaceQueryParams(query, params)
	result, err := s.db.Exec(ctx, q, arr...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), err
}
func (s *smsOtpSettingsRepo) GetById(ctx context.Context, req *pb.SmsOtpSettingsPrimaryKey) (*pb.SmsOtpSettings, error) {
	response := &pb.SmsOtpSettings{}
	err := s.db.QueryRow(ctx, `
		SELECT 
			login, 
			password, 
			project_id, 
			environment_id, 
			number_of_otp,
			default_otp
		FROM "sms_otp_settings"
		WHERE id = $1
		`,
		req.Id,
	).Scan(
		&response.Login,
		&response.Password,
		&response.ProjectId,
		&response.EnvironmentId,
		&response.NumberOfOtp,
		&response.DefaultOtp,
	)
	if err != nil {
		return nil, err
	}
	response.Id = req.Id
	return response, nil
}
func (s *smsOtpSettingsRepo) GetList(ctx context.Context, req *pb.GetListSmsOtpSettingsRequest) (*pb.SmsOtpSettingsResponse, error) {
	res := make([]*pb.SmsOtpSettings, 0)
	rows, err := s.db.Query(ctx, `
		SELECT 
			id,
			login, 
			password, 
			project_id, 
			environment_id, 
			number_of_otp,
			default_otp
		FROM "sms_otp_settings"
		WHERE project_id = $1 AND environment_id = $2
		`,
		req.ProjectId, req.EnvironmentId,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		response := &pb.SmsOtpSettings{}
		err = rows.Scan(
			&response.Id,
			&response.Login,
			&response.Password,
			&response.ProjectId,
			&response.EnvironmentId,
			&response.NumberOfOtp,
			&response.DefaultOtp,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, response)
	}
	return &pb.SmsOtpSettingsResponse{
		Items: res,
	}, nil
}
func (s *smsOtpSettingsRepo) Delete(ctx context.Context, req *pb.SmsOtpSettingsPrimaryKey) (int64, error) {

	queryDeleteFromUserProject := `DELETE FROM "sms_otp_settings" WHERE id = $1`

	result, err := s.db.Exec(ctx, queryDeleteFromUserProject, req.Id)
	if err != nil {
		return 0, err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return 0, errors.New("sms otp settings not found")
	}
	return result.RowsAffected(), nil
}
