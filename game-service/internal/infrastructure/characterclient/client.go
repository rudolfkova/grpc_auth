package characterclient

import (
	"context"
	"fmt"
	"strings"

	characterv1 "character/proto/character/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ResolvePlayCharacter loads a play template or row from character-service.
func ResolvePlayCharacter(ctx context.Context, addr, token string, userID int64, characterID string) (*characterv1.ResolvePlayCharacterResponse, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, fmt.Errorf("character service address is empty")
	}
	characterID = strings.TrimSpace(characterID)
	if characterID == "" {
		return nil, fmt.Errorf("character id is empty")
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("character_service=%q grpc dial: %w", addr, err)
	}
	defer func() { _ = conn.Close() }()

	ctx = withServiceToken(ctx, token)
	cli := characterv1.NewCharacterServiceClient(conn)
	resp, err := cli.ResolvePlayCharacter(ctx, &characterv1.ResolvePlayCharacterRequest{
		UserId: userID,
		Id:     characterID,
	})
	if err != nil {
		return nil, fmt.Errorf("character_service=%q: %w", addr, err)
	}
	return resp, nil
}

// PlayCharacter identifies the row/template for SavePlaySession.
type PlayCharacter struct {
	ID          string
	DisplayName string
	Description string
}

// SavePlaySession writes character.data after a play session (create if not yet persisted, else replace).
func SavePlaySession(ctx context.Context, addr, token string, userID int64, persisted bool, ch PlayCharacter, data []byte, schemaVersion int32) error {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return fmt.Errorf("character service address is empty")
	}
	id := strings.TrimSpace(ch.ID)
	if id == "" {
		return fmt.Errorf("character id is empty")
	}
	if data == nil {
		data = []byte{}
	}
	if schemaVersion == 0 {
		schemaVersion = 1
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("character_service=%q grpc dial: %w", addr, err)
	}
	defer func() { _ = conn.Close() }()

	ctx = withServiceToken(ctx, token)
	cli := characterv1.NewCharacterServiceClient(conn)

	if persisted {
		_, err := cli.ReplaceCharacterData(ctx, &characterv1.ReplaceCharacterDataRequest{
			UserId:          userID,
			Id:              id,
			Data:            data,
			SchemaVersion:   schemaVersion,
			ExpectedVersion: 0,
		})
		if err != nil {
			return fmt.Errorf("character_service=%q replace character data: %w", addr, err)
		}
		return nil
	}

	displayName := strings.TrimSpace(ch.DisplayName)
	if displayName == "" {
		displayName = "Adventurer"
	}
	_, err = cli.CreateCharacter(ctx, &characterv1.CreateCharacterRequest{
		UserId:          userID,
		Id:              id,
		DisplayName:     displayName,
		Description:     strings.TrimSpace(ch.Description),
		Data:            data,
		SchemaVersion:   schemaVersion,
	})
	if err != nil {
		return fmt.Errorf("character_service=%q create character: %w", addr, err)
	}
	return nil
}

func withServiceToken(ctx context.Context, token string) context.Context {
	if strings.TrimSpace(token) == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "x-service-token", token)
}
