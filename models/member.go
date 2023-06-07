package models

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Member struct {
	UUID         string `json:"_key,omitempty"`
	PassHash     string `json:"passhash"`
	MemberName   string `json:"membername"`
	Email        string `json:"email"`
	RegTimestamp int64  `json:"regdate"`
}

type MemberInput struct {
	MemberName string `json:"membername"`
	Email      string `json:"email"`
}

type RegisterInput struct {
	Email           string `json:"email"`
	MemberName      string `json:"membername"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordconfirm"`
}

type LoginInput struct {
	Email      string `json:"email"`
	MemberName string `json:"membername"`
	Password   string `json:"password"`
}

type RegLoginInput interface {
	RegisterInput | LoginInput
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
	query := `INSERT INTO members (field1, field2, ...) VALUES (:field1, :field2, ...)`
	_, err := s.client.NamedExecContext(ctx, query, member)
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
