package gameapp

import (
	"context"
	"strings"

	"game/internal/infrastructure/characterclient"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// PrepareCharacterJoin выставляет стартовую сетку/HP/stats из character.data (первый заход user_id в этот инстанс).
func (s *Service) PrepareCharacterJoin(userID int64, characterData []byte) {
	d := gamekit.ParseCharacterPlayData(characterData)
	s.engine.EnsurePlayerJoin(userID, d)
}

// PersistCharacterPlaySession пишет character.data в character-service после выхода последнего WS этого user_id.
func (s *Service) PersistCharacterPlaySession(ctx context.Context, addr, token string, userID int64, persisted bool, ch characterclient.PlayCharacter) error {
	if strings.TrimSpace(addr) == "" {
		return nil
	}
	raw, err := s.engine.PlayerCharacterData(userID)
	if err != nil {
		return err
	}
	return characterclient.SavePlaySession(ctx, addr, token, userID, persisted, ch, raw, gamekit.CharacterDataSchemaVersion)
}
