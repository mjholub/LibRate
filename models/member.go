package models

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Member struct {
	UUID         string `json:"_key,omitempty" db:"uuid"`
	PassHash     string `json:"passhash" db:"passhash"`
	MemberName   string `json:"membername" db:"nick"`
	Email        string `json:"email" db:"email"`
	RegTimestamp int64  `json:"regdate" db:"reg_timestamp"`
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
}

func NewMemberStorage(client *sqlx.DB) *MemberStorage {
	return &MemberStorage{client: client}
}

func (s *MemberStorage) Save(ctx context.Context, member *Member) error {
	query := `INSERT INTO members (uuid, passhash, nick, email, reg_timestamp) 
            VALUES (uuid_generate_v4(), :passhash, :nick, :email, :reg_timestamp)`
	stmt, err := s.client.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	params := struct {
		PassHash     string `db:"passhash"`
		MemberName   string `db:"nick"`
		Email        string `db:"email"`
		RegTimestamp int64  `db:"reg_timestamp"`
	}{
		PassHash:     member.PassHash,
		MemberName:   member.MemberName,
		Email:        member.Email,
		RegTimestamp: member.RegTimestamp,
	}

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
	query := `SELECT * FROM members WHERE ` + keyName + ` = ?`
	member := &Member{}
	err := s.client.GetContext(ctx, member, query, key)
	if err != nil {
		return nil, fmt.Errorf("failed to read member: %v", err)
	}
	return member, nil
}
