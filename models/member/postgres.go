package member

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lib/pq"
	"github.com/samber/lo"
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

	// TODO: verify if there is no unnecessary copying here
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
	query := `UPDATE members SET field1 = :field1, field2 = :field2, ... WHERE id = :id`
	_, err := s.client.NamedExecContext(ctx, query, member)
	if err != nil {
		return fmt.Errorf("failed to update member: %v", err)
	}
	return nil
}

func (s *PgMemberStorage) Delete(ctx context.Context, member *Member) error {
	query := `DELETE FROM members WHERE id = :id`
	_, err := s.client.NamedExecContext(ctx, query, member)
	if err != nil {
		return fmt.Errorf("failed to delete member: %v", err)
	}
	return nil
}

func (s *PgMemberStorage) Read(ctx context.Context, value string, keyNames ...string) (*Member, error) {
	if lo.Contains(keyNames, "email_or_username") {
		keyNames = []string{"email", "nick"}
	}
	query := fmt.Sprintf("SELECT * FROM members WHERE %s = $1 OR %s = $1 LIMIT 1", keyNames[0], keyNames[1])
	member := &Member{}
	err := s.client.GetContext(ctx, member, query, value)
	if err != nil {
		return nil, fmt.Errorf("failed to read member: %v", err)
	}
	return member, nil
}

// GetID retrieves the ID required for JWT on the basis of one of the credentials,
// i.e. email or login
func (s *PgMemberStorage) GetID(ctx context.Context, credential string) (uint32, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		query := `SELECT id FROM members WHERE email = $1 OR nick = $2`
		var id uint32
		err := s.client.Get(&id, query, credential, credential)
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

// TODO: implement
func (s *PgMemberStorage) GetSessionTimeout(
	ctx context.Context, memberID int, deviceID uuid.UUID,
) (timeout int, err error) {
	return 0, fmt.Errorf("GetSessionTimeout not implemented yet")
}

func (s *PgMemberStorage) LookupDevice(ctx context.Context, deviceID uuid.UUID) error {
	return fmt.Errorf("LookupDevice not implemented yet")
}

// Check checks if a member with the given email or nickname already exists
func (s *PgMemberStorage) Check(c context.Context, email, nickname string) (bool, error) {
	select {
	case <-c.Done():
		return false, c.Err()
	default:
		query := `SELECT id FROM members WHERE email = $1 OR nick = $2`
		var id uint32
		err := s.client.Get(&id, query, email, nickname)
		// for example if sql.ErrNoRows is returned, it means that the member does not exist
		if err != nil {
			return false, err
		}
		return true, nil
	}
}
