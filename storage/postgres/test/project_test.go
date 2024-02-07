package postgres_test

// import (
// 	"context"
// 	"testing"
// 	"time"
// 	"ucode/ucode_go_auth_service/config"
// 	"ucode/ucode_go_auth_service/genproto/auth_service"
// 	"ucode/ucode_go_auth_service/storage"
// 	"ucode/ucode_go_auth_service/storage/postgres"
// )

// type project struct {
// 	t *testing.T
// 	s storage.StorageI
// }

// func TestProject(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
// 	defer cancel()

// 	cfg := config.Load()
// 	t.Log(cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDatabase)
// 	strg, err := postgres.NewPostgres(ctx, cfg)
// 	if err != nil {
// 		t.Log(err)
// 		t.FailNow()
// 	}

// 	comPKey, err := strg.Company().Register(ctx, &auth_service.RegisterCompanyRequest{Name: "medion"})
// 	if err != nil {
// 		t.Log(err)
// 		t.FailNow()
// 	}

// 	c := project{
// 		t: t,
// 		s: strg,
// 	}
// 	pKey, err := c.create(ctx, comPKey.Id)
// 	if err != nil {
// 		t.Log(err)
// 		t.FailNow()
// 	}

// 	c.getByPK(ctx, pKey.Id)
// 	c.getList(ctx)
// 	c.update(ctx, pKey.Id)
// 	c.delete(ctx, pKey.Id)

// }

// func (c project) create(ctx context.Context, companyId string) (pKey *auth_service.ProjectPrimaryKey, err error) {
// 	pKey, err = c.s.Project().Create(ctx, &auth_service.CreateProjectRequest{
// 		CompanyId: companyId,
// 		Name:      "Ucode",
// 		Domain:    "test-admin.u-code.io",
// 	})

// 	if err != nil {
// 		c.t.Log("create", err)
// 		c.t.Fail()
// 	}

// 	return pKey, err
// }

// func (c project) getByPK(ctx context.Context, id string) (res *auth_service.Project, err error) {
// 	res, err = c.s.Project().GetByPK(ctx, &auth_service.ProjectPrimaryKey{
// 		Id: id,
// 	})

// 	if err != nil {
// 		c.t.Log("getByPK", err)
// 		c.t.Fail()
// 	}

// 	return res, err
// }

// func (c project) getList(ctx context.Context) (res *auth_service.GetProjectListResponse, err error) {
// 	res, err = c.s.Project().GetList(ctx, &auth_service.GetProjectListRequest{
// 		Offset: 0,
// 		Limit:  11,
// 	})

// 	if err != nil {
// 		c.t.Log("getList", err)
// 		c.t.Fail()
// 	}

// 	return res, err
// }

// func (c project) update(ctx context.Context, id string) (rowsAffected int64, err error) {
// 	rowsAffected, err = c.s.Project().Update(ctx, &auth_service.UpdateProjectRequest{
// 		Id:     id,
// 		Name:   "Ucode-updated",
// 		Domain: "test-admin.u-code.io.updated",
// 	})

// 	if err != nil {
// 		c.t.Log("update", err)
// 		c.t.Fail()
// 	}

// 	return rowsAffected, err
// }

// func (c project) delete(ctx context.Context, id string) (rowsAffected int64, err error) {
// 	rowsAffected, err = c.s.Project().Delete(ctx, &auth_service.ProjectPrimaryKey{
// 		Id: id,
// 	})

// 	if err != nil {
// 		c.t.Log("delete", err)
// 		c.t.Fail()
// 	}

// 	return rowsAffected, err
// }
