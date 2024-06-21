package media

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/samber/lo"

	scn "github.com/georgysavva/scany/v2/pgxscan"
)

func (p *PeopleStorage) GetPersonNames(ctx context.Context, id int32) (Person, error) {
	var person Person
	select {
	case <-ctx.Done():
		return Person{}, ctx.Err()
	default:
		err := scn.Get(ctx, p.dbConn, &person, "SELECT first_name, last_name, other_names, nick_names FROM person WHERE id = $1", id)
		if err != nil {
			return Person{}, err
		}
		return person, nil
	}
}

func (p *PeopleStorage) GetPerson(ctx context.Context, id int64) (Person, error) {
	var person Person
	select {
	case <-ctx.Done():
		return Person{}, ctx.Err()
	default:
		err := scn.Get(ctx, p.dbConn, &person, "SELECT * FROM person WHERE id = $1", id)
		if err != nil {
			return Person{}, err
		}
		return person, nil
	}
}

func (p *PeopleStorage) GetGroup(ctx context.Context, id int32) (Group, error) {
	var group Group
	select {
	case <-ctx.Done():
		return Group{}, ctx.Err()
	default:
		err := scn.Get(ctx, p.dbConn, &group, "SELECT * FROM group WHERE id = $1", id)
		if err != nil {
			return Group{}, err
		}
		return group, nil
	}
}

func (p *PeopleStorage) GetArtistsByName(ctx context.Context, name string) (persons []Person, groups []Group, err error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		var person *Person

		if err := scn.Select(ctx, p.newDBConn, &person, `SELECT *
FROM person
WHERE (first_name LIKE $1 OR last_name LIKE $1) OR $1 LIKE ANY(nick_names)`, name); err != nil {
			return nil, nil, err
		}

		// Query for groups
		var groups []Group
		if err = scn.Select(ctx, p.newDBConn, &groups, "SELECT * FROM group WHERE name LIKE $1", name); err != nil {
			return nil, nil, err
		}

		return persons, groups, nil
	}
}

func (p *PeopleStorage) GetStudio(ctx context.Context, id int32) (*Studio, error) {
	var studio Studio
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := scn.Get(ctx, p.dbConn, &studio, "SELECT * FROM studio WHERE id = $1", id)
		if err != nil {
			return nil, err
		}
		return &studio, nil
	}
}

func (p *PeopleStorage) GetGroupName(ctx context.Context, id int32) (Group, error) {
	var group Group
	select {
	case <-ctx.Done():
		return Group{}, ctx.Err()
	default:
		err := scn.Get(ctx, p.dbConn, &group, "SELECT name FROM group WHERE id = $1", id)
		if err != nil {
			return Group{}, err
		}
		return group, nil
	}
}

func (g *Group) Validate() error {
	GroupKinds := []string{
		"Orchestra",
		"Choir",
		"Ensemble",
		"Collective",
		"Band",
		"Troupe",
		"Other",
	}
	if lo.Contains(GroupKinds, g.Kind) {
		return nil
	}
	return fmt.Errorf("invalid group kind: %s, must be one of %s", g.Kind, strings.Join(GroupKinds, ", "))
}

func (p *PeopleStorage) GetID(ctx context.Context, name, kind string) (id uuid.UUID, err error) {
	select {
	case <-ctx.Done():
		return uuid.Nil, ctx.Err()
	default:
		switch kind {
		case "group":
			err := scn.Get(ctx, p.newDBConn, &id,
				"SELECT id FROM group WHERE name = $1 AND kind = $2 LIMIT 1",
				name, kind)
			if err != nil {
				return uuid.Nil, err
			}
			return id, nil
		case "person":
			firstName := strings.Split(name, " ")[0]
			lastName := strings.Split(name, " ")[1]
			err := scn.Get(ctx, p.newDBConn, &id,
				"SELECT id FROM person WHERE first_name = $1 AND last_name = $2 LIMIT 1",
				firstName, lastName)
			if err != nil {
				return uuid.Nil, err
			}
			return id, nil
		case "studio":
			err := scn.Get(ctx, p.newDBConn, &id,
				"SELECT id FROM studio WHERE name = $1 LIMIT 1",
				name)
			if err != nil {
				return uuid.Nil, err
			}
			return id, nil
		default:
			return uuid.Nil, fmt.Errorf("invalid kind: %s", kind)
		}
	}
}
