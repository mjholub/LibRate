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

	row := s.newClient.QueryRow(ctx, query, value)
	if err := row.Scan(&member); err != nil {
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

		res, err := db.SerializableParametrizedTx[int](ctx,
			s.newClient, "get-member-id", query, credential, credential)
		if err != nil {
			return 0, fmt.Errorf("failed to get member id: %v", err)
		}
		return res[0], nil
	}
}

// GetPassHash retrieves the password hash required for JWT on the basis of one of the credentials,
// i.e. email or login
func (s *PgMemberStorage) GetPassHash(ctx context.Context, email, login string) (string, error) {
	query := `SELECT passhash FROM members WHERE email = $1 OR nick = $2`
	tx, err := s.newClient.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %v", err)
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	stmt, err := tx.Prepare(ctx, "get-member-passhash", query)
	if err != nil {
		return "", fmt.Errorf("failed to prepare statement: %v", err)
	}

	var pHash string

	err = tx.QueryRow(ctx, stmt.Name,
		email, login).Scan(&pHash)
	if err != nil {
		return "", fmt.Errorf("failed to get passhash for params: %+v", map[string]string{"email": email, "login": login})
	}

	return pHash, nil
}
