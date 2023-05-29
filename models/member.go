package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"codeberg.org/mjh/LibRate/cfg"
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

func NewMemberStorage(conf cfg.DgraphConfig) (*MemberStorage, *grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	conn, err := client.ConnectToService(ctx, conf.Host, conf.GRPCPort)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Dgraph: %v", err)
	}
	dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	if dgraphClient == nil {
		return nil, nil, fmt.Errorf("failed to create Dgraph client")
	}

	return &MemberStorage{
			client: dgraphClient,
		},
		conn,
		nil
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

	memberJSON, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal member: %v", err)
	}

	_, err = txn.Mutate(ctx, &api.Mutation{
		CommitNow: true,
		SetJson:   memberJSON,
	})

	return err
}

// Update updates a member in the Dgraph Database
func (s *MemberStorage) Update(ctx context.Context, member *Member) error {
	txn := s.client.NewTxn()

	memberJSON, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal member: %v", err)
	}

	_, err = txn.Mutate(ctx, &api.Mutation{
		CommitNow: true,
		SetJson:   memberJSON,
	})

	return err
}

// Delete deletes a member from the Dgraph Database
func (s *MemberStorage) Delete(ctx context.Context, member *Member) error {
	txn := s.client.NewTxn()

	memberJSON, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal member: %v", err)
	}

	_, err = txn.Mutate(ctx, &api.Mutation{
		CommitNow:  true,
		DeleteJson: memberJSON,
	})

	return err
}
