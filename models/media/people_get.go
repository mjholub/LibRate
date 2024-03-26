package media

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

func (p *PeopleStorage) GetPersonNames(ctx context.Context, id int32) (Person, error) {
	var person Person
	select {
	case <-ctx.Done():
		return Person{}, ctx.Err()
	default:
		err := p.dbConn.Get(&person, "SELECT first_name, last_name, other_names, nick_names FROM person WHERE id = $1", id)
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
		err := p.dbConn.Get(&person, "SELECT * FROM person WHERE id = $1", id)
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
		err := p.dbConn.Get(&group, "SELECT * FROM group WHERE id = $1", id)
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
		rows, err := p.newDBConn.Query(ctx, `SELECT *
FROM person
WHERE (first_name LIKE $1 OR last_name LIKE $1) OR $1 LIKE ANY(nick_names)`, name)
		if err != nil && err != sql.ErrNoRows {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var person Person
			if err := rows.Scan(&person.ID, &person.FirstName, &person.OtherNames, &person.LastName,
				&person.NickNames, &person.Roles, &person.Works, &person.Birth, &person.Death,
				&person.Website, &person.Bio, &person.Photos, &person.Hometown, &person.Residence,
				&person.Added, &person.Modified); err != nil {
				return nil, nil, err
			}
			persons = append(persons, person)
		}

		if err = rows.Err(); err != nil {
			return nil, nil, err
		}

		// Query for groups
		rows, err = p.newDBConn.Query(ctx, "SELECT * FROM group WHERE name LIKE $1", name)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var group Group
			if err := rows.Scan(&group.ID, &group.Locations, &group.Name, &group.Active,
				&group.Formed, &group.Disbanded, &group.Website, &group.Photos, &group.Works,
				&group.Members, &group.PrimaryGenre, &group.SecondaryGenres, &group.Kind,
				&group.Added, &group.Modified, &group.Wikipedia, &group.Bandcamp, &group.Soundcloud, &group.Bio); err != nil {
				return nil, nil, err
			}
			groups = append(groups, group)
		}

		if err := rows.Err(); err != nil {
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
		err := p.dbConn.Get(&studio, "SELECT * FROM studio WHERE id = $1", id)
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
		err := p.dbConn.Get(&group, "SELECT name FROM group WHERE id = $1", id)
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

func (p *PeopleStorage) GetID(ctx context.Context, name, kind string) (id int32, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		switch kind {
		case "group":
			err := p.dbConn.GetContext(ctx, &id,
				"SELECT id FROM group WHERE name = $1 AND kind = $2 LIMIT 1",
				name, kind)
			if err != nil {
				return 0, err
			}
			return id, nil
		case "person":
			firstName := strings.Split(name, " ")[0]
			lastName := strings.Split(name, " ")[1]
			err := p.dbConn.GetContext(ctx, &id,
				"SELECT id FROM person WHERE first_name = $1 AND last_name = $2 LIMIT 1",
				firstName, lastName)
			if err != nil {
				return 0, err
			}
			return id, nil
		case "studio":
			err := p.dbConn.GetContext(ctx, &id,
				"SELECT id FROM studio WHERE name = $1 LIMIT 1",
				name)
			if err != nil {
				return 0, err
			}
			return id, nil
		default:
			return 0, fmt.Errorf("invalid kind: %s", kind)
		}
	}
}
