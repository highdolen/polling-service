package handler

import (
	"net/http"

	"polling-service/internal/middleware"
)

func NewRouter(
	publicHandler *PublicHandler,
	adminHandler *AdminHandler,
	adminToken string,
	rateLimiter *middleware.RateLimiter,
) http.Handler {
	mux := http.NewServeMux()

	voteHandler := rateLimiter.Middleware(
		http.HandlerFunc(publicHandler.Vote),
	)

	mux.HandleFunc(
		"GET /api/v1/polls/active",
		publicHandler.GetActivePoll,
	)

	mux.HandleFunc(
		"GET /api/v1/polls/{id}",
		publicHandler.GetPoll,
	)

	mux.Handle(
		"POST /api/v1/polls/{id}/vote",
		voteHandler,
	)

	mux.Handle(
		"POST /api/v1/admin/polls",
		middleware.AdminAuth(adminToken)(
			http.HandlerFunc(adminHandler.CreatePoll),
		),
	)

	mux.Handle(
		"GET /api/v1/admin/polls",
		middleware.AdminAuth(adminToken)(
			http.HandlerFunc(adminHandler.ListPolls),
		),
	)

	mux.Handle(
		"GET /api/v1/admin/polls/{id}/results",
		middleware.AdminAuth(adminToken)(
			http.HandlerFunc(adminHandler.GetResults),
		),
	)

	return middleware.CORS(mux)
}
