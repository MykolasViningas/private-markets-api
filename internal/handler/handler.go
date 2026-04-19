package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"private-markets-api/internal/apperror"
	"private-markets-api/internal/service"

	"github.com/go-chi/chi/v5"
)

type CreateFundRequest struct {
	Name          string  `json:"name"`
	VintageYear   int     `json:"vintage_year"`
	TargetSizeUSD float64 `json:"target_size_usd"`
	Status        string  `json:"status"`
}

type UpdateFundRequest struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	VintageYear   int     `json:"vintage_year"`
	TargetSizeUSD float64 `json:"target_size_usd"`
	Status        string  `json:"status"`
}

type CreateInvestorRequest struct {
	Name         string `json:"name"`
	InvestorType string `json:"investor_type"`
	Email        string `json:"email"`
}

type CreateInvestmentRequest struct {
	InvestorID     string  `json:"investor_id"`
	AmountUSD      float64 `json:"amount_usd"`
	InvestmentDate string  `json:"investment_date"`
}

type Handler struct {
	fundService       *service.FundService
	investorService   *service.InvestorService
	investmentService *service.InvestmentService
	logger            *slog.Logger
}

func NewHandler(fundService *service.FundService, investorService *service.InvestorService, investmentService *service.InvestmentService, logger *slog.Logger) *Handler {
	return &Handler{
		fundService:       fundService,
		investorService:   investorService,
		investmentService: investmentService,
		logger:            logger,
	}
}

func writeErrorResponse(w http.ResponseWriter, status int, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apperror.ErrorResponse{ErrorCode: code, Message: message})
}

func (h *Handler) ListFunds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	funds, err := h.fundService.List(ctx)
	if err != nil {
		h.logger.Error("failed to list funds", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, apperror.ErrorCodeInternalServerError, "internal server error")
		return
	}

	h.logger.Info("funds retrieved", "count", len(funds))

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(funds); err != nil {
		h.logger.Error("failed to encode funds", "error", err)
	}
}

func (h *Handler) CreateFund(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req CreateFundRequest
	if err := decoder.Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeBadRequest, "bad request: invalid payload")
		return
	}

	if req.Name == "" || req.Status == "" || req.VintageYear == 0 || req.TargetSizeUSD == 0 {
		h.logger.Error("missing or empty required fields")
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeMissingField, "required fields are missing or empty")
		return
	}

	ctx := r.Context()
	fund, err := h.fundService.Create(ctx, req.Name, req.VintageYear, req.TargetSizeUSD, req.Status)
	if err != nil {
		h.logger.Error("failed to create fund", "error", err)
		switch {
		case errors.Is(err, apperror.ErrInvalidVintageYear):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidVintageYear, err.Error())
		case errors.Is(err, apperror.ErrInvalidTargetSize):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidTargetSize, err.Error())
		case errors.Is(err, apperror.ErrInvalidStatus):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidStatus, err.Error())
		case errors.Is(err, apperror.ErrFundAlreadyExists):
			writeErrorResponse(w, http.StatusConflict, apperror.ErrorCodeFundAlreadyExists, err.Error())
		default:
			writeErrorResponse(w, http.StatusInternalServerError, apperror.ErrorCodeInternalServerError, "internal server error")
		}
		return
	}

	h.logger.Info("fund created", "id", fund.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(fund); err != nil {
		h.logger.Error("failed to encode fund", "error", err)
	}
}

func (h *Handler) UpdateFund(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req UpdateFundRequest
	if err := decoder.Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeBadRequest, "bad request: invalid payload")
		return
	}

	if req.ID == "" || req.Name == "" || req.Status == "" || req.VintageYear == 0 || req.TargetSizeUSD == 0 {
		h.logger.Error("missing or empty required fields")
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeMissingField, "required fields are missing or empty")
		return
	}

	ctx := r.Context()
	fund, err := h.fundService.Update(ctx, req.ID, req.Name, req.VintageYear, req.TargetSizeUSD, req.Status)
	if err != nil {
		h.logger.Error("failed to update fund", "error", err)
		switch {
		case errors.Is(err, apperror.ErrFundNotFound):
			writeErrorResponse(w, http.StatusNotFound, apperror.ErrorCodeFundNotFound, err.Error())
		case errors.Is(err, apperror.ErrInvalidVintageYear):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidVintageYear, err.Error())
		case errors.Is(err, apperror.ErrInvalidTargetSize):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidTargetSize, err.Error())
		case errors.Is(err, apperror.ErrInvalidStatus):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidStatus, err.Error())
		default:
			writeErrorResponse(w, http.StatusInternalServerError, apperror.ErrorCodeInternalServerError, "internal server error")
		}
		return
	}

	h.logger.Info("fund updated", "id", fund.ID)

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(fund); err != nil {
		h.logger.Error("failed to encode fund", "error", err)
	}
}

func (h *Handler) GetFund(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.logger.Error("missing fund id in path")
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeBadRequest, "missing fund id")
		return
	}

	ctx := r.Context()
	fund, err := h.fundService.Get(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrFundNotFound) {
			writeErrorResponse(w, http.StatusNotFound, apperror.ErrorCodeFundNotFound, "fund not found")
			return
		}
		h.logger.Error("failed to get fund", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, apperror.ErrorCodeInternalServerError, "internal server error")
		return
	}

	h.logger.Info("fund retrieved", "id", fund.ID)

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(fund); err != nil {
		h.logger.Error("failed to encode fund", "error", err)
	}
}

