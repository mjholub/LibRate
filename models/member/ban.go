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
	Error     BanStatusType = "error"
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
	switch target := target.(type) {
	case Member:
		if err := storage.QueryRowx("SELECT ban_status FROM members WHERE id = $1", target.ID).Scan(&b); err != nil {
			return Error
		}
		return b
	case Device:
		if err := storage.QueryRowx("SELECT ban_status FROM devices WHERE id = $1", target.ID).Scan(&b); err != nil {
			return Error
		}
		return b
	case net.IP:
		// here's where I tnink we'll need a separate table for bans so that querying that data is easier
		// especially since every ban always needs an offending member's ID/webfinger
		return None
	default:
		return None
	}
}
