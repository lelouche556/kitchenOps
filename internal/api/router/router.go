package router

import (
	"net/http"

	"SwishAssignment/internal/api/controllers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(kitchen *controllers.KitchenController) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", kitchen.Health)
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/orders/confirm", kitchen.ConfirmOrder)
		api.Post("/allocator/run-once", kitchen.AllocateOnce)
		api.Post("/tasks/{taskID}/start", kitchen.StartTask)
		api.Post("/tasks/{taskID}/complete", kitchen.CompleteTask)
		api.Get("/tasks/{taskID}", kitchen.GetTask)
	})

	return r
}
