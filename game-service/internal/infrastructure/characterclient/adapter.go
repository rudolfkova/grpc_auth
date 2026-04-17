package characterclient

import (
	"context"
	"fmt"
	"strings"

	"game/internal/domain/ports"
)

// SessionsAdapter — реализация ports.CharacterSessions через gRPC character-service.
type SessionsAdapter struct {
	addr  string
	token string
}

func NewSessionsAdapter(addr, token string) *SessionsAdapter {
	return &SessionsAdapter{
		addr:  strings.TrimSpace(addr),
		token: token,
	}
}

func (a *SessionsAdapter) ResolvePlayCharacter(ctx context.Context, userID int64, characterID string) (*ports.ResolveCharacterResponse, error) {
	resp, err := ResolvePlayCharacter(ctx, a.addr, a.token, userID, characterID)
	if err != nil {
		return nil, err
	}
	ch := resp.GetCharacter()
	if ch == nil || strings.TrimSpace(ch.GetId()) == "" {
		return nil, fmt.Errorf("character resolve returned empty character")
	}

	return &ports.ResolveCharacterResponse{
		Persisted: resp.GetPersisted(),
		Character: ports.CharacterIdentity{
			ID:          ch.GetId(),
			DisplayName: ch.GetDisplayName(),
			Description: ch.GetDescription(),
		},
		Data: ch.GetData(),
	}, nil
}

func (a *SessionsAdapter) SavePlaySession(ctx context.Context, req ports.SaveCharacterSessionRequest) error {
	return SavePlaySession(ctx, a.addr, a.token, req.UserID, req.Persisted, PlayCharacter{
		ID:          req.Character.ID,
		DisplayName: req.Character.DisplayName,
		Description: req.Character.Description,
	}, req.Data, req.SchemaVersion)
}
