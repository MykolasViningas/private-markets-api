package service

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"private-markets-api/internal/apperror"
	"private-markets-api/internal/model"
	"private-markets-api/internal/repository"
)

type InvestorService struct {
	repo repository.Repository
}

func NewInvestorService(repo repository.Repository) *InvestorService {
	return &InvestorService{repo: repo}
}

var validTypes = []string{"individual", "institutional", "family office"}

func (s *InvestorService) List(ctx context.Context) ([]model.Investor, error) {
	return s.repo.ListInvestors(ctx)
}

func (s *InvestorService) Create(ctx context.Context, name string, investorType string, email string) (model.Investor, error) {
	if strings.TrimSpace(name) == "" {
		return model.Investor{}, apperror.ErrInvalidName
	}
	if !isValidInvestorType(investorType) {
		return model.Investor{}, apperror.ErrInvalidInvestorType
	}
	if !isValidEmail(email) {
		return model.Investor{}, apperror.ErrInvalidEmail
	}

	investorType = strings.Replace(investorType, " ", "_", 1) // Replace space with underscore for storage

	investor := model.Investor{
		Name:         name,
		InvestorType: investorType,
		Email:        email,
	}

	return s.repo.CreateInvestor(ctx, investor)
}

func (s *InvestorService) Get(ctx context.Context, id string) (model.Investor, error) {
	if id == "" {
		return model.Investor{}, apperror.ErrMissingID
	}
	return s.repo.GetInvestor(ctx, id)
}

func isValidInvestorType(investorType string) bool {
	return slices.Contains(validTypes, strings.ToLower(investorType))
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
