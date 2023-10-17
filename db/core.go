package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/avast/retry-go/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db/bootstrap"
)

func CreateDsn(dsn *cfg.DBConfig) string {
	switch dsn.SSL {
	case "require", "verify-ca", "verify-full", "disable":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database, dsn.SSL)
		return data
	case "prefer":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database, "require")
		return data
	case "unknown":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database)
		return data
	default:
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database)
		return data
	}
}

func Connect(conf *cfg.Config, noSubProcess bool) (*sqlx.DB, error) {
	// create a whitelist of launch commands to avond arbitrary code execution

	data := CreateDsn(&conf.DBConfig)
	var db *sqlx.DB

	err := retry.Do(
		func() error {
			var err error
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			db, err = sqlx.ConnectContext(ctx, conf.Engine, data)
			return err
		},
		retry.Attempts(5),
		retry.Delay(1*time.Second), // Delay between retries
		retry.OnRetry(func(n uint, _ error) {
			fmt.Printf("Attempt %d failed; retrying...", n)
			if !noSubProcess {
				err := launch(conf)
				if err != nil {
					return
				}
			}
		}),
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ConnectNeo4j(conf *cfg.Config) (neo4j.DriverWithContext, error) {
	dsn := fmt.Sprintf("bolt://%s:%d",
		conf.Host, conf.Port)
	auth := neo4j.BasicAuth(conf.User, conf.Password, "")
	neo4jConf := func(cf *config.Config) {
		cf.TelemetryDisabled = true
	}
	return neo4j.NewDriverWithContext(dsn,
		auth, neo4jConf)
}

func launch(conf *cfg.Config) error {
	if conf.StartCmd == "skip" {
		return nil
	}

	whitelist := []string{
		// standalone
		"pg_ctl start -D /var/lib/postgresql/data",
		"pg_ctl start -D /var/lib/postgresql/data -l /var/lib/postgresql/data/logfile",
		// sysvinit
		"service postgresql start",
		// systemd
		"systemctl start postgresql",
		// openrc
		"rc-service postgresql start",
		"/etc/init.d/postgresql start",
		// s6
		"s6-svc -u /var/run/s6/services/postgresql",
		// supervisord
		"supervisorctl start postgresql",
		// runit
		"sv start postgresql",
		// launchd
		"launchctl start homebrew.mxcl.postgresql",
		// containerized
		"docker run --name postgresql -e POSTGRES_PASSWORD=postgres -d postgres",
		"podman run --name postgresql -e POSTGRES_PASSWORD=postgres -d postgres",
		"kubectl run postgresql --image=postgres --env=\"POSTGRES_PASSWORD=postgres\"",
		"docker-compose up -d postgresql",
	}

	rootcmds := []string{
		"sudo",
		"su -c",
		"doas",
		"please",
	}

	// combine the two lists
	allcmds := lo.FlatMap(whitelist, func(s string, _ int) []string {
		return lo.Map(rootcmds, func(s2 string, _ int) string {
			withRoot := fmt.Sprintf("%s %s", s2, s)
			// also add the command without root, e.g. for containerized environments
			withoutRoot := s
			return withRoot + "\n" + withoutRoot
		})
	})

	if conf.StartCmd == "" {
		err := errors.New(`no start command provided for database server\n
				Please provide one under a 'start_cmd' key in the database section of the config file.\n
				Waiting for manual start of database server for 10 seconds.
				`)
		fmt.Printf("failed to start postgresql service: %v", err)
		time.Sleep(10 * time.Second)
	}

	if !lo.Contains(allcmds, conf.StartCmd) {
		err := errors.New(`start command not in whitelist\n
				Aborting to prevent arbitrary code execution.\n
				Please use a command from the whitelist provided in db/core.go
				`)
		return fmt.Errorf("failed to start postgresql service: %w", err)
	} else if conf.StartCmd != "" && lo.Contains(allcmds, conf.StartCmd) {
		cmd := exec.CommandContext(context.Background(), conf.StartCmd+" &")
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to start postgresql service: %w", err)
		}
	}

	return nil
}

func createExtension(db *sqlx.DB, extName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx,
		fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s" SCHEMA public;`, extName))
	if err != nil {
		return fmt.Errorf("failed to create extension %s: %w", extName, err)
	}

	return nil
}

func InitDB(conf *cfg.Config, noSubProcess, exitAfter bool, log *zerolog.Logger) error {
	if exitAfter {
		// nolint:revive
		defer func() {
			fmt.Println("Database initialized. Exiting...")
			os.Exit(0)
		}()
	}
	db, err := Connect(conf, noSubProcess)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS public; SET search_path TO public;`)
	if err != nil {
		return fmt.Errorf("failed to create public schema: %w", err)
	}
	err = bootstrap.Schemas(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created schemas")

	// set up the extensions
	var mu sync.Mutex
	errChan := make(chan error)
	mu.Lock()
	extNames := []string{"pgcrypto", "uuid-ossp", "pg_trgm", "sequential_uuids"}
	log.Info().Msg("Creating extensions...")
	for i := range extNames {
		go func(i int) {
			errChan <- createExtension(db, extNames[i])
		}(i)
	}
	for i := 0; i < len(extNames); i++ {
		err = <-errChan
		if err != nil {
			return err
		}
	}
	close(errChan)
	mu.Unlock()
	log.Info().Msg("Created extensions")
	err = bootstrap.CDN(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to create cdn tables: %w", err)
	}
	log.Info().Msg("Created cdn tables")
	err = bootstrap.Places(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created places tables")
	err = bootstrap.MediaCore(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Creating media tables: 1/2...")
	err = bootstrap.People(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created people tables")
	err = bootstrap.Roles(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created roles tables")
	err = bootstrap.MediaCreators(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created media_creators tables")
	err = bootstrap.PeopleMeta(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created people_meta tables")
	err = bootstrap.Media(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Creating media tables complete")
	err = bootstrap.CreatorGroups(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created creator_groups tables")
	err = bootstrap.AlbumArtists(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created album_artists tables")
	err = bootstrap.Studio(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created studio tables")
	err = bootstrap.Books(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created books tables")
	err = bootstrap.Cast(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created cast tables")
	err = bootstrap.Members(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created members tables")
	err = bootstrap.MembersProfilePic(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created members profilepic tables")
	err = bootstrap.Review(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created review tables")

	return nil
}
