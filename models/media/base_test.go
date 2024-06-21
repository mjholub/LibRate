package media

import (
	"context"
	"testing"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustConnectDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dsn := db.CreateDsn(&cfg.TestConfig.DBConfig)

	db, err := pgxpool.New(ctx, dsn)
	assert.NoError(t, err)
	return db
}

func createZerologNopPtr() *zerolog.Logger {
	l := zerolog.Nop()
	return &l
}

func TestNewStorage(t *testing.T) {
	db := mustConnectDB(t)
	require.NotNil(t, db)
	l := createZerologNopPtr()
	require.NotNil(t, l)
	require.IsType(t, &zerolog.Logger{}, l)
	ks := NewKeywordStorage(mustConnectDB(t), createZerologNopPtr())
	require.NotNil(t, ks)
	ps := NewPeopleStorage(mustConnectDB(t), createZerologNopPtr())
	require.NotNil(t, ps)

	store := NewStorage(db, l, ks, ps)
	require.Equalf(t, db, store.db, "db")
	require.Equalf(t, l, store.Log, "Log")
	require.Equalf(t, ks, store.ks, "ks")
	require.Equalf(t, ps, store.Ps, "Ps")
}

func TestStorage_Get(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx context.Context
		id  uuid.UUID
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantMedia *Media
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			gotMedia, err := ms.Get(tt.args.ctx, tt.args.id)
			tt.assertion(t, err)
			assert.Equal(t, tt.wantMedia, gotMedia)
		})
	}
}

func TestStorage_GetImagePath(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx context.Context
		id  uuid.UUID
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantPath  string
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			gotPath, err := ms.GetImagePath(tt.args.ctx, tt.args.id)
			tt.assertion(t, err)
			assert.Equal(t, tt.wantPath, gotPath)
		})
	}
}

func TestStorage_GetKind(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx context.Context
		id  uuid.UUID
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			got, err := ms.GetKind(tt.args.ctx, tt.args.id)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetGenres(t *testing.T) {
	type args struct {
		ms      *Storage
		ctx     context.Context
		kind    string
		all     bool
		columns []string
	}
	tests := []struct {
		name      string
		args      args
		want      []G
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGenres(tt.args.ms, tt.args.ctx, tt.args.kind, tt.args.all, tt.args.columns...)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStorage_GetGenre(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx  context.Context
		kind string
		lang string
		name string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantGenre *Genre
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			gotGenre, err := ms.GetGenre(tt.args.ctx, tt.args.kind, tt.args.lang, tt.args.name)
			tt.assertion(t, err)
			assert.Equal(t, tt.wantGenre, gotGenre)
		})
	}
}

func TestStorage_GetDetails(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx       context.Context
		mediaKind string
		id        uuid.UUID
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      interface{}
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			got, err := ms.GetDetails(tt.args.ctx, tt.args.mediaKind, tt.args.id)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStorage_GetRandom(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx            context.Context
		count          int
		blacklistKinds []string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantMwks  map[uuid.UUID]string
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			gotMwks, err := ms.GetRandom(tt.args.ctx, tt.args.count, tt.args.blacklistKinds...)
			tt.assertion(t, err)
			assert.Equal(t, tt.wantMwks, gotMwks)
		})
	}
}

func TestStorage_Add(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx   context.Context
		props *Media
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMediaID *uuid.UUID
		assertion   assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			gotMediaID, err := ms.Add(tt.args.ctx, tt.args.props)
			tt.assertion(t, err)
			assert.Equal(t, tt.wantMediaID, gotMediaID)
		})
	}
}

func TestStorage_AddCreators(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx      context.Context
		id       uuid.UUID
		creators []Person
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			tt.assertion(t, ms.AddCreators(tt.args.ctx, tt.args.id, tt.args.creators))
		})
	}
}

func TestStorage_GetAll(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	tests := []struct {
		name      string
		fields    fields
		want      []*interface{}
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			got, err := ms.GetAll()
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStorage_Update(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx     context.Context
		key     string
		value   interface{}
		mediaID uuid.UUID
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			tt.assertion(t, ms.Update(tt.args.ctx, tt.args.key, tt.args.value, tt.args.mediaID))
		})
	}
}

func TestStorage_Delete(t *testing.T) {
	type fields struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
	type args struct {
		ctx     context.Context
		mediaID uuid.UUID
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				db:  tt.fields.db,
				Log: tt.fields.Log,
				ks:  tt.fields.ks,
				Ps:  tt.fields.Ps,
			}
			tt.assertion(t, ms.Delete(tt.args.ctx, tt.args.mediaID))
		})
	}
}
