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
			name, aliases, roles, birth, death, website, bio, hometown, residence)
				VALUES (COALESCE($1, ''), COALESCE($2, '{}'), COALESCE($3, '{}'),
				COALESCE($4, NULL), COALESCE($5, NULL), COALESCE($6, ''),
				COALESCE($7, 'No bio yet!'), COALESCE($8, NULL), COALESCE($9, NULL))
				RETURNING id;`,
			person.Name, pq.StringArray(person.Aliases), pq.StringArray(person.Roles),
			person.Birth, person.Death, person.Website,
			person.Bio, person.Hometown, person.Residence).
			Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to insert person: %w", err)
		}

		return &id, tx.Commit(ctx)
	}
}

func (ps *PeopleStorage) CreateGroup(ctx context.Context, group *Group) (*uuid.UUID, error) {
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

		// insert the group
		if err = tx.QueryRow(ctx, `
				INSERT INTO people.group (
			name, aliases, website, bio, active, formed, disbanded,
			kind, wikipedia, bandcamp, soundcloud)
				VALUES (COALESCE($1, ''), COALESCE($2, '{}'), $3,
				COALESCE($4, ''), COALESCE($5, ''), COALESCE($6, true), COALESCE($7, NULL),
				COALESCE($8, ''), $9, $10, $11)
				RETURNING id;`,
			group.Name, pq.StringArray(group.Aliases),
			group.Website, group.Bio, group.Active, group.Formed.Time, group.Disbanded.Time,
			group.Kind, group.Wikipedia, group.Bandcamp, group.Soundcloud).
			Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to insert group: %w", err)
		}

		return &id, tx.Commit(ctx)
	}
}
