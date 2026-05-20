package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"private-markets-api/internal/apperror"
	service_mocks "private-markets-api/internal/mocks/service"
	"private-markets-api/internal/model"
)

func TestHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fundService := service_mocks.NewMockFundServiceInterface(ctrl)
	investorService := service_mocks.NewMockInvestorServiceInterface(ctrl)
	investmentService := service_mocks.NewMockInvestmentServiceInterface(ctrl)

	handler := NewHandler(fundService, investorService, investmentService, logger)

	t.Run("ListFunds", func(t *testing.T) {
		tests := []struct {
			name           string
			expectedStatus int
			expectedBody   []model.Fund
			expectError    bool
			errorCode      int
		}{
			{
				name:           "success",
				expectedStatus: http.StatusOK,
				expectedBody: []model.Fund{{
					ID:            "f1",
					Name:          "Growth Fund",
					VintageYear:   2025,
					TargetSizeUSD: 1500000,
					Status:        "fundraising",
				}},
			},
			{
				name:           "internal service error",
				expectedStatus: http.StatusInternalServerError,
				expectError:    true,
				errorCode:      apperror.ErrorCodeInternalServerError,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/funds", nil)
				w := httptest.NewRecorder()

				if tt.expectError {
					fundService.EXPECT().List(gomock.Any()).Return(nil, errors.New("boom"))
				} else {
					fundService.EXPECT().List(gomock.Any()).Return(tt.expectedBody, nil)
				}

				handler.ListFunds(w, req)

				resp := w.Result()
				defer resp.Body.Close()

				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}

				if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
					t.Fatalf("expected Content-Type application/json, got %q", ct)
				}

				if tt.expectError {
					var body apperror.ErrorResponse
					if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
						t.Fatalf("failed to decode error response: %v", err)
					}
					if body.ErrorCode != tt.errorCode {
						t.Fatalf("expected error code %d, got %d", tt.errorCode, body.ErrorCode)
					}
					return
				}

				var actual []model.Fund
				if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if len(actual) != len(tt.expectedBody) {
					t.Fatalf("expected %d funds, got %d", len(tt.expectedBody), len(actual))
				}
				if len(actual) > 0 && actual[0].ID != tt.expectedBody[0].ID {
					t.Fatalf("expected fund ID %q, got %q", tt.expectedBody[0].ID, actual[0].ID)
				}
			})
		}
	})

	t.Run("CreateFund", func(t *testing.T) {
		tests := []struct {
			name           string
			body           CreateFundRequest
			expectedStatus int
			expectError    bool
			errorCode      int
			mockFn         func()
		}{
			{
				name: "success",
				body: CreateFundRequest{
					Name:          "New Fund",
					VintageYear:   2025,
					TargetSizeUSD: 5000000,
					Status:        "fundraising",
				},
				expectedStatus: http.StatusCreated,
				mockFn: func() {
					fundService.EXPECT().Create(gomock.Any(), "New Fund", 2025, 5000000.0, "fundraising").
						Return(model.Fund{
							ID:            "f1",
							Name:          "New Fund",
							VintageYear:   2025,
							TargetSizeUSD: 5000000,
							Status:        "fundraising",
						}, nil)
				},
			},
			{
				name:           "missing required fields",
				body:           CreateFundRequest{},
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				errorCode:      apperror.ErrorCodeMissingField,
				mockFn:         func() {},
			},
			{
				name: "invalid vintage year",
				body: CreateFundRequest{
					Name:          "New Fund",
					VintageYear:   1800,
					TargetSizeUSD: 5000000,
					Status:        "fundraising",
				},
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				errorCode:      apperror.ErrorCodeInvalidVintageYear,
				mockFn: func() {
					fundService.EXPECT().Create(gomock.Any(), "New Fund", 1800, 5000000.0, "fundraising").
						Return(model.Fund{}, apperror.ErrInvalidVintageYear)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.mockFn()

				body, _ := json.Marshal(tt.body)
				req := httptest.NewRequest(http.MethodPost, "/funds", bytes.NewReader(body))
				w := httptest.NewRecorder()

				handler.CreateFund(w, req)

				resp := w.Result()
				defer resp.Body.Close()

				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}

				if tt.expectError {
					var errResp apperror.ErrorResponse
					json.NewDecoder(resp.Body).Decode(&errResp)
					if errResp.ErrorCode != tt.errorCode {
						t.Fatalf("expected error code %d, got %d", tt.errorCode, errResp.ErrorCode)
					}
				}
			})
		}
	})

	t.Run("UpdateFund", func(t *testing.T) {
		tests := []struct {
			name           string
			body           UpdateFundRequest
			expectedStatus int
			expectError    bool
			errorCode      int
			mockFn         func()
		}{
			{
				name: "success",
				body: UpdateFundRequest{
					ID:            "f1",
					Name:          "Updated Fund",
					VintageYear:   2025,
					TargetSizeUSD: 6000000,
					Status:        "investing",
				},
				expectedStatus: http.StatusOK,
				mockFn: func() {
					fundService.EXPECT().Update(gomock.Any(), "f1", "Updated Fund", 2025, 6000000.0, "investing").
						Return(model.Fund{
							ID:            "f1",
							Name:          "Updated Fund",
							VintageYear:   2025,
							TargetSizeUSD: 6000000,
							Status:        "investing",
						}, nil)
				},
			},
			{
				name:           "missing required fields",
				body:           UpdateFundRequest{},
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				errorCode:      apperror.ErrorCodeMissingField,
				mockFn:         func() {},
			},
			{
				name: "fund not found",
				body: UpdateFundRequest{
					ID:            "nonexistent",
					Name:          "Fund",
					VintageYear:   2025,
					TargetSizeUSD: 5000000,
					Status:        "fundraising",
				},
				expectedStatus: http.StatusNotFound,
				expectError:    true,
				errorCode:      apperror.ErrorCodeFundNotFound,
				mockFn: func() {
					fundService.EXPECT().Update(gomock.Any(), "nonexistent", "Fund", 2025, 5000000.0, "fundraising").
						Return(model.Fund{}, apperror.ErrFundNotFound)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.mockFn()

				body, _ := json.Marshal(tt.body)
				req := httptest.NewRequest(http.MethodPut, "/funds", bytes.NewReader(body))
				w := httptest.NewRecorder()

				handler.UpdateFund(w, req)

				resp := w.Result()
				defer resp.Body.Close()

				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}

				if tt.expectError {
					var errResp apperror.ErrorResponse
					json.NewDecoder(resp.Body).Decode(&errResp)
					if errResp.ErrorCode != tt.errorCode {
						t.Fatalf("expected error code %d, got %d", tt.errorCode, errResp.ErrorCode)
					}
				}
			})
		}
	})

	t.Run("GetFund", func(t *testing.T) {
		tests := []struct {
			name           string
			fundID         string
			expectedStatus int
			expectError    bool
			mockFn         func()
		}{
			{
				name:           "success",
				fundID:         "f1",
				expectedStatus: http.StatusOK,
				mockFn: func() {
					fundService.EXPECT().Get(gomock.Any(), "f1").
						Return(model.Fund{
							ID:            "f1",
							Name:          "Fund",
							VintageYear:   2025,
							TargetSizeUSD: 5000000,
							Status:        "fundraising",
						}, nil)
				},
			},
			{
				name:           "missing fund id",
				fundID:         "",
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				mockFn:         func() {},
			},
			{
				name:           "fund not found",
				fundID:         "nonexistent",
				expectedStatus: http.StatusNotFound,
				expectError:    true,
				mockFn: func() {
					fundService.EXPECT().Get(gomock.Any(), "nonexistent").
						Return(model.Fund{}, apperror.ErrFundNotFound)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.mockFn()

				req := httptest.NewRequest(http.MethodGet, "/funds/"+tt.fundID, nil)
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("id", tt.fundID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

				w := httptest.NewRecorder()

				handler.GetFund(w, req)

				resp := w.Result()
				defer resp.Body.Close()

				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}
			})
		}
	})

	t.Run("ListInvestors", func(t *testing.T) {
		tests := []struct {
			name           string
			expectedStatus int
			expectError    bool
			mockFn         func()
		}{
			{
				name:           "success",
				expectedStatus: http.StatusOK,
				mockFn: func() {
					investorService.EXPECT().List(gomock.Any()).
						Return([]model.Investor{{
							ID:           "inv1",
							Name:         "John Doe",
							InvestorType: "individual",
							Email:        "john@example.com",
						}}, nil)
				},
			},
			{
				name:           "service error",
				expectedStatus: http.StatusInternalServerError,
				expectError:    true,
				mockFn: func() {
					investorService.EXPECT().List(gomock.Any()).
						Return(nil, errors.New("db error"))
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.mockFn()

				req := httptest.NewRequest(http.MethodGet, "/investors", nil)
				w := httptest.NewRecorder()

				handler.ListInvestors(w, req)

				resp := w.Result()
				defer resp.Body.Close()

				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}
			})
		}
	})

	t.Run("CreateInvestor", func(t *testing.T) {
		tests := []struct {
			name           string
			body           CreateInvestorRequest
			expectedStatus int
			expectError    bool
			errorCode      int
			mockFn         func()
		}{
			{
				name: "success",
				body: CreateInvestorRequest{
					Name:         "Jane Smith",
					InvestorType: "institutional",
					Email:        "jane@example.com",
				},
				expectedStatus: http.StatusCreated,
				mockFn: func() {
					investorService.EXPECT().Create(gomock.Any(), "Jane Smith", "institutional", "jane@example.com").
						Return(model.Investor{
							ID:           "inv1",
							Name:         "Jane Smith",
							InvestorType: "institutional",
							Email:        "jane@example.com",
						}, nil)
				},
			},
			{
				name:           "missing required fields",
				body:           CreateInvestorRequest{},
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				errorCode:      apperror.ErrorCodeMissingField,
				mockFn:         func() {},
			},
			{
				name: "invalid email",
				body: CreateInvestorRequest{
					Name:         "Bob",
					InvestorType: "individual",
					Email:        "invalid-email",
				},
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				errorCode:      apperror.ErrorCodeInvalidEmail,
				mockFn: func() {
					investorService.EXPECT().Create(gomock.Any(), "Bob", "individual", "invalid-email").
						Return(model.Investor{}, apperror.ErrInvalidEmail)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.mockFn()

				body, _ := json.Marshal(tt.body)
				req := httptest.NewRequest(http.MethodPost, "/investors", bytes.NewReader(body))
				w := httptest.NewRecorder()

				handler.CreateInvestor(w, req)

				resp := w.Result()
				defer resp.Body.Close()

				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}

				if tt.expectError {
					var errResp apperror.ErrorResponse
					json.NewDecoder(resp.Body).Decode(&errResp)
					if errResp.ErrorCode != tt.errorCode {
						t.Fatalf("expected error code %d, got %d", tt.errorCode, errResp.ErrorCode)
					}
				}
			})
		}
	})

	t.Run("GetInvestmentsByFundID", func(t *testing.T) {
		tests := []struct {
			name           string
			fundID         string
			expectedStatus int
			expectError    bool
			mockFn         func()
		}{
			{
				name:           "success",
				fundID:         "f1",
				expectedStatus: http.StatusOK,
				mockFn: func() {
					investmentService.EXPECT().GetInvestmentsByFundID(gomock.Any(), "f1").
						Return([]model.Investment{}, nil)
				},
			},
			{
				name:           "missing fund id",
				fundID:         "",
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				mockFn:         func() {},
			},
			{
				name:           "fund not found",
				fundID:         "nonexistent",
				expectedStatus: http.StatusNotFound,
				expectError:    true,
				mockFn: func() {
					investmentService.EXPECT().GetInvestmentsByFundID(gomock.Any(), "nonexistent").
						Return(nil, apperror.ErrFundNotFound)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.mockFn()

				req := httptest.NewRequest(http.MethodGet, "/funds/"+tt.fundID+"/investments", nil)
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("fundID", tt.fundID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

				w := httptest.NewRecorder()

				handler.GetInvestmentsByFundID(w, req)

				resp := w.Result()
				defer resp.Body.Close()

				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}
			})
		}
	})

	t.Run("CreateInvestment", func(t *testing.T) {
		tests := []struct {
			name           string
			fundID         string
			body           CreateInvestmentRequest
			expectedStatus int
			expectError    bool
			errorCode      int
			mockFn         func()
		}{
			{
				name:   "success",
				fundID: "f1",
				body: CreateInvestmentRequest{
					InvestorID:     "inv1",
					AmountUSD:      100000,
					InvestmentDate: "2025-01-15",
				},
				expectedStatus: http.StatusCreated,
				mockFn: func() {
					investmentService.EXPECT().CreateInvestment(gomock.Any(), "f1", "inv1", 100000.0, "2025-01-15").
						Return(model.Investment{
							ID:             "inv_tx1",
							FundID:         "f1",
							InvestorID:     "inv1",
							AmountUSD:      100000,
							InvestmentDate: time.Time{},
						}, nil)
				},
			},
			{
				name:           "missing fund id",
				fundID:         "",
				body:           CreateInvestmentRequest{},
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				errorCode:      apperror.ErrorCodeBadRequest,
				mockFn:         func() {},
			},
			{
				name:           "missing required fields",
				fundID:         "f1",
				body:           CreateInvestmentRequest{},
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				errorCode:      apperror.ErrorCodeMissingField,
				mockFn:         func() {},
			},
			{
				name:   "fund not found",
				fundID: "f1",
				body: CreateInvestmentRequest{
					InvestorID:     "inv1",
					AmountUSD:      100000,
					InvestmentDate: "2025-01-15",
				},
				expectedStatus: http.StatusNotFound,
				expectError:    true,
				errorCode:      apperror.ErrorCodeFundNotFound,
				mockFn: func() {
					investmentService.EXPECT().CreateInvestment(gomock.Any(), "f1", "inv1", 100000.0, "2025-01-15").
						Return(model.Investment{}, apperror.ErrFundNotFound)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.mockFn()

				body, _ := json.Marshal(tt.body)
				req := httptest.NewRequest(http.MethodPost, "/funds/"+tt.fundID+"/investments", bytes.NewReader(body))
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("fundID", tt.fundID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

				w := httptest.NewRecorder()

				handler.CreateInvestment(w, req)

				resp := w.Result()
				defer resp.Body.Close()

				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}

				if tt.expectError {
					var errResp apperror.ErrorResponse
					json.NewDecoder(resp.Body).Decode(&errResp)
					if errResp.ErrorCode != tt.errorCode {
						t.Fatalf("expected error code %d, got %d", tt.errorCode, errResp.ErrorCode)
					}
				}
			})
		}
	})
}
