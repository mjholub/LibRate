package media

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/lib/pq"
)

func (ps *PeopleStorage) CreatePerson(ctx context.Context, person *Person) (*uuid.UUID, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		tx, err := ps.newDBConn.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		var id uuid.UUID

		// insert the person
		if err = tx.QueryRow(ctx, `
				INSERT INTO people.person (
			first_name, other_names, last_name, 
		nick_names, roles, birth, death, website, bio, photos, hometown, residence)
				VALUES (COALESCE($1, ''), 
				COALESCE($2, ''), 
				COALESCE($3, '{}'),
				COALESCE($4, ''),
				COALESCE($5, '{}', COALESCE($6, '{}'), COALESCE($7, NULL),
				COALESCE($8, NULL), COALESCE($9, NULL), COALESCE($10, NULL),
				COALESCE($11, '{}'), COALESCE($12, NULL), COALESCE($13, NULL))
	RETURNING id;`, person.Name, person.FirstName, pq.StringArray(person.OtherNames),
			person.LastName, pq.StringArray(person.NickNames), pq.StringArray(person.Roles), person.Birth,
			person.Death, person.Website, person.Bio, person.Photos, person.Hometown,
			person.Residence).Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to insert person: %w", err)
		}

		return &id, tx.Commit(ctx)
	}
}
