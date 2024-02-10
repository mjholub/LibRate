package member

import (
	"context"
	"fmt"
)

func (s *PgMemberStorage) Delete(ctx context.Context, memberName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := s.client.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %v", err)
		}
		defer tx.Rollback()
		_, err = s.client.ExecContext(ctx, `DELETE FROM members WHERE nick = $1`, memberName)
		if err != nil {
			return fmt.Errorf("failed to delete member: %v", err)
		}
		return tx.Commit()
	}
}
