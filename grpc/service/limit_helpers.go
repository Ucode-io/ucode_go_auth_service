package service

import (
	"context"

	"ucode/ucode_go_auth_service/config"
	pbc "ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func checkUserProjectLimit(ctx context.Context, services client.ServiceManagerI, strg storage.StorageI, fareId, companyId string) error {
	if len(fareId) == 0 {
		return nil
	}

	count, err := strg.User().GetCompanyUsersCount(ctx, companyId)
	if err != nil {
		return status.Error(codes.Internal, "error getting users count")
	}

	limitResp, err := services.BillingServiceClient().CompareFunction(ctx, &pbc.CompareFunctionRequest{
		Type:   config.FARE_USERS,
		FareId: fareId,
		Count:  count + 1,
	})
	if err != nil {
		return status.Error(codes.Internal, "error checking user limit")
	}

	if !limitResp.HasAccess {
		return status.Error(codes.ResourceExhausted, "you have reached the user limit on your current plan. Please upgrade to add more users.")
	}

	return nil
}

func checkUgenBuildersLimit(ctx context.Context, services client.ServiceManagerI, strg storage.StorageI, fareId, projectId string) error {
	if len(fareId) == 0 {
		return nil
	}

	count, err := strg.User().GetCompanyUsersCount(ctx, projectId)
	if err != nil {
		return status.Error(codes.Internal, "error getting users count")
	}

	limitResp, err := services.BillingServiceClient().CompareFunction(ctx, &pbc.CompareFunctionRequest{
		Type:   config.FARE_USERS,
		FareId: fareId,
		Count:  count + 1,
	})
	if err != nil {
		return status.Error(codes.Internal, "error checking user limit")
	}

	if !limitResp.HasAccess {
		return status.Error(codes.ResourceExhausted, "you have reached the builder user limit on your current plan. Please upgrade to add more builder users.")
	}

	return nil
}
