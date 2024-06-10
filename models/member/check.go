package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"

	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"github.com/samber/lo"
)

func (s *PgMemberStorage) IsBlocked(ctx context.Context, fr *FollowBlockRequest) (blocked bool, err error) {
	select {
	case <-ctx.Done():
		return true, ctx.Err()
	default:
		tx, err := s.newClient.BeginTx(ctx, pgx.TxOptions{
			AccessMode:     pgx.ReadOnly,
			DeferrableMode: pgx.NotDeferrable,
		})
		if err != nil {
			return true, fmt.Errorf("failed to begin transaction: %v", err)
		}

		defer func() {
			if e := recover(); e != nil {
				if rb := tx.Rollback(ctx); rb != nil {
					err = fmt.Errorf("failed to rollback transaction: %v", rb)
				} else {
					err = fmt.Errorf("transaction rolled back: %v", e)
				}
			}
		}()

		_, err = tx.Prepare(ctx, "blocked", `
			SELECT true AS blocked
			FROM public.member_blocks
			WHERE
    		(requester_webfinger = $1 AND target_webfinger = $2)
    	OR
    		(requester_webfinger = $2 AND target_webfinger = $1)
			LIMIT 1`)
		if err != nil {
			return true, fmt.Errorf("failed to prepare statement: %v", err)
		}

		row, err := tx.Query(ctx, "blocked", fr.Requester, fr.Target)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, nil
			}
			return true, fmt.Errorf("failed to execute statement: %v", err)
		}
		defer row.Close()
		if !row.Next() {
			return false, nil
		}

		s.log.Debug().Msgf("query executed. Row: %+v", row)

		if err = row.Scan(&blocked); err != nil {
			return true, fmt.Errorf("failed to scan row: %v", err)
		}

		return blocked, tx.Commit(ctx)
	}
}

// Check checks if a member with the given email or nickname already exists
func (s *PgMemberStorage) Check(c context.Context, email, nickname string) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Error().Msgf("recovered from panic: %v", r)
		}
	}()

	select {
	case <-c.Done():
		return false, c.Err()
	default:
		query := `SELECT id FROM members WHERE email = $1 OR nick = $2`
		var id uint32
		if err := s.newClient.QueryRow(c, query, email, nickname).Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				id = 0
				return false, nil
			}
			return false, err
		}
		return true, nil
	}
}

// if exact is true, the role must match exactly
// otherwise we can match moderators with admins on tasks that can be performed by both
func (s *PgMemberStorage) HasRole(ctx context.Context, name, role string, exact bool) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		var roles pq.StringArray
		err := s.newClient.QueryRow(ctx, `SELECT roles FROM members WHERE nick = $1`, name).Scan(&roles)
		if err != nil {
			return false
		}
		if lo.Contains(roles, "mod") && !exact {
			return true
		}
		return lo.Contains(roles, role)
	}
}

// viewer and viewee are both identified by webfinger
func (s *PgMemberStorage) VerifyViewability(ctx context.Context, viewer, viewee string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		var (
			visibility string
			canView    bool
			err        error
		)
		if viewer != "" {
			_, err = mail.ParseAddress(viewer)
			if err != nil {
				return false, fmt.Errorf("invalid viewer: %v", err)
			}
		}
		_, err = mail.ParseAddress(viewee)
		if err != nil {
			return false, fmt.Errorf("invalid viewee: %v", err)
		}
		err = s.newClient.QueryRow(ctx, `SELECT visibility FROM members WHERE webfinger = $1`, viewee).Scan(&visibility)
		if err != nil {
			return false, fmt.Errorf("failed to get visibility: %v", err)
		}
		if viewer == "" {
			if visibility == "public" {
				return true, nil
			} else {
				return false, nil
			}
		}
		if viewer == viewee {
			return true, nil
		}
		switch visibility {
		case "public":
			return true, nil
		case "private":
			return false, nil
		case "followers_only":
			// compare based on the webfinger junction table
			err = s.newClient.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM public.followers WHERE
	follower = $1 AND followee = $2
) AS is_follower`).Scan(&canView, viewer, viewee)
			if err != nil {
				return false, fmt.Errorf("failed to check if viewer is a follower: %v", err)
			}
			return canView, nil
		default:
			return false, fmt.Errorf("invalid visibility: %s", visibility)
		}
	}
}
