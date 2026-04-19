package repository

import (
	"context"
	"errors"

	"private-markets-api/internal/apperror"
	"private-markets-api/internal/model"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	// Fund methods
	ListFunds(ctx context.Context) ([]model.Fund, error)
	CreateFund(ctx context.Context, fund model.Fund) (model.Fund, error)
	UpdateFund(ctx context.Context, fund model.Fund) (model.Fund, error)
	GetFund(ctx context.Context, id string) (model.Fund, error)
	// Investor methods
	ListInvestors(ctx context.Context) ([]model.Investor, error)
	CreateInvestor(ctx context.Context, investor model.Investor) (model.Investor, error)
	GetInvestor(ctx context.Context, id string) (model.Investor, error)
	// Investment methods
	GetInvestmentsByFundID(ctx context.Context, fundID string) ([]model.Investment, error)
	CreateInvestment(ctx context.Context, investment model.Investment) (model.Investment, error)
}

type Repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// Fund methods
func (r *Repo) ListFunds(ctx context.Context) ([]model.Fund, error) {
	query := `SELECT f.id,
		f.name,
		f.vintage_year,
		f.target_size_usd,
		fs.name as status,
		f.created_at,
		f.updated_at
	FROM funds f
	JOIN fund_statuses fs ON f.status_id = fs.id
	ORDER BY f.created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var funds []model.Fund
	for rows.Next() {
		var f model.Fund
		err := rows.Scan(&f.ID, &f.Name, &f.VintageYear, &f.TargetSizeUSD, &f.Status, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		funds = append(funds, f)
	}
	return funds, rows.Err()
}

func (r *Repo) CreateFund(ctx context.Context, fund model.Fund) (model.Fund, error) {
	query := `INSERT INTO funds (name, vintage_year, target_size_usd, status_id)
		SELECT $1, $2, $3, id
		FROM fund_statuses
		WHERE name = $4
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, fund.Name, fund.VintageYear, fund.TargetSizeUSD, fund.Status).Scan(&fund.ID, &fund.CreatedAt, &fund.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return model.Fund{}, apperror.ErrFundAlreadyExists
		}
	}
	return fund, err
}

func (r *Repo) UpdateFund(ctx context.Context, fund model.Fund) (model.Fund, error) {
	query := `UPDATE funds
		SET name = $1,
			vintage_year = $2,
			target_size_usd = $3,
			status_id = (SELECT id FROM fund_statuses WHERE name = $4),
			updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at`
	err := r.db.QueryRow(ctx, query, fund.Name, fund.VintageYear, fund.TargetSizeUSD, fund.Status, fund.ID).Scan(&fund.UpdatedAt)
	if err != nil {
		return model.Fund{}, err
	}
	return fund, nil
}

func (r *Repo) GetFund(ctx context.Context, id string) (model.Fund, error) {
	query := `SELECT f.id,
		f.name,
		f.vintage_year,
		f.target_size_usd,
		fs.name as status,
		f.created_at,
		f.updated_at
	FROM funds f
	JOIN fund_statuses fs ON f.status_id = fs.id
	WHERE f.id = $1`
	var f model.Fund
	err := r.db.QueryRow(ctx, query, id).Scan(&f.ID, &f.Name, &f.VintageYear, &f.TargetSizeUSD, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

// Investor methods
func (r *Repo) ListInvestors(ctx context.Context) ([]model.Investor, error) {
	query := `SELECT i.id,
		i.name,
		it.name as investor_type,
		i.email,
		i.created_at,
		i.updated_at
	FROM investors i
	JOIN investor_types it ON i.investor_type_id = it.id
	ORDER BY i.created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investors []model.Investor
	for rows.Next() {
		var i model.Investor
		err := rows.Scan(&i.ID, &i.Name, &i.InvestorType, &i.Email, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return nil, err
		}
		investors = append(investors, i)
	}
	return investors, rows.Err()
}

func (r *Repo) CreateInvestor(ctx context.Context, investor model.Investor) (model.Investor, error) {
	query := `INSERT INTO investors (name, investor_type_id, email)
		SELECT $1, id, $3
		FROM investor_types
		WHERE name = $2
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, investor.Name, investor.InvestorType, investor.Email).Scan(&investor.ID, &investor.CreatedAt, &investor.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Investor{}, apperror.ErrInvalidInvestorType
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return model.Investor{}, apperror.ErrInvestorAlreadyExists
		}
	}
	return investor, err
}

func (r *Repo) GetInvestor(ctx context.Context, id string) (model.Investor, error) {
	query := `SELECT i.id,
		i.name,
		it.name as investor_type,
		i.email,
		i.created_at,
		i.updated_at
	FROM investors i
	JOIN investor_types it ON i.investor_type_id = it.id
	WHERE i.id = $1`
	var i model.Investor
	err := r.db.QueryRow(ctx, query, id).Scan(&i.ID, &i.Name, &i.InvestorType, &i.Email, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

// Investment methods
func (r *Repo) GetInvestmentsByFundID(ctx context.Context, fundID string) ([]model.Investment, error) {
	query := `SELECT id,
		investor_id,
		fund_id,
		amount_usd,
		investment_date,
		created_at,
		updated_at
	FROM investments
	WHERE fund_id = $1
	ORDER BY investment_date DESC`
	rows, err := r.db.Query(ctx, query, fundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investments []model.Investment
	for rows.Next() {
		var inv model.Investment
		err := rows.Scan(&inv.ID, &inv.InvestorID, &inv.FundID, &inv.AmountUSD, &inv.InvestmentDate, &inv.CreatedAt, &inv.UpdatedAt)
		if err != nil {
			return nil, err
		}
		investments = append(investments, inv)
	}
	return investments, rows.Err()
}

func (r *Repo) CreateInvestment(ctx context.Context, investment model.Investment) (model.Investment, error) {
	query := `INSERT INTO investments (fund_id, investor_id, amount_usd, investment_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, investment.FundID, investment.InvestorID, investment.AmountUSD, investment.InvestmentDate).Scan(&investment.ID, &investment.CreatedAt, &investment.UpdatedAt)
	return investment, err
}
