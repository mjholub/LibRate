package member

import (
	"net"

	"github.com/jmoiron/sqlx"
)

const (
	Current   BanStatusType = "current"
	Expired   BanStatusType = "expired"
	Permanent BanStatusType = "permanent"
	None      BanStatusType = "none"
)

type (
	// see https://vladimir.varank.in/notes/2023/09/compile-time-safety-for-enumerations-in-go/
	BanStatusType string

	Ban interface {
		getStatus(target interface{}) BanStatusType
	}

	// BanStatus holds information about a ban
	// It can apply to a member, device or IP address/range
	BanStatus struct {
		Status BanStatusType `json:"status" db:"status"`
	}
)

func (b BanStatusType) getStatus(target interface{}, storage *sqlx.DB) BanStatusType {
	switch target.(type) {
	case Member:
		storage.QueryRowx("SELECT ban_status FROM members WHERE id = $1", target.(Member).ID).Scan(&b)
	case Device:
		storage.QueryRowx("SELECT ban_status FROM devices WHERE id = $1", target.(Device).ID).Scan(&b)
	case net.IP:
		// here's where I tnink we'll need a separate table for bans so that querying that data is easier
		// especially since every ban always needs an offending member's ID/webfinger
		return None
	default:
		return None
	}
	return b
}
