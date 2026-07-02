package service

import (
	"context"
	"time"

	pbc "ucode/ucode_go_auth_service/genproto/company_service"

	"github.com/google/uuid"
	"github.com/saidamir98/udevs_pkg/logger"
)

const userSeatTransactionType = "user_seat_purchase"

type userSeatCharge struct {
	headProjectID string
	transactionID string
}

func (s *userService) reserveUserSeat(ctx context.Context, project *pbc.Project, creatorID string) (*userSeatCharge, error) {
	if project.GetPerUserPrice() <= 0 {
		return nil, nil
	}

	head, err := s.services.ProjectServiceClient().GetUgenProjectByCompanyId(ctx, &pbc.GetUgenProjectByCompanyIdReq{
		CompanyId: project.GetCompanyId(),
	})
	if err != nil {
		s.log.Error("!!!reserveUserSeat--->GetUgenProjectByCompanyId", logger.Error(err))
		return nil, err
	}

	charge, err := s.services.BillingServiceClient().ChargeProjectBalance(ctx, &pbc.ChargeProjectBalanceRequest{
		ProjectId:       head.GetProjectId(),
		Amount:          project.GetPerUserPrice(),
		CurrencyId:      project.GetPerUserCurrencyId(),
		CreatorId:       creatorID,
		Comment:         "user seat: " + project.GetTitle(),
		TransactionType: userSeatTransactionType,
		ExternalId:      uuid.NewString(),
	})
	if err != nil {
		s.log.Error("!!!reserveUserSeat--->ChargeProjectBalance", logger.Error(err))
		return nil, err
	}

	return &userSeatCharge{
		headProjectID: head.GetProjectId(),
		transactionID: charge.GetTransactionId(),
	}, nil
}

func (s *userService) releaseUserSeat(charge *userSeatCharge) {
	if charge == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if _, err := s.services.BillingServiceClient().RefundProjectBalance(ctx, &pbc.RefundProjectBalanceRequest{
		ProjectId:     charge.headProjectID,
		TransactionId: charge.transactionID,
		Comment:       "refund: user add failed",
	}); err != nil {
		s.log.Error("!!!releaseUserSeat--->RefundProjectBalance failed; manual refund may be required",
			logger.Error(err),
			logger.String("head_project_id", charge.headProjectID),
			logger.String("charge_transaction_id", charge.transactionID),
		)
	}
}
