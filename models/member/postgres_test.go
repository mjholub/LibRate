package member

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/cfg"

	"github.com/pashagolub/pgxmock/v3"
)

func TestBan(t *testing.T) {
	logger := zerolog.Nop()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	assert.Implements(t, mock, &pgxpool.Pool{})
	defer mock.Close()

	// FIXME: pgxpool.Pool doesn't implement pgxmock.PgxPoolIface (AsConn missing)
	// need to create own implementation with the methods we need
	storage := NewSQLStorage(nil, &mock, &logger, &cfg.TestConfig)
	ends := time.Now().Add(24 * time.Hour)
	_, mask, err := net.ParseCIDR("192.0.2.0/24")
	require.NoError(t, err)

	uuid, err := uuid.NewV4()
	require.NoErrorf(t, err, "failed to generate uuid: %v", err)

	bi := &BanInput{
		Reason:    "spam",
		Ends:      ends,
		CanAppeal: false,
		Mask:      mask,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO public.bans`).
		WithArgs(uuid, bi.Reason, bi.Ends, false, bi.Mask).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	user := &Member{
		UUID: uuid,
	}

	err = storage.Ban(context.Background(), user, bi)
	assert.NoErrorf(t, err, "failed to ban user: %v", err)
}
