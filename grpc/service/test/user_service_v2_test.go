package service_test

import (
	"context"
	"testing"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"

	"github.com/stretchr/testify/assert"
)

var userId string

func TestV2CreateUser(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	user := &auth_service.CreateUserRequest{
		Login:                 fakeData.UserName(),
		Password:              fakeData.Characters(5),
		Email:                 fakeData.Email(),
		Phone:                 fakeData.PhoneNumber(),
		CompanyId:             config.CreadentialsForTest[conf.Environment]["companyId"],
		ResourceType:          1,
		ResourceEnvironmentId: config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
		ProjectId:             config.CreadentialsForTest[conf.Environment]["projectId"],
		ClientTypeId:          config.CreadentialsForTest[conf.Environment]["clientTypeId"],
		RoleId:                config.CreadentialsForTest[conf.Environment]["roleId"],
		Active:                1,
	}

	t.Log("This is a log message during the test.", user)

	res, err := svcs.UserService().V2CreateUser(context.Background(), user)

	userId = res.Id
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestV2GetUserByID(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if userId != "" {
		res, err := svcs.UserService().V2GetUserByID(context.Background(), &auth_service.UserPrimaryKey{
			Id:        userId,
			ProjectId: config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
			// ResourceEnvironmentId: config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
			ResourceType: 1,
			// ClientTypeId:          config.CreadentialsForTest[conf.Environment]["clientTypeId"],
			// CompanyId:             config.CreadentialsForTest[conf.Environment]["companyId"],
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	}
}

func TestV2GetUserList(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	res, err := svcs.UserService().V2GetUserList(context.Background(), &auth_service.GetUserListRequest{
		Limit:                 10,
		ProjectId:             config.CreadentialsForTest[conf.Environment]["projectId"],
		ResourceEnvironmentId: config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
		ResourceType:          1,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestV2UpdateUser(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	user := &auth_service.UpdateUserRequest{
		Id:           userId,
		Login:        fakeData.UserName(),
		Email:        fakeData.Email(),
		Phone:        fakeData.PhoneNumber(),
		CompanyId:    config.CreadentialsForTest[conf.Environment]["companyId"],
		ResourceType: 1,
		ProjectId:    config.CreadentialsForTest[conf.Environment]["projectId"],
	}

	res, err := svcs.UserService().V2UpdateUser(context.Background(), user)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestGetProjectsByUserId(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if userId != "" {
		res, err := svcs.UserService().GetProjectsByUserId(context.Background(), &auth_service.GetProjectsByUserIdReq{
			UserId: userId,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	}
}

func TestV2GetUserByLoginTypes(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if userId != "" {
		res, err := svcs.UserService().GetUserByID(context.Background(), &auth_service.UserPrimaryKey{
			Id:           userId,
			ProjectId:    config.CreadentialsForTest[conf.Environment]["projectId"],
			ResourceType: 1,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		resp, err := svcs.UserService().V2GetUserByLoginTypes(context.Background(), &auth_service.GetUserByLoginTypesRequest{
			Email:        res.GetEmail(),
			Phone:        res.GetPhone(),
			Login:        res.GetLogin(),
			ResourceType: 1,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	}
}

func TestGetUserProjects(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if userId != "" {
		resp, err := svcs.UserService().GetUserProjects(context.Background(), &auth_service.UserPrimaryKey{
			Id: userId,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	}
}

func TestGetUserByUsername(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if userId != "" {
		res, err := svcs.UserService().GetUserByID(context.Background(), &auth_service.UserPrimaryKey{
			Id:           userId,
			ProjectId:    config.CreadentialsForTest[conf.Environment]["projectId"],
			ResourceType: 1,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		resp, err := svcs.UserService().GetUserByUsername(context.Background(), &auth_service.GetUserByUsernameRequest{
			Username: res.GetEmail(),
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	}
}

func TestV2DeleteUser(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if userId != "" {
		res, err := svcs.UserService().V2DeleteUser(context.Background(),
			&auth_service.UserPrimaryKey{
				Id: userId,
				// ResourceEnvironmentId: config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
				ResourceType: 1,
				// ClientTypeId:          config.CreadentialsForTest[conf.Environment]["clientTypeId"],
				// CompanyId:             config.CreadentialsForTest[conf.Environment]["companyId"],
				// ProjectId:             config.CreadentialsForTest[conf.Environment]["projectId"],
				IsTest: true,
			},
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	}
}
