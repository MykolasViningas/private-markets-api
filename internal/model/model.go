package model

import "time"

type Fund struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	VintageYear   int       `json:"vintage_year"`
	TargetSizeUSD float64   `json:"target_size_usd"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Investor struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	InvestorType string    `json:"investor_type"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Investment struct {
	ID             string    `json:"id"`
	InvestorID     string    `json:"investor_id"`
	FundID         string    `json:"fund_id"`
	AmountUSD      float64   `json:"amount_usd"`
	InvestmentDate time.Time `json:"investment_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
