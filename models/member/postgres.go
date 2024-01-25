package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/db"
)

func (s *PgMemberStorage) Save(ctx context.Context, member *Member) error {
	// first, check if nick or email is already taken
	findEmailOrNickQuery := `SELECT id FROM members WHERE nick = $1 OR email = $2`
	var id uint32
	err := s.client.Get(&id, findEmailOrNickQuery, member.MemberName, member.Email)
	if err == nil {
		return fmt.Errorf("email %s or nick %s is already taken", member.Email, member.MemberName)
	}

	query := `INSERT INTO members (uuid, passhash, nick, email, reg_timestamp, active, roles) 
	VALUES (:uuid, :passhash, :nick, :email, to_timestamp(:reg_timestamp), :active, :roles)`
	stmt, err := s.client.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	params := map[string]interface{}{
		"uuid":          member.UUID,
		"passhash":      member.PassHash,
		"nick":          member.MemberName,
		"email":         member.Email,
		"reg_timestamp": member.RegTimestamp.Unix(),
		"active":        true,
		"roles":         pq.StringArray(member.Roles),
	}

	s.log.Debug().Msgf("params: %v", params)
	fmt.Printf("params: %v", params)
	fmt.Printf("statement: %v", stmt)

	_, err = stmt.ExecContext(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to save member: %v", err)
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
		{"homepage", member.Homepage},
		{"irc", member.IRC},
		{"xmpp", member.XMPP},
		{"matrix", member.Matrix},
		{"visibility", member.Visibility},
		{"following_uri", member.FollowingURI},
		{"followers_uri", member.FollowersURI},
		{"session_timeout", member.SessionTimeout},
		{"public_key_pem", member.PublicKeyPem},
		{"profilepic_id", member.ProfilePicID},
		{"nick", member.MemberName},
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
		"roles":           pq.StringArray(member.Roles),
		"homepage":        member.Homepage,
		"irc":             member.IRC,
		"xmpp":            member.XMPP,
		"matrix":          member.Matrix,
		"visibility":      member.Visibility,
		"following_uri":   member.FollowingURI,
		"followers_uri":   member.FollowersURI,
		"session_timeout": member.SessionTimeout,
		"public_key_pem":  member.PublicKeyPem,
		"nick":            member.MemberName,
		"profilepic_id":   member.ProfilePicID,
	})
	if err != nil {
		return fmt.Errorf("failed to update member: %v", err)
	}
	s.log.Trace().Msgf("result of UPDATE query: %v", res)

	return nil
}

func (s *PgMemberStorage) Delete(ctx context.Context, member *Member) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := s.client.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %v", err)
		}
		defer tx.Rollback()
		_, err = s.client.ExecContext(ctx, `DELETE FROM members WHERE id = $1`, member.ID)
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

func (s *PgMemberStorage) VerifyViewability(ctx context.Context, viewer, viewee string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		var visibility string
		var canView bool
		viewer = db.Sanitize([]string{viewer})[0]
		viewee = db.Sanitize([]string{viewee})[0]
		err := s.newClient.QueryRow(ctx, `SELECT visibility FROM members WHERE nick = $1`, viewee).Scan(&visibility)
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

// RequestFollow creates a follow request in the local database
// upon the reception of a request into the inbox
func (s *PgMemberStorage) RequestFollow(ctx context.Context, fr *FollowRequest) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := s.client.NamedExecContext(ctx, `INSERT INTO follow_requests (actor_id, follows_id) VALUES (:actor_id, :follows_id)`, fr)
		if err != nil {
			return fmt.Errorf("failed to save follow request: %v", err)
		}
		return nil
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
		_, err := uuid.Parse(member.UUID.String())
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
