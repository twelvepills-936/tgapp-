package service

import (
	"context"
	"errors"
	"fmt"

	"gitlab16.skiftrade.kz/templates/go/errorcodes"
	"gitlab16.skiftrade.kz/templates/go/internal"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
	api "gitlab16.skiftrade.kz/templates/go/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewService constructs gRPC handlers bound to use cases.
func NewService(uc internal.UseCase) *service {
	return &service{
		uc: uc,
	}
}

type service struct {
	api.UnimplementedCyberMateServer
	uc internal.UseCase
}

func (s *service) RegisterByTelegram(ctx context.Context, req *api.RegisterByTelegramRequest) (*api.RegisterByTelegramResponse, error) {
	out, err := s.uc.RegisterByTelegram(ctx, ucModels.RegisterByTelegramInput{
		InitDataRaw: req.GetInitDataRaw(),
		StartParam:  req.GetStartParam(),
	})
	if err != nil {
		switch {
		case errors.Is(err, ucModels.ErrProfileAlreadyRegistered):
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("%s: %s", errorcodes.ProfileAlreadyRegistered, ucModels.ErrProfileAlreadyRegistered.Error()))
		case errors.Is(err, ucModels.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("%s: %s", errorcodes.InvalidArgument, err.Error()))
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("%s: %s", errorcodes.Internal, ucModels.ErrInternalServerError.Error()))
		}
	}
	return &api.RegisterByTelegramResponse{ProfileId: out.ProfileID}, nil
}

func (s *service) GetUserByTelegramId(ctx context.Context, req *api.GetUserByTelegramIdRequest) (*api.GetUserResponse, error) {
	out, err := s.uc.GetUserByTelegramID(ctx, req.GetTelegramId())
	if err != nil {
		return nil, mapProfileLookupError(err)
	}
	return &api.GetUserResponse{Data: &api.User{
		Id:      out.Data.ID,
		Name:    out.Data.Name,
		Surname: out.Data.Username,
	}}, nil
}

func (s *service) GetWalletByTelegramId(ctx context.Context, req *api.GetWalletByTelegramIdRequest) (*api.GetWalletResponse, error) {
	out, err := s.uc.GetWalletByTelegramID(ctx, req.GetTelegramId())
	if err != nil {
		return nil, mapProfileLookupError(err)
	}

	transactions := make([]*api.WalletTransactionItem, 0, len(out.Transactions))
	for _, item := range out.Transactions {
		transactions = append(transactions, &api.WalletTransactionItem{
			Id:          item.ID,
			Date:        item.Date,
			Type:        item.Type,
			Amount:      item.Amount,
			Status:      item.Status,
			Description: item.Description,
		})
	}

	return &api.GetWalletResponse{
		Wallet: &api.WalletData{
			Id:               out.Wallet.ID,
			ProfileId:        out.Wallet.ProfileID,
			Balance:          out.Wallet.Balance,
			TotalEarned:      out.Wallet.TotalEarned,
			BalanceAvailable: out.Wallet.BalanceAvailable,
		},
		Transactions: transactions,
	}, nil
}

func (s *service) GetReferralsByTelegramId(ctx context.Context, req *api.GetReferralsByTelegramIdRequest) (*api.GetReferralsResponse, error) {
	out, err := s.uc.GetReferralsByTelegramID(ctx, req.GetTelegramId())
	if err != nil {
		return nil, mapProfileLookupError(err)
	}

	items := make([]*api.ReferralItem, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, &api.ReferralItem{
			Id:                  item.ID,
			TelegramId:          item.TelegramID,
			Name:                item.Name,
			Username:            item.Username,
			CompletedTasksCount: item.CompletedTasksCount,
			Earnings:            item.Earnings,
		})
	}

	return &api.GetReferralsResponse{Items: items}, nil
}

func (s *service) CreatePromptHistory(ctx context.Context, req *api.CreatePromptHistoryRequest) (*api.CreatePromptHistoryResponse, error) {
	out, err := s.uc.SavePromptHistory(ctx, ucModels.SavePromptHistoryInput{
		TelegramID: req.GetTelegramId(),
		Prompt:     req.GetPrompt(),
		Category:   req.GetCategory(),
	})
	if err != nil {
		switch {
		case errors.Is(err, ucModels.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("%s: %s", errorcodes.InvalidArgument, err.Error()))
		default:
			return nil, mapProfileLookupError(err)
		}
	}

	return &api.CreatePromptHistoryResponse{Item: &api.PromptHistoryItem{
		Id:        out.Item.ID,
		Prompt:    out.Item.Prompt,
		Category:  out.Item.Category,
		CreatedAt: out.Item.CreatedAt,
	}}, nil
}

func (s *service) GetPromptHistoryByTelegramId(ctx context.Context, req *api.GetPromptHistoryByTelegramIdRequest) (*api.GetPromptHistoryResponse, error) {
	out, err := s.uc.GetPromptHistoryByTelegramID(ctx, req.GetTelegramId())
	if err != nil {
		return nil, mapProfileLookupError(err)
	}

	items := make([]*api.PromptHistoryItem, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, &api.PromptHistoryItem{
			Id:        item.ID,
			Prompt:    item.Prompt,
			Category:  item.Category,
			CreatedAt: item.CreatedAt,
		})
	}

	return &api.GetPromptHistoryResponse{Items: items}, nil
}

func mapProfileLookupError(err error) error {
	switch {
	case errors.Is(err, ucModels.ErrProfileNotFound):
		return status.Error(codes.NotFound, fmt.Sprintf("%s: %s", errorcodes.ProfileNotFound, ucModels.ErrProfileNotFound.Error()))
	case errors.Is(err, ucModels.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%s: %s", errorcodes.InvalidArgument, err.Error()))
	default:
		return status.Error(codes.Internal, fmt.Sprintf("%s: %s", errorcodes.Internal, ucModels.ErrInternalServerError.Error()))
	}
}
