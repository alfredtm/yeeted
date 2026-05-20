package handler

import (
	"net/http"

	"github.com/alfredtm/yeeted/internal/model"
)

func NewRouter() http.Handler {
	store := model.NewStore()
	items := NewItemsHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", Health)
	mux.HandleFunc("GET /items", items.List)
	mux.HandleFunc("POST /items", items.Create)
	mux.HandleFunc("GET /items/{id}", items.Get)
	mux.HandleFunc("DELETE /items/{id}", items.Delete)
	return mux
}
