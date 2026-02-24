// Package handler ...
package handler

import (
	authv1 "gateway/proto/auth/v1"
	"net/http"
	"time"
)

// AuthHandler ...
type AuthHandler struct {
	client authv1.AuthServiceClient
}

// NewAuthHandler ...
func NewAuthHandler(client authv1.AuthServiceClient) *AuthHandler {
	return &AuthHandler{client: client}
}

// Register POST /auth/register
// Body: { "email": "...", "password": "..." }
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	resp, err := h.client.Register(r.Context(), &authv1.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user_id": resp.GetUserId(),
	})
}

// Login POST /auth/login
// Body: { "email": "...", "password": "...", "app_id": 1 }
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		AppID    int32  `json:"app_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if req.AppID == 0 {
		req.AppID = 1
	}

	resp, err := h.client.Login(r.Context(), &authv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		AppId:    req.AppID,
	})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":       resp.GetAccessToken(),
		"refresh_token":      resp.GetRefreshToken(),
		"access_expires_at":  timeOrNil(resp.GetAccessExpiresAt().AsTime()),
		"refresh_expires_at": timeOrNil(resp.GetRefreshExpiresAt().AsTime()),
	})
}

// Logout POST /auth/logout
// Body: { "refresh_token": "..." }
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	resp, err := h.client.Logout(r.Context(), &authv1.LogoutRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": resp.GetSuccess(),
	})
}

// Refresh POST /auth/refresh
// Body: { "refresh_token": "..." }
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	resp, err := h.client.RefreshToken(r.Context(), &authv1.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":       resp.GetAccessToken(),
		"refresh_token":      resp.GetRefreshToken(),
		"access_expires_at":  timeOrNil(resp.GetAccessExpiresAt().AsTime()),
		"refresh_expires_at": timeOrNil(resp.GetRefreshExpiresAt().AsTime()),
	})
}

// IsAdmin GET /auth/is-admin?user_id=42
func (h *AuthHandler) IsAdmin(w http.ResponseWriter, r *http.Request) {
	userID, err := queryInt64(r, "user_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "user_id query param required")
		return
	}

	resp, err := h.client.IsAdmin(r.Context(), &authv1.IsAdminRequest{
		UserId: userID,
	})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"is_admin": resp.GetIsAdmin(),
	})
}

func timeOrNil(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
