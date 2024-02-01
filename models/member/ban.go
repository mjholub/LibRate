package member

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *PgMemberStorage) Ban(ctx context.Context, member *Member, input *BanInput) (err error) {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := s.newClient.BeginTx(ctx, pgx.TxOptions{
			AccessMode:     pgx.ReadWrite,
			DeferrableMode: pgx.NotDeferrable,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %v", err)
		}
		defer func() {
			if e := tx.Rollback(ctx); e != nil {
				if rb := tx.Rollback(ctx); rb != nil {
					err = fmt.Errorf("failed to rollback transaction: %v", rb)
				} else {
					err = fmt.Errorf("transaction rolled bact: %v", err)
				}
			}
		}()

		// occurrence and start time are set by the database
		_, err = tx.Prepare(ctx, "ban", `
			INSERT INTO bans (member_uuid, reason, ends, can_appeal, mask)
			VALUES ($1, $2, $3, $4, $5)`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %v", err)
		}

		_, err = tx.Exec(ctx, "ban", member.UUID, input.Reason, input.Ends, input.CanAppeal, input.Mask)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %v", err)
		}

		return tx.Commit(ctx)

	}
}

func (s *PgMemberStorage) Unban(ctx context.Context, member *Member) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := member.UUID.MarshalBinary()
		if err != nil {
			return fmt.Errorf("invalid UUID: %v", member.UUID)
		}
		_, err = s.client.ExecContext(ctx, `DELETE FROM bans WHERE member_uuid = $1`, member.UUID)
		if err != nil {
			return fmt.Errorf("failed to delete ban: %v", err)
		}
		return nil
	}
}
