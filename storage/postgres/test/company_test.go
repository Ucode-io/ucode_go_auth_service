package postgres_test

import (
	"context"
	"testing"
	"time"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/storage"
	"ucode/ucode_go_auth_service/storage/postgres"
)

type company struct {
	t *testing.T
	s storage.StorageI
}

func TestCompany(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cfg := config.BaseLoad()
	t.Log(cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDatabase)
	strg, err := postgres.NewPostgres(ctx, cfg)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	c := company{
		t: t,
		s: strg,
	}
	pKey, err := c.register(ctx)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	c.update(ctx, pKey.Id)
	c.remove(ctx, pKey.Id)

}

func (c company) register(ctx context.Context) (pKey *auth_service.CompanyPrimaryKey, err error) {
	pKey, err = c.s.Company().Register(ctx, &auth_service.RegisterCompanyRequest{
		Name: "Ucode",
	})

	if err != nil {
		c.t.Log("register", err)
		c.t.Fail()
	}

	return pKey, err
}

func (c company) update(ctx context.Context, id string) (rowsAffected int64, err error) {
	rowsAffected, err = c.s.Company().Update(ctx, &auth_service.UpdateCompanyRequest{
		Id:   id,
		Name: "Ucode-updated",
	})

	if err != nil {
		c.t.Log("update", err)
		c.t.Fail()
	}

	return rowsAffected, err
}

func (c company) remove(ctx context.Context, id string) (rowsAffected int64, err error) {
	rowsAffected, err = c.s.Company().Remove(ctx, &auth_service.CompanyPrimaryKey{
		Id: id,
	})

	if err != nil {
		c.t.Log("remove", err)
		c.t.Fail()
	}

	return rowsAffected, err
}
