package member

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// RequestFollow creates a follow request in the local database
// upon the reception of a request into the inbox
func (s *PgMemberStorage) RequestFollow(ctx context.Context, fr *FollowBlockRequest) FollowResponse {
	select {
	case <-ctx.Done():
		return FollowResponse{
			Status: "failed",
			Error:  ctx.Err(),
		}
	default:
		s.log.Debug().Msgf("incoming request: %+v", fr)
		_, e := mail.ParseAddress(fr.Requester)
		_, err := mail.ParseAddress(fr.Target)
		if err != nil || e != nil {
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("could not sanitize follow request: %v, %v", err, e),
			}
		}
		s.log.Debug().Msg("sanitized webfingers")

		_, err = s.GetID(ctx, strings.Split(fr.Target, "@")[0])
		if err != nil {
			return FollowResponse{
				Status: "not_found",
				Error:  fmt.Errorf("follow target %s not found: %v", fr.Target, err),
			}
		}
		s.log.Debug().Msg("target found")

		blocked, err := s.IsBlocked(ctx, fr)
		if err != nil {
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("failed to check if %s is blocked by %s: %v", fr.Requester, fr.Target, err),
			}
		}
		if blocked {
			return FollowResponse{
				Status: "blocked",
			}
		}
		s.log.Debug().Msg("not blocked")

		// check for duplicate follow request
		st, err := s.client.PreparexContext(ctx, `SELECT EXISTS (
    SELECT 1
    FROM public.follow_requests AS fr
    LEFT JOIN public.followers AS f ON fr.requester_webfinger = f.follower AND fr.target_webfinger = f.followee
    WHERE fr.requester_webfinger = $1 AND fr.target_webfinger = $2
) AS duplicate;`)
		if err != nil {
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("failed to prepare statement to check for duplicate request: %v", err),
			}
		}
		var duplicate bool
		if err = st.GetContext(ctx, &duplicate, fr.Requester, fr.Target); err != nil {
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("failed to check for duplicate request: %v", err),
			}
		}
		if duplicate {
			return FollowResponse{
				Status: "already_following",
			}
		}

		s.log.Debug().Msg("no duplicate request")

		// check if target has auto_accept_follow enabled in public.member_prefs
		var autoAcceptFollow bool
		st, err = s.client.PreparexContext(ctx, `
		SELECT auto_accept_follow
		FROM public.member_prefs 
		WHERE member_id = (SELECT id FROM public.members WHERE webfinger = $1)`)
		if err != nil {
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("failed to prepare statement to check if follow acceptance is enabled: %v", err),
			}
		}

		if err = st.GetContext(ctx, &autoAcceptFollow, fr.Target); err != nil {
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("failed to check if follow acceptance is enabled: %v", err),
			}
		}

		if autoAcceptFollow {
			s.log.Debug().Msg("auto accept follow enabled")
			_, err = s.client.ExecContext(ctx, `INSERT INTO public.followers (follower, followee, notifications, reblogs) 
			VALUES ($1, $2, $3, $4)`,
				fr.Requester, fr.Target, fr.Notify, fr.Reblogs)
			if err != nil {
				return FollowResponse{
					Status: "failed",
					Error:  fmt.Errorf("failed to save follower: %v", err),
				}
			}
			now := time.Now()
			return FollowResponse{
				Status:     "accepted",
				AcceptTime: &now,
			}
		}

		s.log.Debug().Msg("auto accept follow disabled")

		row := s.client.QueryRowContext(ctx, `INSERT INTO follow_requests 
		(requester_webfinger, target_webfinger, reblogs, notifications)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, fr.Requester, fr.Target, fr.Reblogs, fr.Notify)

		var id int64

		if err = row.Scan(&id); err != nil {
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("failed to save follow request: %v", err),
			}
		}
		return FollowResponse{
			Status: "pending",
			ID:     id,
		}
	}
}

func (s *PgMemberStorage) AcceptFollow(ctx context.Context, accepter string, requestID int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := s.newClient.BeginTx(ctx, pgx.TxOptions{
			AccessMode:     pgx.ReadWrite,
			DeferrableMode: pgx.Deferrable,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %v", err)
		}
		//nolint:errcheck // In case of failure during commit, "err" from commit will be returned
		defer tx.Rollback(ctx)

		// check if the request with the given ID was sent to the accepter
		var target string

		err = tx.QueryRow(ctx, `SELECT target_webfinger FROM public.follow_requests WHERE id = $1`, requestID).Scan(&target)
		if err != nil {
			return fmt.Errorf("error chectking if request with ID %d exists: %v", requestID, err)
		}
		if target != accepter {
			return fmt.Errorf("request with ID %d does not belong to %s", requestID, accepter)
		}

		batch := &pgx.Batch{}
		// copy all rows but 'created' (creation set by DB) to followers
		batch.Queue(`INSERT INTO public.followers VALUES (
				SELECT reblogs, notifications, requester_webfinger, target_webfinger FROM public.follow_requests WHERE id = $1)
			)`, requestID)
		batch.Queue(`DELETE FROM public.follow_requests WHERE id = $1`, requestID)

		br := tx.SendBatch(ctx, batch)
		_, err = br.Exec()
		br.Close()
		if err != nil {
			return fmt.Errorf("failed to execute batch: %v", err)
		}
		return tx.Commit(ctx)
	}
}

// TODO: add option to send a note to the requester stating the reason for rejection
func (s *PgMemberStorage) RejectFollow(ctx context.Context, rejecter string, requestID int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

		// check if the request with the given ID was sent to the accepter
		var target string

		err := s.newClient.QueryRow(ctx, `SELECT target_webfinger FROM public.follow_requests WHERE id = $1`, requestID).Scan(&target)
		if err != nil {
			return fmt.Errorf("error checking if request with ID %d exists: %v", requestID, err)
		}

		if target != rejecter {
			return fmt.Errorf("request with ID %d does not belong to %s", requestID, rejecter)
		}

		_, err = s.client.ExecContext(ctx, `DELETE FROM public.follow_requests WHERE id = $1`, requestID)
		if err != nil {
			return fmt.Errorf("failed to delete follow request: %v", err)
		}
		return nil
	}
}

func (s *PgMemberStorage) CancelFollow(ctx context.Context, requester string, requestID int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// check if the request with the given ID was sent by the requester
		var initiator string
		err := s.newClient.QueryRow(ctx, `SELECT requester_webfinger 
		FROM public.follow_requests WHERE id = $1`, requestID).Scan(&initiator)
		if err != nil {
			return fmt.Errorf("failed to check if request with ID %d exists: %v", requestID, err)
		}
		if initiator != requester {
			return fmt.Errorf("request with ID %d does not belong to %s", requestID, requester)
		}

		_, err = s.newClient.Exec(ctx, `DELETE FROM public.follow_requests WHERE id = $1`, requestID)
		if err != nil {
			return fmt.Errorf("failed to delete follow request: %v", err)
		}
		return nil
	}
}

// RemoveFollower handles both the followee and follower initiated removal of a follower
// due to the reciprocal nature of the relationship
// It can also deal with cancelling pending follow requests
func (s *PgMemberStorage) RemoveFollower(ctx context.Context, follower, followee string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := mail.ParseAddress(follower)
		_, e := mail.ParseAddress(followee)
		if err != nil || e != nil {
			return fmt.Errorf("failed to sanitize follow request: %v, %v", err, e)
		}
		// check if there is a pending follow request from follower to followee
		var pending bool

		row := s.newClient.QueryRow(ctx, `SELECT EXISTS(
			SELECT 1 FROM public.follow_requests WHERE
			requester_webfinger = $1 AND target_webfinger = $2
		) AS pending`, follower, followee)
		if err = row.Scan(&pending); err != nil {
			return fmt.Errorf("failed to check if follow request is pending: %v", err)
		}
		if pending {
			_, err = s.newClient.Exec(ctx, `DELETE FROM public.follow_requests WHERE requester_webfinger = $1 AND target_webfinger = $2`, follower, followee)
			if err != nil {
				return fmt.Errorf("failed to delete follow request: %v", err)
			}
			return nil
		}

		_, err = s.newClient.Exec(ctx, `DELETE FROM public.followers WHERE follower = $1 AND followee = $2`, follower, followee)
		if err != nil {
			return fmt.Errorf("failed to delete follower: %v", err)
		}
		return nil
	}
}

func (s *PgMemberStorage) GetFollowRequests(
	ctx context.Context,
	member string,
	kind string,
) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		var (
			rows pgx.Rows
			err  error
		)
		switch kind {
		case "received":
			rows, err = s.newClient.Query(ctx, `SELECT * FROM follow_requests WHERE target_webfinger = $1`, member)
			if err != nil {
				return nil, fmt.Errorf("failed to get received follow requests: %v", err)
			}
			defer rows.Close()
			var received []FollowRequestIn
			for rows.Next() {
				var r FollowRequestIn
				if err = rows.Scan(&r); err != nil {
					return nil, fmt.Errorf("failed to scan row: %v", err)
				}
				received = append(received, r)
			}
			return received, nil
		case "sent":
			rows, err = s.newClient.Query(ctx, `SELECT * FROM follow_requests WHERE requester_webfinger = $1`, member)
			if err != nil {
				return nil, fmt.Errorf("failed to get sent follow requests: %v", err)
			}
			defer rows.Close()
			var sent []FollowResponse
			for rows.Next() {
				var r FollowResponse
				if err = rows.Scan(&r); err != nil {
					return nil, fmt.Errorf("failed to scan row: %v", err)
				}
				sent = append(sent, r)
			}
			return sent, nil
		case "all":
			// combine sent and received requests
			rows, err = s.newClient.Query(ctx, `
			SELECT * FROM follow_requests WHERE target_webfinger = $1
			UNION
			SELECT * FROM follow_requests WHERE requester_webfinger = $1`, member)
			if err != nil {
				return uint8(0), fmt.Errorf("failed to get follow requests: %v", err)
			}
			defer rows.Close()
			var group FollowRequestGroup

			for rows.Next() {
				var sent FollowResponse
				var received FollowRequestIn

				if err = rows.Scan(&sent, &received); err != nil {
					return nil, fmt.Errorf("failed to scan row: %v", err)
				}
				group.Sent = append(group.Sent, sent)
				group.Received = append(group.Received, received)
			}
			return group, nil
		default:
			return nil, fmt.Errorf("invalid follow request type: %s", kind)
		}
	}
}

func (s *PgMemberStorage) GetFollowStatus(ctx context.Context, follower, followee string) FollowResponse {
	select {
	case <-ctx.Done():
		return FollowResponse{
			Status: "failed",
			Error:  ctx.Err(),
		}
	default:
		var resp FollowResponse
		_, err := mail.ParseAddress(followee)
		if err != nil {
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("invalid followee: %v", err),
			}
		}
		tx, err := s.newClient.BeginTx(ctx, pgx.TxOptions{
			AccessMode:     pgx.ReadOnly,
			DeferrableMode: pgx.NotDeferrable,
		})
		if err != nil {
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("failed to begin transaction: %v", err),
			}
		}
		//nolint:errcheck // In case of failure during commit, "err" from commit will be returned
		defer tx.Rollback(ctx)
		// we'll check both the followers and follow_requests tables
		followRow := tx.QueryRow(ctx, `SELECT reblogs, notifications, created FROM public.followers
		WHERE follower = $1 AND followee = $2
		LIMIT 1`, follower, followee)
		followErr := followRow.Scan(&resp.Reblogs, &resp.Notify, &resp.AcceptTime)
		requestRow := tx.QueryRow(ctx, `
		SELECT id, reblogs, notifications FROM public.follow_requests
		WHERE requester_webfinger = $1 AND target_webfinger = $2
		LIMIT 1`, follower, followee)
		requestErr := requestRow.Scan(&resp.ID, &resp.Reblogs, &resp.Notify)
		switch {
		case followErr.Error() == pgx.ErrNoRows.Error() && requestErr.Error() == pgx.ErrNoRows.Error():
			return FollowResponse{
				Status: "not_found",
			}
		case followErr == nil:
			return FollowResponse{
				Status:     "accepted",
				Reblogs:    resp.Reblogs,
				Notify:     resp.Notify,
				AcceptTime: resp.AcceptTime,
			}
		case requestErr == nil:
			return FollowResponse{
				Status:  "pending",
				Reblogs: resp.Reblogs,
				Notify:  resp.Notify,
			}
		default:
			return FollowResponse{
				Status: "failed",
				Error:  fmt.Errorf("failed to check follow status: %v/%v", followErr, requestErr),
			}
		}
	}
}
