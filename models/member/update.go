package member

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"codeberg.org/mjh/LibRate/db"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
)

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
