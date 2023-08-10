package service_test

import (
	"context"
	"encoding/json"
	"testing"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"

	"github.com/stretchr/testify/assert"
)

var (
	roleId string
)

func TestAddRole(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	role := &auth_service.V2AddRoleRequest{
		ClientTypeId: config.CreadentialsForTest[conf.Environment]["clientTypeId"],
		Name:         fakeData.Name(),
		ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
		ResourceType: 1,
		GrantAccess:  false,
	}
	res, err := svcs.PermissionService().V2AddRole(context.Background(), role)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)

	userData := res.Data.AsMap()["data"].(map[string]interface{})

	if bytes, err := json.Marshal(userData); err == nil {
		t.Log("userData", string(bytes))
	}

	t.Log("userData[response.guid].(string) ==========", userData["guid"].(string))

	roleId = userData["guid"].(string)
}
func TestV2GetRoleById(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if roleId != "" {
		res, err := svcs.PermissionService().V2GetRoleById(context.Background(), &auth_service.V2RolePrimaryKey{
			Id:           roleId,
			ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
			ResourceType: 1,
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	}
}
func TestV2GetRolesList(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	res, err := svcs.PermissionService().V2GetRolesList(context.Background(), &auth_service.V2GetRolesListRequest{
		Limit:        10,
		ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
		ClientTypeId: config.CreadentialsForTest[conf.Environment]["clientTypeId"],
		ResourceType: 1,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestV2UpdateRole(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	role := &auth_service.V2UpdateRoleRequest{
		Guid:         roleId,
		ClientTypeId: config.CreadentialsForTest[conf.Environment]["clientTypeId"],
		Name:         fakeData.Name(),
		ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
		ResourceType: 1,
		GrantAccess:  false,
	}

	res, err := svcs.PermissionService().V2UpdateRole(context.Background(), role)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestV2RemoveRole(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	res, err := svcs.PermissionService().V2RemoveRole(context.Background(),
		&auth_service.V2RolePrimaryKey{
			Id:           roleId,
			ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
			ResourceType: 1,
		},
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
