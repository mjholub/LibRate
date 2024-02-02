package member

import (
	"context"
	"database/sql"
	"net"
	"sync"
	"time"

	"golang.org/x/text/language"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
)

// Member holds the core information about a member
type (
	Member struct {
		ID       int       `json:"-" db:"id"`
		UUID     uuid.UUID `json:"uuid,omitempty" db:"uuid"`
		PassHash string    `json:"-" db:"passhash"`
		// MemberName != webfinger
		MemberName string `json:"memberName" db:"nick,unique" validate:"required,alphanumunicode,min=3,max=30" example:"lain"`
		// email like
		Webfinger        string         `json:"webfinger" db:"webfinger,unique" validate:"required,email" example:"lain@librate.club"`
		DisplayName      sql.NullString `json:"displayName,omitempty" db:"display_name" example:"Lain Iwakura"`
		Email            string         `json:"email" db:"email" validate:"required,email" example:"lain@wired.jp"`
		Bio              sql.NullString `json:"bio,omitempty" db:"bio" example:"Wherever you go, everyone is connected."`
		Active           bool           `json:"active" db:"active" example:"true"`
		Roles            pq.StringArray `json:"roles,omitempty" db:"roles" example:"[\"admin\", \"moderator\"]"`
		RegTimestamp     time.Time      `json:"regdate" db:"reg_timestamp" example:"2020-01-01T00:00:00Z"`
		ProfilePicID     sql.NullInt64  `json:"-" db:"profilepic_id"`
		ProfilePicSource string         `json:"profile_pic,omitempty" db:"-" example:"/static/img/profile/lain.jpg"`
		Visibility       string         `json:"visibility" db:"visibility" example:"followers_only"`
		FollowingURI     string         `json:"following_uri" db:"following_uri"` // URI for getting the following list of this account
		FollowersURI     string         `json:"followers_uri" db:"followers_uri"` // URI for getting the followers list of this account
		SessionTimeout   sql.NullInt64  `json:"-" db:"session_timeout"`
		PublicKeyPem     string         `jsonld:"publicKeyPem,omitempty" json:"publicKeyPem" db:"public_key_pem"`
		CustomFields     pgtype.JSONB   `json:"customFields,omitempty" db:"custom_fields"`
	}

	// TODO: move the password here
	Preferences struct {
		// NOTE:not using DB tags since the structure for this is actually flat in the DB
		UX              UXPreferences              `json:"ux,omitempty"`
		PrivacySecurity PrivacySecurityPreferences `json:"privsec,omitempty"`
	}

	// Theme is stored in localStorage since some people might prefer to have a different theme on
	// different devices
	UXPreferences struct {
		Locale language.Tag `json:"locale,omitempty" db:"locale"`
		// everything is calculated relative to the maximum scale of 0-100
		RatingScaleLower int16 `json:"rating_scale_lower,omitempty" db:"rating_scale_lower" validate:"ltfield=RatinScaleUpper",min=0,max=1" default:"1"`
		RatingScaleUpper int16 `json:"rating_scale_upper,omitempty" db:"rating_scale_upper" validate:"min=2,max=100" default:"10"`
	}

	PrivacySecurityPreferences struct {
		MessageAutohideWords pq.StringArray `json:"message_autohide_words,omitempty" db:"message_autohide_words"`
		// domain names alone, without protocol etc.
		MutedInstances      pq.StringArray `json:"muted_instances,omitempty" db:"muted_instances"`
		AutoAcceptFollow    bool           `json:"auto_accept_follow,omitempty" db:"auto_accept_follow" default:"true"`
		LocallySearchable   bool           `json:"locally_searchable,omitempty" db:"locally_searchable" default:"true"`
		FederatedSearchable bool           `json:"searchable_to_federated,omitempty" db:"searchable_to_federated" default:"true"`
		RobotsSearchable    bool           `json:"robots_searchable,omitempty" db:"robots_searchable" default:"false"`
	}

	Device struct {
		FriendlyName sql.NullString `json:"friendlyName,omitempty" db:"friendly_name"`
		// KnownIPs is used to improve the security in case of logging in from unknown locations
		KnownIPs  []net.IP  `json:"knownIPs,omitempty" db:"known_ips"`
		LastLogin time.Time `json:"lastLogin,omitempty" db:"last_login"`
		BanStatus BanStatus `json:"banStatus,omitempty" db:"ban_status"`
		ID        uuid.UUID `json:"id" db:"id,unique,notnull"`
	}

	FollowBlockRequest struct {
		ID        int64     `json:"id,omitempty" db:"id"`
		Requester string    `json:"requester,omitempty" db:"requester_webfinger"`
		Target    string    `json:"target" db:"target_webfinger"`
		Reblogs   bool      `json:"reblogs" db:"reblogs" default:"true" sql:"-"` // only used for follow requests
		Notify    bool      `json:"notify" db:"notify" default:"true" sql:"-"`
		Created   time.Time `json:"created,omitempty" db:"created"`
	}

	FollowResponse struct {
		// when checking status, we treat not_found as not following
		// but when creating a request, we treat not_found as target account not existing
		Status     string     `json:"status" db:"status", validate:"required,oneof=accepted failed pending blocked already_following not_found"`
		ID         int64      `json:"id,omitempty" db:"id"`
		Reblogs    bool       `json:"reblogs,omitempty" db:"reblogs"`
		Notify     bool       `json:"notify,omitempty" db:"notify"`
		Error      error      `json:"-"`
		AcceptTime *time.Time `json:"acceptTime" db:"created"`
	}
	// received follow request
	FollowRequestIn struct {
		ID        int64      `json:"id" db:"id"`
		Requester string     `json:"requester" db:"requester_webfinger"`
		Created   *time.Time `json:"created" db:"created"`
	}

	FollowRequestGroup struct {
		Sent     []FollowResponse  `json:"sent" db:"sent"`
		Received []FollowRequestIn `json:"received" db:"received"`
	}

	// Follower represents a follower-followee relationship
	Follower struct {
		ID       uint32    `json:"id" db:"id"`
		Created  time.Time `json:"created" db:"created"`
		Follower string    `json:"follower" db:"follower"`
		Followee string    `json:"followee" db:"followee"`
	}

	// Input holds the information required to create a new member account
	Input struct {
		MemberName string `json:"membername"`
		Email      string `json:"email"`
		Password   string `json:"password"`
	}

	// BanInput is used to ban a member
	BanInput struct {
		Reason    string    `json:"reason" db:"reason" validate:"required" example:"spam"`
		Ends      time.Time `json:"ends" db:"ends" validate:"required" example:"2038-01-16T00:00:00Z"`
		CanAppeal bool      `json:"canAppeal" db:"can_appeal" validate:"required" example:"true"`
		// usage: https://pkg.go.dev/net#ParseCIDR
		Mask     *net.IPNet `json:"mask" db:"mask"`
		BannedBy string     `json:"bannedBy" db:"banned_by"`
	}

	// BanStatus is used to retrieve the ban details
	BanStatus struct {
		BanInput
		MemberUUID uuid.UUID `json:"memberUUID" db:"member_uuid"`
		// Occurrence is the n-th time a ban has been issued
		Occurrence int16     `json:"occurrence" db:"occurrence"`
		Started    time.Time `json:"started" db:"started"`
	}

	// TODO: debload this interface
	Storer interface {
		Save(ctx context.Context, member *Member) error
		Read(ctx context.Context, key string, keyNames ...string) (*Member, error)
		HasRole(ctx context.Context, name, role string, exact bool) bool
		Ban(ctx context.Context, member *Member, input *BanInput) error
		Unban(ctx context.Context, member *Member) error
		VerifyViewability(ctx context.Context, viewer, viewee string) (bool, error)
		// Check checks if a member with the given email or nickname already exists
		Check(ctx context.Context, email, nickname string) (bool, error)
		Update(ctx context.Context, member *Member) error
		Delete(ctx context.Context, member *Member) error
		GetID(ctx context.Context, key string) (int, error)
		GetPassHash(email, login string) (string, error)
		CreateSession(ctx context.Context, member *Member) (string, error)
		IsBlocked(ctx context.Context, fr *FollowBlockRequest) (blocked bool, err error)
		FollowStorer
		Exporter
	}

	Exporter interface {
		Export(ctx context.Context, memberName, format string) ([]byte, error)
	}

	FollowStorer interface {
		RequestFollow(ctx context.Context, fr *FollowBlockRequest) FollowResponse
		//	UpdateFollow(ctx context.Context, fr *FollowBlockRequest) error
		AcceptFollow(ctx context.Context, accepter string, requestID int64) error
		CancelFollow(ctx context.Context, canceler string, requestID int64) error
		RejectFollow(ctx context.Context, rejecter string, requestID int64) error
		GetFollowStatus(ctx context.Context, follower, followee string) FollowResponse
		GetFollowRequests(ctx context.Context, member string, reqType string) (any, error)
		RemoveFollower(ctx context.Context, follower, followee string) error
	}

	PgMemberStorage struct {
		client        *sqlx.DB
		newClient     *pgxpool.Pool
		log           *zerolog.Logger
		config        *cfg.Config
		nicknameCache []string
		cacheMutex    sync.RWMutex
	}
)

func NewSQLStorage(client *sqlx.DB, newClient *pgxpool.Pool, log *zerolog.Logger, conf *cfg.Config) *PgMemberStorage {
	return &PgMemberStorage{client: client, newClient: newClient, log: log, config: conf}
}
