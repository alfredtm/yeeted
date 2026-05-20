package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/alfredtm/yeeted/internal/metrics"
	"github.com/alfredtm/yeeted/internal/model"
)

func NewRouter(pool *pgxpool.Pool) http.Handler {
	_ = pool

	store := model.NewStore()
	items := NewItemsHandler(store)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(metrics.Middleware)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", Health)
	r.Method(http.MethodGet, "/metrics", metrics.Handler())

	r.Get("/items", items.List)
	r.Post("/items", items.Create)
	r.Get("/items/{id}", items.Get)
	r.Delete("/items/{id}", items.Delete)

	return otelhttp.NewHandler(r, "http.server")
}