func (h *Handler) ListInvestors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	investors, err := h.investorService.List(ctx)
	if err != nil {
		h.logger.Error("failed to list investors", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, apperror.ErrorCodeInternalServerError, "internal server error")
		return
	}

	h.logger.Info("investors retrieved", "count", len(investors))

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(investors); err != nil {
		h.logger.Error("failed to encode investors", "error", err)
	}
}

func (h *Handler) CreateInvestor(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req CreateInvestorRequest
	if err := decoder.Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeBadRequest, "bad request: invalid payload")
		return
	}

	if req.Name == "" || req.InvestorType == "" || req.Email == "" {
		h.logger.Error("missing or empty required fields")
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeMissingField, "required fields are missing or empty")
		return
	}

	ctx := r.Context()
	investor, err := h.investorService.Create(ctx, req.Name, req.InvestorType, req.Email)
	if err != nil {
		h.logger.Error("failed to create investor", "error", err)
		switch {
		case errors.Is(err, apperror.ErrInvalidInvestorType):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidInvestorType, err.Error())
		case errors.Is(err, apperror.ErrInvestorAlreadyExists):
			writeErrorResponse(w, http.StatusConflict, apperror.ErrorCodeInvestorAlreadyExists, err.Error())
		case errors.Is(err, apperror.ErrInvalidEmail):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidEmail, err.Error())
		case errors.Is(err, apperror.ErrInvalidName):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidName, err.Error())
		default:
			writeErrorResponse(w, http.StatusInternalServerError, apperror.ErrorCodeInternalServerError, "internal server error")
		}
		return
	}

	h.logger.Info("investor created", "id", investor.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(investor); err != nil {
		h.logger.Error("failed to encode investor", "error", err)
	}
}

func (h *Handler) GetInvestmentsByFundID(w http.ResponseWriter, r *http.Request) {
	fundID := chi.URLParam(r, "fundID")
	if fundID == "" {
		h.logger.Error("missing fund_id in path")
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeBadRequest, "missing fund_id")
		return
	}

	ctx := r.Context()
	investments, err := h.investmentService.GetInvestmentsByFundID(ctx, fundID)
	if err != nil {
		h.logger.Error("failed to get investments", "error", err)
		switch {
		case errors.Is(err, apperror.ErrMissingFundID):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeMissingField, err.Error())
		case errors.Is(err, apperror.ErrFundNotFound):
			writeErrorResponse(w, http.StatusNotFound, apperror.ErrorCodeFundNotFound, err.Error())
		default:
			writeErrorResponse(w, http.StatusInternalServerError, apperror.ErrorCodeInternalServerError, "internal server error")
		}
		return
	}

	h.logger.Info("investments retrieved", "fundID", fundID, "count", len(investments))

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(investments); err != nil {
		h.logger.Error("failed to encode investments", "error", err)
	}
}

func (h *Handler) CreateInvestment(w http.ResponseWriter, r *http.Request) {
	fundID := chi.URLParam(r, "fundID")
	if fundID == "" {
		h.logger.Error("missing fund_id in path")
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeBadRequest, "missing fund_id")
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req CreateInvestmentRequest
	if err := decoder.Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeBadRequest, "bad request: invalid payload")
		return
	}

	if req.InvestorID == "" || req.AmountUSD == 0 || req.InvestmentDate == "" {
		h.logger.Error("missing or empty required fields")
		writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeMissingField, "required fields are missing or empty")
		return
	}

	ctx := r.Context()
	investment, err := h.investmentService.CreateInvestment(ctx, fundID, req.InvestorID, req.AmountUSD, req.InvestmentDate)
	if err != nil {
		h.logger.Error("failed to create investment", "error", err)
		switch {
		case errors.Is(err, apperror.ErrFundNotFound):
			writeErrorResponse(w, http.StatusNotFound, apperror.ErrorCodeFundNotFound, err.Error())
		case errors.Is(err, apperror.ErrInvalidAmountUSD):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidAmountUSD, err.Error())
		case errors.Is(err, apperror.ErrMissingFundID):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeMissingField, err.Error())
		case errors.Is(err, apperror.ErrMissingInvestmentDate):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeMissingInvestmentDate, err.Error())
		case errors.Is(err, apperror.ErrInvestorNotFound):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvestorNotFound, err.Error())
		case errors.Is(err, apperror.ErrMissingInvestorID):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeMissingInvestorID, err.Error())
		case errors.Is(err, apperror.ErrInvalidInvestmentDate):
			writeErrorResponse(w, http.StatusBadRequest, apperror.ErrorCodeInvalidDateFormat, err.Error())
		default:
			writeErrorResponse(w, http.StatusInternalServerError, apperror.ErrorCodeInternalServerError, "internal server error")
		}
		return
	}

	h.logger.Info("investment created", "id", investment.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(investment); err != nil {
		h.logger.Error("failed to encode investment", "error", err)
	}
}
