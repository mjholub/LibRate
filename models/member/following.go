package member

import (
	"context"
	"fmt"
)

func (m *Member) IsFollowing(ctx context.Context, memberID uint32) (bool, error) {
	return false, fmt.Errorf("IsFollowing not implemented yet")
}
