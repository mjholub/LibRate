package models

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
)

type Member struct {
	UUID       string `json:"_key,omitempty"`
	PassHash string `json:"passhash"`
	MemberName string `json:"membername"`
	Email    string `json:"email"`
	RegTimestamp int64 `json:"regdate"`
}

type MemberInput struct {
	MemberName string `json:"membername"`
	Email    string `json:"email"`
}

type RegisterInput struct {
	Email    string `json:"email"`
	MemberName string `json:"membername"`
	Password string `json:"password"`
	PasswordConfirm string `json:"passwordconfirm"`
}

type LoginInput struct {
	Email    string `json:"email"`
	MemberName string `json:"membername"`
	Password string `json:"password"`
}

type RegLoginInput interface {
	RegisterInput | LoginInput
}

// MemberStorer implements the authboss.ServerStorer interface
type MemberStorer struct {
}

func NewMemberStorer() *MemberStorer {
	return &MemberStorer{}
}

func (m *Member) GetPID() string {
	return m.ID
}

func (m *Member) PutPID(key string) {
	m.ID = key
}

// GetMemberByID retrieves a member by ID from the ArangoDB database
func (ms *MemberStorer) Load(ctx context.Context, key string) (member *Member, err error) {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		return nil, err
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
		Authentication: driver.BasicAuthentication("root", ""),
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	db, err := client.Database(ctx, "your_database_name")
	if err != nil {
		return nil, err
	}

	members, err := db.Collection(ctx, "members")
	if err != nil {
		return nil, err
	}
	_ = members
	// FIXME: remove or find use

	query := fmt.Sprintf("FOR u IN members FILTER u._key == '%d' RETURN u", id)
	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	if cursor.HasMore() {
		_, err := cursor.ReadDocument(ctx, &member)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("member not found")
	}

	return &member, nil
}

// SaveMember saves a new member to the ArangoDB database
func (ms *MemberStorer) Save(ctx context.Context, member *Member) error {
	// connect and auth to ArangoDB, get database and collection
	auth, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to ArangoDB: %w", err)
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     auth,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	db, err := client.Database(ctx, os.Getenv("ARANGO_DB_NAME"))
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	members, err := db.Collection(ctx, "members")
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// insert member into Collection
	meta, err := members.CreateDocument(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to create member: %w", err)
	}

	member.ID = meta.Key

	return nil
}

// UpdateMember updates a member in the ArangoDB Database
func UpdateMember(id string, input MemberInput) (*Member, error) {
	// connect and auth to ArandoDB, get database and collection
	auth, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to ArangoDB: %w", err)
	}
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     auth,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	db, err := client.Database(ctx, os.Getenv("ARANGO_DB_NAME"))
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	members, err := db.Collection(ctx, "members")
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// update member in collection
	meta, err, _ := members.UpdateDocument(ctx, id, input)
	if err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}

	member := Member{
		ID:       meta.Key,
		MemberName: input.MemberName,
		Email:    input.Email,
	}

	return &member, nil
}


