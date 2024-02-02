package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/db"
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

func (s *PgMemberStorage) Update(ctx context.Context, member *Member) error {
	fieldsToUpdate := make(map[string]interface{})

	setClause := "SET "

	if member.Roles != nil {
		fieldsToUpdate["roles"] = pq.StringArray(member.Roles)
		setClause += "roles = :roles, "
	}

	for _, field := range []struct {
		name  string
		value interface{}
	}{
		{"display_name", member.DisplayName},
		{"email", member.Email},
		{"bio", member.Bio},
		{"active", member.Active},
		{"visibility", member.Visibility},
		{"following_uri", member.FollowingURI},
		{"followers_uri", member.FollowersURI},
		{"session_timeout", member.SessionTimeout},
		{"public_key_pem", member.PublicKeyPem},
		{"profilepic_id", member.ProfilePicID},
		{"nick", member.MemberName},
		{"custom_fields", member.CustomFields},
	} {
		if db.IsNotNull(field.value) {
			fieldsToUpdate[field.name] = field.value
			setClause += field.name + " = :" + field.name + ", "
		}
	}

	// remove the trailing comma and space
	setClause = strings.TrimSuffix(setClause, ", ")
	s.log.Trace().Msgf("setClause: %s", setClause)

	stmt := fmt.Sprintf(`
		UPDATE public.members AS m
		%s
		FROM (SELECT id FROM members WHERE nick = :nick) AS subquery
		WHERE m.id = subquery.id
	`, setClause)
	namedQuery, err := s.client.PrepareNamedContext(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to update member: %v", err)
	}
	s.log.Trace().Msgf("namedQuery: %+v", *namedQuery)
	defer namedQuery.Close()

	var res sql.Result
	res, err = namedQuery.ExecContext(ctx, map[string]interface{}{
		"display_name":    member.DisplayName,
		"email":           member.Email,
		"bio":             member.Bio,
		"active":          member.Active,
		"visibility":      member.Visibility,
		"following_uri":   member.FollowingURI,
		"followers_uri":   member.FollowersURI,
		"session_timeout": member.SessionTimeout,
		"public_key_pem":  member.PublicKeyPem,
		"nick":            member.MemberName,
		"profilepic_id":   member.ProfilePicID,
		"custom_fields":   member.CustomFields,
	})
	if err != nil {
		return fmt.Errorf("failed to update member: %v", err)
	}
	s.log.Trace().Msgf("result of UPDATE query: %v", res)

	return nil
}

func (s *PgMemberStorage) UpdatePassword(ctx context.Context, nick, pass string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := s.newClient.BeginTx(ctx, pgx.TxOptions{
			AccessMode: pgx.ReadWrite,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %v", err)
		}

		// nolint:errcheck // In case of failure during commit, "err" from commit will be returned
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, `UPDATE members SET passhash = $1 WHERE nick = $2`, pass, nick)

		if err != nil {
			return fmt.Errorf("failed to update password: %v", err)
		}

		return tx.Commit(ctx)
	}
}

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

func (s *PgMemberStorage) Read(ctx context.Context, value string, keyNames ...string) (*Member, error) {
	if lo.Contains(keyNames, "email_or_username") {
		keyNames = []string{"email", "nick"}
	}
	keyNames = db.Sanitize(keyNames)
	var query string
	if len(keyNames) == 2 {
		query = fmt.Sprintf("SELECT * FROM members WHERE %s = $1 OR %s = $1 LIMIT 1", keyNames[0], keyNames[1])
	} else {
		query = fmt.Sprintf("SELECT * FROM members WHERE %s = $1 LIMIT 1", keyNames[0])
	}
	member := &Member{}
	st, err := s.client.PreparexContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to read member: %v", err)
	}
	defer st.Close()

	err = st.GetContext(ctx, member, value)
	if err != nil {
		return nil, fmt.Errorf("failed to read member: %v", err)
	}

	return member, nil
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

// GetID retrieves the ID required for JWT on the basis of one of the credentials,
// i.e. email or login
func (s *PgMemberStorage) GetID(ctx context.Context, credential string) (id int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		query := `SELECT id FROM members WHERE email = $1 OR nick = $2`
		err = s.client.Get(&id, query, credential, credential)
		if err != nil {
			return 0, fmt.Errorf("failed to get member id: %v", err)
		}
		return id, nil
	}
}

// GetPassHash retrieves the password hash required for JWT on the basis of one of the credentials,
// i.e. email or login
func (s *PgMemberStorage) GetPassHash(email, login string) (string, error) {
	query := `SELECT passhash FROM members WHERE email = $1 OR nick = $2`
	var passHash string
	err := s.client.Get(&passHash, query, email, login)
	if err != nil {
		return "", fmt.Errorf("failed to get passhash: %v", err)
	}
	return passHash, nil
}

// CreateSession creates a JWT token for the member
func (s *PgMemberStorage) CreateSession(ctx context.Context, m *Member) (t string, err error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		token := *jwt.New(jwt.SigningMethodHS512)
		claims := token.Claims.(jwt.MapClaims)
		claims["id"] = m.ID
		if m.MemberName != "" {
			claims["membername"] = m.MemberName
		} else {
			claims["email"] = m.Email
		}
		claims["exp"] = time.Now().Add(time.Hour * 12).Unix()

		t, err = token.SignedString([]byte(s.config.Secret))
		if err != nil {
			return "", fmt.Errorf("failed to sign token: %v", err)
		}
		return t, nil
	}
}

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
		err := s.client.Get(&id, query, email, nickname)
		// for example if sql.ErrNoRows is returned, it means that the member does not exist
		if err != nil {
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
