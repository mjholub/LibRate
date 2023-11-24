package member

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/go-ap/activitypub"
	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models/static"
)

// TODO: convert to enum if this becomes possible in Go
// nolint:gochecknoglobals // roles
const (
	admin uint8 = iota
	mod
	regular
	creator
)

// Member holds the core information about a member
type (
	Member struct {
		// TODO: convert to int64, postgres doesn't support unsigned by default anyway
		ID             uint32                 `json:"id" db:"id"`
		UUID           string                 `json:"_key,omitempty" db:"uuid"`
		PassHash       string                 `json:"passhash" db:"passhash"`
		MemberName     string                 `json:"memberName" db:"nick"` // i.e. @nick@instance
		DisplayName    sql.NullString         `json:"displayName,omitempty" db:"display_name"`
		Email          string                 `json:"email" db:"email" validate:"required,email"`
		Bio            sql.NullString         `json:"bio,omitempty" db:"bio"`
		Active         bool                   `json:"active" db:"active"`
		Roles          []uint8                `json:"roles,omitempty" db:"roles"`
		RegTimestamp   time.Time              `json:"regdate" db:"reg_timestamp"`
		ProfilePic     *static.Image          `json:"profilepic,omitempty" db:"profilepic_id"`
		Homepage       sql.NullString         `json:"homepage,omitempty" db:"homepage"`
		IRC            sql.NullString         `json:"irc,omitempty" db:"irc"`
		XMPP           sql.NullString         `json:"xmpp,omitempty" db:"xmpp"`
		Matrix         sql.NullString         `json:"matrix,omitempty" db:"matrix"`
		Visibility     string                 `json:"visibility" db:"visibility"`
		Followers      activitypub.Collection `json:"followers,omitempty" db:"followers"`
		SessionTimeout sql.NullInt64          `json:"sessionTimeout,omitempty" db:"sessionTimeout"`
		// TODO: add database migration
		PublicKeyPem string `jsonld:"publicKeyPem,omitempty" json:"publicKeyPem,omitempty" db:"public_key_pem"`
	}

	Device struct {
		FriendlyName sql.NullString `json:"friendlyName,omitempty" db:"friendly_name"`
		// KnownIPs is used to improve the security in case of logging in from unknown locations
		KnownIPs  []net.IP  `json:"knownIPs,omitempty" db:"known_ips"`
		LastLogin time.Time `json:"lastLogin,omitempty" db:"last_login"`
		BanStatus BanStatus `json:"banStatus,omitempty" db:"ban_status"`
		ID        uuid.UUID `json:"id" db:"id,unique,notnull"`
	}

	FollowRequest struct {
		ID        int64  `json:"id" db:"id"`
		ActorID   string `json:"actor_id" db:"actor_id"`
		FollowsID string `json:"follows_id" db:"follows_id"`
	}

	// Follower represents a follower-followee relationship
	Follower struct {
		ID       uint32 `json:"id" db:"id"`
		Follower uint32 `json:"follower" db:"follower"`
		Followee uint32 `json:"followee" db:"followee"`
	}

	// Input holds the information required to create a new member account
	Input struct {
		MemberName string `json:"membername"`
		Email      string `json:"email"`
		Password   string `json:"password"`
	}

	MemberStorer interface {
		Save(ctx context.Context, member *Member) error
		Read(ctx context.Context, keyName, key string) (*Member, error)
		// Check checks if a member with the given email or nickname already exists
		Check(ctx context.Context, email, nickname string) (bool, error)
		Update(ctx context.Context, member *Member) error
		Delete(ctx context.Context, member *Member) error
		GetID(ctx context.Context, key string) (uint32, error)
		GetPassHash(email, login string) (string, error)
		// GetSessionTimeout retrieves the preferred timeout until the session expires,
		// represented as number of seconds
		GetSessionTimeout(ctx context.Context, memberID int, deviceID uuid.UUID) (timeout int, err error)
		LookupDevice(ctx context.Context, deviceID uuid.UUID) error
		CreateSession(ctx context.Context, member *Member) (string, error)
		RequestFollow(ctx context.Context, fr *FollowRequest) error
	}

	PgMemberStorage struct {
		client        *sqlx.DB
		log           *zerolog.Logger
		config        *cfg.Config
		nicknameCache []string
		cacheMutex    sync.RWMutex
	}

	Neo4jMemberStorage struct {
		client neo4j.DriverWithContext
		log    *zerolog.Logger
		config *cfg.Config
	}
)

func NewSQLStorage(client *sqlx.DB, log *zerolog.Logger, conf *cfg.Config) *PgMemberStorage {
	return &PgMemberStorage{client: client, log: log, config: conf}
}

func NewNeo4jStorage(client neo4j.DriverWithContext, log *zerolog.Logger, conf *cfg.Config) Neo4jMemberStorage {
	return Neo4jMemberStorage{client: client, log: log, config: conf}
}

func mapRoleCodesToStrings(roles []uint8) []string {
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		switch role {
		case admin:
			roleStrings[i] = "admin"
		case mod:
			roleStrings[i] = "mod"
		case regular:
			roleStrings[i] = "regular"
		case creator:
			roleStrings[i] = "creator"
		default:
			panic(fmt.Sprintf("invalid role: %d", role))
		}
	}
	return roleStrings
}
