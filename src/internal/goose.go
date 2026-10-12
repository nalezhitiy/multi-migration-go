package internal

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func openDatabase(
	ctx context.Context,
	cfg DBConfig,
) (*sql.DB, error) {
	var driver string

	switch cfg.Dialect {
	case "postgres":
		driver = "pgx"
	case "mysql":
		driver = "mysql"
	case "clickhouse":
		driver = "clickhouse"
	default:
		return nil, fmt.Errorf(
			"unsupported database dialect %q",
			cfg.Dialect,
		)
	}

	db, err := sql.Open(driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return db, nil
}

func newGooseProvider(
	ctx context.Context,
	cfg DBConfig,
	migrationsDir string,
	tableName string,
) (*goose.Provider, error) {
	db, err := openDatabase(ctx, cfg)
	if err != nil {
		return nil, err
	}

	var dialect goose.Dialect

	switch cfg.Dialect {
	case "postgres":
		dialect = goose.DialectPostgres
	case "mysql":
		dialect = goose.DialectMySQL
	case "clickhouse":
		dialect = goose.DialectClickHouse
	default:
		db.Close()
		return nil, fmt.Errorf(
			"unsupported database dialect %q",
			cfg.Dialect,
		)
	}

	provider, err := goose.NewProvider(
		dialect,
		db,
		os.DirFS(migrationsDir),
		goose.WithTableName(tableName),
	)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create goose provider: %w", err)
	}

	return provider, nil
}

func runMigrationCommand(
	ctx context.Context,
	provider *goose.Provider,
	args CLIArgs,
) error {
	switch args.Command {
	case "up":
		_, err := provider.Up(ctx)
		return err

	case "up-by-one":
		_, err := provider.UpByOne(ctx)
		return err

	case "up-to":
		version, err := parseVersionArg(args.Args)
		if err != nil {
			return err
		}

		_, err = provider.UpTo(ctx, version)
		return err

	case "down":
		_, err := provider.Down(ctx)
		return err

	case "down-to":
		version, err := parseVersionArg(args.Args)
		if err != nil {
			return err
		}

		_, err = provider.DownTo(ctx, version)
		return err

	case "status":
		statuses, err := provider.Status(ctx)
		if err != nil {
			return err
		}

		for _, status := range statuses {
			fmt.Printf(
				"version=%d migration=%s state=%s\n",
				status.Source.Version,
				status.Source.Path,
				status.State,
			)
		}

		return nil

	case "version":
		version, err := provider.GetDBVersion(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("Current database version: %d\n", version)
		return nil

	case "redo":
		version, err := provider.GetDBVersion(ctx)
		if err != nil {
			return err
		}

		if version == 0 {
			return nil
		}

		if _, err := provider.ApplyVersion(ctx, version, false); err != nil {
			return fmt.Errorf("rollback migration %d: %w", version, err)
		}

		if _, err := provider.ApplyVersion(ctx, version, true); err != nil {
			return fmt.Errorf("reapply migration %d: %w", version, err)
		}

		return nil

	case "reset":
		_, err := provider.DownTo(ctx, 0)
		return err
	default:
		return fmt.Errorf(
			"command %q is not implemented yet",
			args.Command,
		)
	}
}

func parseVersionArg(args []string) (int64, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf(
			"command requires exactly one version argument",
		)
	}

	version, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || version < 0 {
		return 0, fmt.Errorf(
			"invalid migration version %q",
			args[0],
		)
	}

	return version, nil
}

func Execute(
	ctx context.Context,
	args CLIArgs,
	cfg *Config,
) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if args.DBName == "" {
		return fmt.Errorf("database is required: use -db=<database-name>")
	}

	if args.DBName == "all" {
		return fmt.Errorf(
			"database 'all' is not supported; specify one database with -db",
		)
	}

	dbCfg, ok := cfg.Databases[args.DBName]
	if !ok {
		return fmt.Errorf(
			"database %q not found in config",
			args.DBName,
		)
	}

	migrationsDir := filepath.Join(
		cfg.Migrations.DirName,
		args.DBName,
	)

	switch args.Command {
	case "create":
		if len(args.Args) != 1 {
			return fmt.Errorf(
				"create requires exactly one migration name",
			)
		}

		if err := os.MkdirAll(migrationsDir, 0755); err != nil {
			return fmt.Errorf(
				"create migrations directory: %w",
				err,
			)
		}

		migrationType := string(args.MigrationType)
		if migrationType == "" {
			migrationType = "sql"
		}

		return goose.Create(
			nil,
			migrationsDir,
			args.Args[0],
			migrationType,
		)

	case "fix":
		return goose.Fix(migrationsDir)
	}

	provider, err := newGooseProvider(
		ctx,
		dbCfg,
		migrationsDir,
		cfg.Migrations.TableName,
	)
	if err != nil {
		return fmt.Errorf(
			"database %q: %w",
			args.DBName,
			err,
		)
	}
	defer provider.Close()

	if err := runMigrationCommand(ctx, provider, args); err != nil {
		return fmt.Errorf(
			"database %q: %w",
			args.DBName,
			err,
		)
	}

	return nil
}
