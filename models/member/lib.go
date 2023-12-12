package member

import (
	"context"
	"database/sql"
	"net"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
)

// Member holds the core information about a member
type (
	Member struct {
		ID               int            `json:"-" db:"id"`
		UUID             uuid.UUID      `json:"uuid,omitempty" db:"uuid"`
		PassHash         string         `json:"-" db:"passhash"`
		MemberName       string         `json:"memberName" db:"nick,unique" validate:"required,alphanumunicode,min=3,max=30"`
		DisplayName      sql.NullString `json:"displayName,omitempty" db:"display_name"`
		Email            string         `json:"email" db:"email" validate:"required,email"`
		Bio              sql.NullString `json:"bio,omitempty" db:"bio"`
		Active           bool           `json:"active" db:"active"`
		Roles            pq.StringArray `json:"roles,omitempty" db:"roles"`
		RegTimestamp     time.Time      `json:"regdate" db:"reg_timestamp"`
		ProfilePicID     sql.NullInt64  `json:"-" db:"profilepic_id"`
		ProfilePicSource string         `json:"profile_pic,omitempty" db:"-"`
		Homepage         sql.NullString `json:"homepage,omitempty" db:"homepage"`
		IRC              sql.NullString `json:"irc,omitempty" db:"irc"`
		XMPP             sql.NullString `json:"xmpp,omitempty" db:"xmpp"`
		Matrix           sql.NullString `json:"matrix,omitempty" db:"matrix"`
		Visibility       string         `json:"visibility" db:"visibility"`
		FollowingURI     string         `json:"following_uri" db:"following_uri"` // URI for getting the following list of this account
		FollowersURI     string         `json:"followers_uri" db:"followers_uri"` // URI for getting the followers list of this account
		SessionTimeout   sql.NullInt64  `json:"-" db:"session_timeout"`
		PublicKeyPem     string         `jsonld:"publicKeyPem,omitempty" json:"-" db:"public_key_pem"`
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
		Read(ctx context.Context, key string, keyNames ...string) (*Member, error)
		// Check checks if a member with the given email or nickname already exists
		Check(ctx context.Context, email, nickname string) (bool, error)
		Update(ctx context.Context, member *Member) error
		Delete(ctx context.Context, member *Member) error
		GetID(ctx context.Context, key string) (int, error)
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
)

func NewSQLStorage(client *sqlx.DB, log *zerolog.Logger, conf *cfg.Config) *PgMemberStorage {
	return &PgMemberStorage{client: client, log: log, config: conf}
}
