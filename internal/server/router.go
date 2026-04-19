package server

import (
	"log/slog"
	"net/http"

	"private-markets-api/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *handler.Handler, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RedirectSlashes)
	r.Use(requestLogger(logger))

	r.Route("/api/v1/private-markets", func(r chi.Router) {
		r.Route("/funds", func(r chi.Router) {
			r.Get("/", h.ListFunds)
			r.Post("/", h.CreateFund)
			r.Put("/", h.UpdateFund)
			r.Get("/{id}", h.GetFund)
			r.Get("/{fundID}/investments", h.GetInvestmentsByFundID)
			r.Post("/{fundID}/investments", h.CreateInvestment)
		})

		r.Route("/investors", func(r chi.Router) {
			r.Get("/", h.ListInvestors)
			r.Post("/", h.CreateInvestor)
		})
	})

	return r
}
