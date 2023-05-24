package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"codeberg.org/mjh/LibRate/internal/client"

	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
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
	Load(ctx context.Context, key string) (*Member, error)
	Save(ctx context.Context, member *Member) error
}

// MemberStorage implements the MemberStorer interface
type MemberStorage struct {
	client *dgo.Dgraph
}

func NewMemberStorage() (*MemberStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	conn, err := client.ConnectToService(ctx, "localhost", "9080")
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Dgraph: %v\n", err)
	}
	dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	return &MemberStorage{
		client: dgraphClient,
	}, nil
}

func (s *MemberStorage) Load(ctx context.Context, key string) (*Member, error) {
	query := fmt.Sprintf(`{
		member(func: eq(_key, "%s")) {
			_key
			passhash
			membername
			email
			regdate
		}
	}`, key)

	resp, err := s.client.NewTxn().Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var result struct {
		Member []*Member `json:"member"`
	}

	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, err
	}

	if len(result.Member) == 0 {
		return nil, fmt.Errorf("member with key %s not found", key)
	}

	return result.Member[0], nil
}

func (s *MemberStorage) Save(ctx context.Context, member *Member) error {
	txn := s.client.NewTxn()

	_, err := txn.Mutate(ctx, &api.Mutation{
		CommitNow: true,
		SetJson:   member,
	})

	return err
}

// Update updates a member in the Dgraph Database
func (s *MemberStorage) Update(ctx context.Context, member *Member) error {
	txn := s.client.NewTxn()

	_, err := txn.Mutate(ctx, &api.Mutation{
		CommitNow: true,
		SetJson:   member,
	})

	return err
}

// Delete deletes a member from the Dgraph Database
func (s *MemberStorage) Delete(ctx context.Context, member *Member) error {
	txn := s.client.NewTxn()

	_, err := txn.Mutate(ctx, &api.Mutation{
		CommitNow:  true,
		DeleteJson: member,
	})

	return err
}
