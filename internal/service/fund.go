package service

import (
	"context"
	"slices"
	"strings"

	"private-markets-api/internal/apperror"
	"private-markets-api/internal/model"
	"private-markets-api/internal/repository"

	"github.com/google/uuid"
)

type FundService struct {
	repo repository.Repository
}

func NewFundService(repo repository.Repository) *FundService {
	return &FundService{repo: repo}
}

var validStatuses = []string{"fundraising", "investing", "closed"}

func (s *FundService) List(ctx context.Context) ([]model.Fund, error) {
	return s.repo.ListFunds(ctx)
}

func (s *FundService) Create(ctx context.Context, name string, vintageYear int, targetSizeUSD float64, status string) (model.Fund, error) {
	if vintageYear < 1900 || vintageYear > 2100 {
		return model.Fund{}, apperror.ErrInvalidVintageYear
	}
	if targetSizeUSD <= 0 {
		return model.Fund{}, apperror.ErrInvalidTargetSize
	}
	if !isValidStatus(status) {
		return model.Fund{}, apperror.ErrInvalidStatus
	}

	fund := model.Fund{
		Name:          name,
		VintageYear:   vintageYear,
		TargetSizeUSD: targetSizeUSD,
		Status:        status,
	}

	return s.repo.CreateFund(ctx, fund)
}

func (s *FundService) Update(ctx context.Context, id string, name string, vintageYear int, targetSizeUSD float64, status string) (model.Fund, error) {
	if id == "" {
		return model.Fund{}, apperror.ErrMissingID
	}
	if _, err := uuid.Parse(id); err != nil {
		return model.Fund{}, apperror.ErrFundNotFound
	}
	if vintageYear < 1900 || vintageYear > 2100 {
		return model.Fund{}, apperror.ErrInvalidVintageYear
	}
	if targetSizeUSD <= 0 {
		return model.Fund{}, apperror.ErrInvalidTargetSize
	}
	if !isValidStatus(status) {
		return model.Fund{}, apperror.ErrInvalidStatus
	}

	fund := model.Fund{
		ID:            id,
		Name:          name,
		VintageYear:   vintageYear,
		TargetSizeUSD: targetSizeUSD,
		Status:        status,
	}

	return s.repo.UpdateFund(ctx, fund)
}

func (s *FundService) Get(ctx context.Context, id string) (model.Fund, error) {
	if id == "" {
		return model.Fund{}, apperror.ErrMissingID
	}
	if _, err := uuid.Parse(id); err != nil {
		return model.Fund{}, apperror.ErrFundNotFound
	}
	return s.repo.GetFund(ctx, id)
}

func isValidStatus(status string) bool {
	return slices.Contains(validStatuses, strings.ToLower(status))
}
