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

var clientTypeId string

func TestV2CreateClientType(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	clientType := &auth_service.V2CreateClientTypeRequest{
		Name:         fakeData.Name(),
		SelfRegister: false,
		SelfRecover:  false,
		ResourceType: 1,
		ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
		DbProjectId:  config.CreadentialsForTest[conf.Environment]["projectId"],
	}

	t.Log("This is a log message during the client test.", clientType)

	res, err := svcs.ClientService().V2CreateClientType(context.Background(), clientType)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)

	data := res.Data.AsMap()["data"].(map[string]interface{})

	if bytes, err := json.Marshal(data); err == nil {
		t.Log("data ======>", string(bytes))
	}

	clientTypeId = data["guid"].(string)
}

func TestV2GetClientTypeByID(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if clientTypeId != "" {
		res, err := svcs.ClientService().V2GetClientTypeByID(context.Background(), &auth_service.V2ClientTypePrimaryKey{
			Id:           clientTypeId,
			ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
			ResourceType: 1,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	}
}

func TestV2GetClientTypeList(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	res, err := svcs.ClientService().V2GetClientTypeList(context.Background(), &auth_service.V2GetClientTypeListRequest{
		Limit:        10,
		ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
		ResourceType: 1,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestV2UpdateClientType(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	clientType := &auth_service.V2UpdateClientTypeRequest{
		Guid:         clientTypeId,
		Name:         fakeData.Name(),
		SelfRegister: true,
		SelfRecover:  false,
		ResourceType: 1,
		ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
		DbProjectId:  config.CreadentialsForTest[conf.Environment]["projectId"],
	}

	res, err := svcs.ClientService().V2UpdateClientType(context.Background(), clientType)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestV2DeleteClientType(t *testing.T) {
	svcs, err := client.NewGrpcClients(conf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if clientTypeId != "" {
		res, err := svcs.ClientService().V2DeleteClientType(context.Background(),
			&auth_service.V2ClientTypePrimaryKey{
				Id:           clientTypeId,
				ProjectId:    config.CreadentialsForTest[conf.Environment]["resourceEnvironmentId"],
				ResourceType: 1,
			},
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	}
}
