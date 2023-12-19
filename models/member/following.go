package member

import (
	"context"
	"fmt"
)

func (m *PgMemberStorage) IsFollowing(ctx context.Context, memberID int) (bool, error) {
	return false, fmt.Errorf("IsFollowing not implemented yet")
}
