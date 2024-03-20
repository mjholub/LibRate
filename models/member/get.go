package member

import (
	"context"
	"fmt"

	"codeberg.org/mjh/LibRate/db"
	"github.com/samber/lo"
)

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

// GetID retrieves the ID required for JWT on the basis of one of the credentials,
// i.e. email or login
func (s *PgMemberStorage) GetID(ctx context.Context, credential string) (id int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		query := `SELECT id_numeric FROM members WHERE email = $1 OR nick = $2`
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
