package apperror

import "errors"

// HTTP Error codes
const (
	ErrorCodeBadRequest            = 1000
	ErrorCodeMissingField          = 1001
	ErrorCodeInvalidVintageYear    = 1002
	ErrorCodeInvalidTargetSize     = 1003
	ErrorCodeInvalidStatus         = 1004
	ErrorCodeFundNotFound          = 1005
	ErrorCodeInvalidInvestorType   = 1006
	ErrorCodeInvalidAmountUSD      = 1007
	ErrorCodeInvalidDateFormat     = 1008
	ErrorCodeInvestorAlreadyExists = 1009
	ErrorCodeFundAlreadyExists     = 1010
	ErrorCodeMissingInvestmentDate = 1011
	ErrorCodeInvestorNotFound      = 1012
	ErrorCodeMissingInvestorID     = 1013
	ErrorCodeMissingFundID         = 1014
	ErrorCodeMissingID             = 1015
	ErrorCodeInvalidName           = 1016
	ErrorCodeInvalidEmail          = 1017
	ErrorCodeInternalServerError   = 2000
)

// Service layer errors
var (
	ErrInvalidVintageYear    = errors.New("invalid vintage_year")
	ErrInvalidTargetSize     = errors.New("target_size_usd must be positive")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrMissingID             = errors.New("id is required")
	ErrMissingFundID         = errors.New("fund_id is required")
	ErrInvalidName           = errors.New("invalid name")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidAmountUSD      = errors.New("amount_usd must be positive")
	ErrInvalidInvestorType   = errors.New("invalid investor type")
	ErrFundNotFound          = errors.New("fund not found")
	ErrInvestorAlreadyExists = errors.New("investor already exists")
	ErrFundAlreadyExists     = errors.New("fund already exists")
	ErrMissingInvestmentDate = errors.New("investment_date is required")
	ErrInvestorNotFound      = errors.New("investor not found")
	ErrMissingInvestorID     = errors.New("investor_id is required")
	ErrInvalidInvestmentDate = errors.New("invalid date format, expected YYYY-MM-DD")
)

type ErrorResponse struct {
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}
