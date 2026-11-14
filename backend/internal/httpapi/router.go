package httpapi

import (
	"net/http"

	"unesarent/backend/internal/auth"
	"unesarent/backend/internal/config"
)

func NewRouter(cfg config.Config, authHandler *AuthHandler, tokens *auth.TokenIssuer) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", Health)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.Handle("GET /api/auth/me", RequireAuth(tokens)(http.HandlerFunc(authHandler.Me)))

	return CORS(cfg.CORSOrigins)(Recover(Logger(mux)))
}
