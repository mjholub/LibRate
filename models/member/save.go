package member

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
)

func (s *PgMemberStorage) Save(ctx context.Context, member *Member) error {
	// first, check if nick or email is already taken
	findEmailOrNickQuery := `SELECT id FROM members WHERE nick = $1 OR email = $2`
	var id uint32
	var mu sync.Mutex
	mu.Lock()
	err := s.client.Get(&id, findEmailOrNickQuery, member.MemberName, member.Email)
	if err == nil {
		return fmt.Errorf("email %s or nick %s is already taken", member.Email, member.MemberName)
	}
	mu.Unlock()

	batch := &pgx.Batch{}
	batch.Queue(`INSERT INTO public.members (uuid, passhash, nick, webfinger, email, reg_timestamp, active, roles)
	VALUES ($1, $2, $3, $4, $5, to_timestamp($6), $7, $8)
	RETURNING id`,
		member.UUID, member.PassHash, member.MemberName, member.Webfinger,
		member.Email, member.RegTimestamp.Unix(), member.Active, pq.StringArray(member.Roles))

	tx, err := s.newClient.BeginTx(ctx, pgx.TxOptions{
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer tx.Rollback(ctx) //nolint:errcheck // In case of failure during commit, "err" from commit will be returned

	conn, err := s.newClient.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	br := tx.SendBatch(ctx, batch)
	row := br.QueryRow()
	if err = row.Scan(&id); err != nil {
		return fmt.Errorf("failed to get member ID: %v", err)
	}
	s.log.Debug().Msgf("id: %d", id)
	br.Close()
	_, err = tx.Exec(ctx, `INSERT INTO member_prefs (member_id) VALUES ($1)`, id)
	if err != nil {
		return fmt.Errorf("failed to save member prefs: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}
