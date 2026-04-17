package gameapp

import (
	"context"
	"time"

	"game/internal/domain/ports"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"google.golang.org/grpc/status"
)

// PrepareCharacterJoin выставляет стартовую сетку/HP/stats из character.data (первый заход user_id в этот инстанс).
func (s *Service) PrepareCharacterJoin(userID int64, characterData []byte) {
	d := gamekit.ParseCharacterPlayData(characterData)
	s.engine.EnsurePlayerJoin(userID, d)
}

func (s *Service) CharacterSessionsEnabled() bool {
	return s.characterSessions != nil
}

func (s *Service) ResolveCharacterJoin(ctx context.Context, userID int64, characterID string) (*ports.ResolveCharacterResponse, error) {
	if s.characterSessions == nil {
		return nil, nil
	}
	started := time.Now()
	resp, err := s.characterSessions.ResolvePlayCharacter(ctx, userID, characterID)
	if err != nil {
		s.telemetry.ObserveGRPCClientCall("character", "resolve_play_character", status.Code(err).String(), time.Since(started))
		return nil, err
	}
	s.telemetry.ObserveGRPCClientCall("character", "resolve_play_character", "OK", time.Since(started))
	return resp, nil
}

// PersistCharacterPlaySession пишет character.data в character-service после выхода последнего WS этого user_id.
func (s *Service) PersistCharacterPlaySession(ctx context.Context, userID int64, persisted bool, ch ports.CharacterIdentity) error {
	if s.characterSessions == nil {
		return nil
	}
	raw, err := s.engine.PlayerCharacterData(userID)
	if err != nil {
		return err
	}
	started := time.Now()
	err = s.characterSessions.SavePlaySession(ctx, ports.SaveCharacterSessionRequest{
		UserID:        userID,
		Persisted:     persisted,
		Character:     ch,
		Data:          raw,
		SchemaVersion: gamekit.CharacterDataSchemaVersion,
	})
	if err != nil {
		s.telemetry.ObserveGRPCClientCall("character", "save_play_session", status.Code(err).String(), time.Since(started))
		return err
	}
	s.telemetry.ObserveGRPCClientCall("character", "save_play_session", "OK", time.Since(started))
	return nil
}
