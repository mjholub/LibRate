package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Members(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err := createEnumType(ctx, db, "role", "public", "admin", "mod", "regular", "creator")
		if err != nil {
			return fmt.Errorf("failed to create role enum: %w", err)
		}
		_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS public.members (
	id serial4 NOT NULL,
	uuid uuid NOT NULL,
	nick varchar(255) NOT NULL,
	email varchar(255) NOT NULL,
	passhash varchar(255) NOT NULL,
	reg_timestamp timestamp NOT NULL DEFAULT now(),
	profilepic_id int8 NULL,
	display_name varchar NULL,
	homepage varchar NULL,
	irc varchar NULL,
	xmpp varchar NULL,
	matrix varchar NULL,
	bio text NULL,
	active bool NOT NULL DEFAULT false,
	roles public."_role" NOT NULL,
	CONSTRAINT members_pkey PRIMARY KEY (id)
);
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
	`)
		if err != nil {
			return fmt.Errorf("failed to create members table: %w", err)
		}
		return nil
	}
}

func MembersProfilePic(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := db.Exec(`
			ALTER TABLE public.members
			ADD CONSTRAINT members_profilepic_fkey FOREIGN KEY (profilepic_id)
			REFERENCES cdn.images(id);
	`)
		if err != nil {
			return fmt.Errorf("failed to create images table: %w", err)
		}
		return nil
	}
}
