package member

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *PgMemberStorage) Delete(ctx context.Context, memberName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := s.newClient.BeginTx(ctx, pgx.TxOptions{
			IsoLevel: pgx.Serializable,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %v", err)
		}
		defer tx.Rollback(ctx)
		_, err = tx.Exec(ctx, `DELETE FROM members WHERE nick = $1`, memberName)
		if err != nil {
			return fmt.Errorf("failed to delete member: %v", err)
		}
		return tx.Commit(ctx)
	}
}
