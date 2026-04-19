package service

import (
	"context"
	"time"

	"private-markets-api/internal/apperror"
	"private-markets-api/internal/model"
	"private-markets-api/internal/repository"

	"github.com/google/uuid"
)

type InvestmentService struct {
	repo repository.Repository
}

func NewInvestmentService(repo repository.Repository) *InvestmentService {
	return &InvestmentService{repo: repo}
}

func (s *InvestmentService) GetInvestmentsByFundID(ctx context.Context, fundID string) ([]model.Investment, error) {
	if fundID == "" {
		return nil, apperror.ErrMissingFundID
	}
	if _, err := uuid.Parse(fundID); err != nil {
		return nil, apperror.ErrFundNotFound
	}
	return s.repo.GetInvestmentsByFundID(ctx, fundID)
}

func (s *InvestmentService) CreateInvestment(ctx context.Context, fundID string, investorID string, amountUSD float64, investmentDate string) (model.Investment, error) {
	if fundID == "" {
		return model.Investment{}, apperror.ErrMissingFundID
	}
	if investorID == "" {
		return model.Investment{}, apperror.ErrMissingInvestorID
	}
	if _, err := uuid.Parse(fundID); err != nil {
		return model.Investment{}, apperror.ErrFundNotFound
	}
	if _, err := uuid.Parse(investorID); err != nil {
		return model.Investment{}, apperror.ErrInvestorNotFound
	}
	if amountUSD <= 0 {
		return model.Investment{}, apperror.ErrInvalidAmountUSD
	}
	if investmentDate == "" {
		return model.Investment{}, apperror.ErrMissingInvestmentDate
	}

	// Validate date format
	parsedDate, err := time.Parse("2006-01-02", investmentDate)
	if err != nil {
		return model.Investment{}, apperror.ErrInvalidInvestmentDate
	}

	// Validate that fund exists
	_, err = s.repo.GetFund(ctx, fundID)
	if err != nil {
		return model.Investment{}, apperror.ErrFundNotFound
	}

	// Validate that investor exists
	_, err = s.repo.GetInvestor(ctx, investorID)
	if err != nil {
		return model.Investment{}, apperror.ErrInvestorNotFound
	}

	investment := model.Investment{
		InvestorID:     investorID,
		FundID:         fundID,
		AmountUSD:      amountUSD,
		InvestmentDate: parsedDate,
	}

	return s.repo.CreateInvestment(ctx, investment)
}
