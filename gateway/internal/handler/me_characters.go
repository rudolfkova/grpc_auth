package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	characterv1 "character/proto/character/v1"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultCharacterListLimit = 50
	maxCharacterListLimit     = 100
)

type accessJWTClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// MeCharactersHandler — список персонажей текущего пользователя через character-service (доверенный gRPC).
type MeCharactersHandler struct {
	charClient characterv1.CharacterServiceClient
	jwtSecret  string
	charToken  string // metadata x-service-token к character-service; может быть пустым
}

// NewMeCharactersHandler ...
func NewMeCharactersHandler(cli characterv1.CharacterServiceClient, jwtSecret, characterServiceToken string) *MeCharactersHandler {
	return &MeCharactersHandler{
		charClient: cli,
		jwtSecret:  strings.TrimSpace(jwtSecret),
		charToken:  strings.TrimSpace(characterServiceToken),
	}
}

// ListMyCharacters GET /api/me/characters — JWT в Authorization: Bearer; без data в теле ответа.
func (h *MeCharactersHandler) ListMyCharacters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeProblem(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	if h.charClient == nil {
		writeProblem(w, http.StatusServiceUnavailable, "character_service_unconfigured", "character_service_addr is not set on gateway")
		return
	}

	raw := strings.TrimSpace(r.Header.Get("Authorization"))
	if raw == "" {
		writeProblem(w, http.StatusUnauthorized, "unauthorized", "missing Authorization header")
		return
	}
	if len(raw) > 7 && strings.EqualFold(raw[:7], "bearer ") {
		raw = strings.TrimSpace(raw[7:])
	}
	if raw == "" {
		writeProblem(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
		return
	}

	userID, err := userIDFromAccessJWT(raw, h.jwtSecret)
	if err != nil || userID == 0 {
		writeProblem(w, http.StatusUnauthorized, "unauthorized", "invalid or expired access token")
		return
	}

	limit, offset := parseListCharactersPagination(r)

	ctx := r.Context()
	if h.charToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", h.charToken)
	}

	resp, err := h.charClient.ListCharacters(ctx, &characterv1.ListCharactersRequest{
		UserId: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		httpSt, code := characterListUpstream(err)
		writeProblem(w, httpSt, code, grpcMessage(err))
		return
	}

	out := make([]map[string]any, 0, len(resp.GetCharacters()))
	for _, c := range resp.GetCharacters() {
		if c == nil {
			continue
		}
		out = append(out, map[string]any{
			"id":             c.GetId(),
			"display_name":   c.GetDisplayName(),
			"description":    c.GetDescription(),
			"schema_version": c.GetSchemaVersion(),
			"version":        c.GetVersion(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"characters": out})
}

func userIDFromAccessJWT(rawToken, secret string) (int64, error) {
	c := &accessJWTClaims{}
	tok, err := jwt.ParseWithClaims(rawToken, c, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return 0, err
	}
	if c.UserID == 0 {
		return 0, errors.New("missing user_id in token")
	}
	return c.UserID, nil
}

func parseListCharactersPagination(r *http.Request) (limit, offset int) {
	limit = defaultCharacterListLimit
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxCharacterListLimit {
		limit = maxCharacterListLimit
	}
	if v := strings.TrimSpace(r.URL.Query().Get("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

func characterListUpstream(err error) (httpStatus int, code string) {
	s, ok := status.FromError(err)
	if !ok {
		return http.StatusBadGateway, "character_upstream_error"
	}
	switch s.Code() {
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "character_service_unavailable"
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, "character_service_timeout"
	case codes.InvalidArgument:
		return http.StatusBadRequest, "character_service_invalid_argument"
	default:
		return http.StatusBadGateway, "character_service_error"
	}
}
