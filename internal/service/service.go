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
		switch {
		case errors.Is(err, ucModels.ErrProfileNotFound):
			return nil, status.Error(codes.NotFound, fmt.Sprintf("%s: %s", errorcodes.ProfileNotFound, ucModels.ErrProfileNotFound.Error()))
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("%s: %s", errorcodes.Internal, ucModels.ErrInternalServerError.Error()))
	}
	return &api.GetUserResponse{Data: &api.User{
		Id:      out.Data.ID,
		Name:    out.Data.Name,
		Surname: out.Data.Username,
	}}, nil
}
