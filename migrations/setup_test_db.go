package migrations

import (
	"context"
	"database/sql"
	"embed"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratePg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed *.sql
var migrationFS embed.FS

func SetupTestDB(t *testing.T) (*sql.DB, context.Context) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections")),
	)

	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, pgContainer.Terminate(ctx))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	RunMigrations(t, db)

	return db, ctx

}

func GetMigrator(db *sql.DB) (*migrate.Migrate, error) {
	driver, err := migratePg.WithInstance(db, &migratePg.Config{})
	if err != nil {
		return nil, err
	}

	sourceDriver, err := iofs.New(migrationFS, ".")
	if err != nil {
		return nil, err
	}

	return migrate.NewWithInstance(
		"iofs",       
		sourceDriver, 
		"test-db",   
		driver,       
	)
}

func RunMigrations(t *testing.T, db *sql.DB) {
	migrator, err := GetMigrator(db)
	require.NoError(t, err)

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
}
