package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"

	"github.com/lib/pq"
)

// TODO: convert to enum if this becomes possible in Go
// nolint:gochecknoglobals // roles
const (
	admin uint8 = iota
	mod
	regular
	creator
)

// Member holds the core information about a member
type Member struct {
	ID           uint32         `json:"id" db:"id"`
	UUID         string         `json:"_key,omitempty" db:"uuid"`
	PassHash     string         `json:"passhash" db:"passhash"`
	MemberName   string         `json:"membername" db:"nick"` // i.e. @nick@instance
	DisplayName  sql.NullString `json:"displayname:omitempty" db:"display_name"`
	Email        string         `json:"email" db:"email" validate:"required,email"`
	Bio          sql.NullString `json:"bio:omitempty" db:"bio"`
	Active       bool           `json:"active" db:"active"`
	Roles        []uint8        `json:"roles" db:"roles"`
	RegTimestamp time.Time      `json:"regdate" db:"reg_timestamp"`
	ProfilePic   *Image         `json:"profilepic:omitempty" db:"profilepic_id"`
	Homepage     sql.NullString `json:"homepage:omitempty" db:"homepage"`
	IRC          sql.NullString `json:"irc:omitempty" db:"irc"`
	XMPP         sql.NullString `json:"xmpp:omitempty" db:"xmpp"`
	Matrix       sql.NullString `json:"matrix:omitempty" db:"matrix"`
}

type MemberInput struct {
	MemberName string `json:"membername"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type MemberStorer interface {
	Save(ctx context.Context, member *Member) error
	Read(ctx context.Context, member *Member) error
	Update(ctx context.Context, member *Member) error
	Delete(ctx context.Context, member *Member) error
}

type MemberStorage struct {
	client *sqlx.DB
	log    *zerolog.Logger
	config *cfg.Config
}

func NewMemberStorage(client *sqlx.DB, log *zerolog.Logger, conf *cfg.Config) *MemberStorage {
	return &MemberStorage{client: client, log: log, config: conf}
}

func mapRoleCodesToStrings(roles []uint8) []string {
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		switch role {
		case admin:
			roleStrings[i] = "admin"
		case mod:
			roleStrings[i] = "mod"
		case regular:
			roleStrings[i] = "regular"
		case creator:
			roleStrings[i] = "creator"
		default:
			panic(fmt.Sprintf("invalid role: %d", role))
		}
	}
	return roleStrings
}

func (s *MemberStorage) Save(ctx context.Context, member *Member) error {
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
		"roles":         pq.Array(mapRoleCodesToStrings(member.Roles)),
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

func (s *MemberStorage) Update(ctx context.Context, member *Member) error {
	query := `UPDATE members SET field1 = :field1, field2 = :field2, ... WHERE id = :id`
	_, err := s.client.NamedExecContext(ctx, query, member)
	if err != nil {
		return fmt.Errorf("failed to update member: %v", err)
	}
	return nil
}

func (s *MemberStorage) Delete(ctx context.Context, member *Member) error {
	query := `DELETE FROM members WHERE id = :id`
	_, err := s.client.NamedExecContext(ctx, query, member)
	if err != nil {
		return fmt.Errorf("failed to delete member: %v", err)
	}
	return nil
}

func (s *MemberStorage) Read(ctx context.Context, keyName, key string) (*Member, error) {
	query := fmt.Sprintf("SELECT * FROM members WHERE %s = $1", keyName)
	member := &Member{}
	err := s.client.GetContext(ctx, member, query, key)
	if err != nil {
		return nil, fmt.Errorf("failed to read member: %v", err)
	}
	return member, nil
}

// GetID retrieves the ID required for JWT on the basis of one of the credentials,
// i.e. email or login
func (s *MemberStorage) GetID(ctx context.Context, credential string) (uint32, error) {
	query := `SELECT id FROM members WHERE email = $1 OR nick = $2`
	var id uint32
	err := s.client.Get(&id, query, credential, credential)
	if err != nil {
		return 0, fmt.Errorf("failed to get member id: %v", err)
	}
	return id, nil
}

func (s *MemberStorage) GetPassHash(email, login string) (string, error) {
	query := `SELECT passhash FROM members WHERE email = $1 OR nick = $2`
	var passHash string
	err := s.client.Get(&passHash, query, email, login)
	if err != nil {
		return "", fmt.Errorf("failed to get passhash: %v", err)
	}
	return passHash, nil
}

func (s *MemberStorage) CreateSession(ctx context.Context, m Member) (t string, err error) {
	token := *jwt.New(jwt.SigningMethodHS256)
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
