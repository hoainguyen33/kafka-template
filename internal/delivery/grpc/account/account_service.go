package account

import (
	"context"
	"kafka-test/internal/models"
	"kafka-test/pkg/logger"

	accountErrors "kafka-test/constant/errors/account_errors"
	grpcErrors "kafka-test/constant/errors/grpc_errors"
	"kafka-test/constant/grpc"
	accountsService "kafka-test/proto/account"

	"github.com/go-playground/validator/v10"
	"github.com/opentracing/opentracing-go"
)

// accountService gRPC Service
type accountService struct {
	accountsService.UnimplementedAccountsServiceServer
	log      logger.Logger
	validate *validator.Validate
}

// NewAccountService accountService constructor
func NewAccountService(log logger.Logger, validate *validator.Validate) *accountService {
	return &accountService{log: log, validate: validate}
}

// Login account
func (p *accountService) Login(ctx context.Context, in *accountsService.LoginReq) (*accountsService.LoginRes, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "accountService.Create")
	defer span.Finish()
	grpc.CreateMessages.Inc()

	acc := &models.Account{
		Username: in.GetUsername(),
		Password: in.GetPassword(),
	}
	// test password and username
	if acc.Username != "test" || acc.Password != "12345678" {
		grpc.ErrorMessages.Inc()
		p.log.Errorf("accountUC.Create: %v", grpcErrors.ErrAccountInvalid)
		ctx.Err()
		return nil, grpcErrors.ErrorResponse(grpcErrors.ErrAccountInvalid, accountErrors.AccountInvalid)
	}
	ctx.Done()
	grpc.SuccessMessages.Inc()
	return &accountsService.LoginRes{Token: acc.ToTokenProto()}, nil
}
