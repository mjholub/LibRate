package member

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"
)

func (s *PgMemberStorage) Save(ctx context.Context, member *Member) error {
	if err := s.validationProvider.Struct(member); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	// first, check if nick or email is already taken
	findEmailOrNickQuery := `SELECT id FROM members WHERE nick = $1 OR email = $2`
	var id uint32
	r := s.newClient.QueryRow(ctx, findEmailOrNickQuery, member.MemberName, member.Email)
	if err := r.Scan(&id); err == nil {
		return fmt.Errorf("email %q or nick %q is already taken", member.Email, member.MemberName)
	}

	row := s.newClient.QueryRow(ctx, `INSERT INTO public.members (passhash, nick, webfinger, email, reg_timestamp, active, roles)
	VALUES ($1, $2, $3, $4, to_timestamp($5), $6, $7)
	RETURNING id_numeric`, member.PassHash, member.MemberName, member.Webfinger,
		member.Email, time.Now().Unix(), member.Active, pq.StringArray(member.Roles))

	if err := row.Scan(&id); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	fmt.Println("id: ", id)

	s.log.Debug().Msgf("id: %d", id)
	_, err := s.newClient.Exec(ctx, `INSERT INTO member_prefs (member_id) VALUES ($1)`, id)
	if err != nil {
		return fmt.Errorf("failed to save member prefs: %v", err)
	}

	return nil
}
