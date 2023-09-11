package models

import (
	"context"
	"fmt"
)

func (s *MemberStorage) CacheNicknames(ctx context.Context) error {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// Use the Read method to retrieve all nicknames
	// Assuming 'nickname' is the column name for nicknames
	query := "SELECT nickname FROM members"
	var nicknames []string
	if err := s.client.SelectContext(ctx, &nicknames, query); err != nil {
		return fmt.Errorf("failed to cache nicknames: %v", err)
	}

	// Update the nickname cache
	s.nicknameCache = nicknames

	s.log.Info().Msgf("cached %d nicknames", len(nicknames))
	return nil
}

func (s *MemberStorage) GetNicknames() []string {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	return s.nicknameCache
}
