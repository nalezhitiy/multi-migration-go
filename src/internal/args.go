package internal

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

type MigrationType string

const (
	MigrationSQL MigrationType = "sql"
	MigrationGo  MigrationType = "go"
)

type CLIArgs struct {
	DBName        string
	Command       string
	Args          []string
	MigrationType MigrationType
}

var supportedCommands = map[string]struct{}{
	"up":        {},
	"up-by-one": {},
	"up-to":     {},
	"down":      {},
	"down-to":   {},
	"redo":      {},
	"reset":     {},
	"status":    {},
	"version":   {},
	"create":    {},
	"fix":       {},
}

// ParseArgs parses and validates CLI arguments.
func ParseArgs(args []string) (CLIArgs, error) {
	fs := flag.NewFlagSet("migrator", flag.ContinueOnError)
	dbName := fs.String("db", "", "Database name")

	if err := fs.Parse(args); err != nil {
		return CLIArgs{}, err
	}

	positional := fs.Args()
	if len(positional) == 0 {
		return CLIArgs{}, fmt.Errorf(
			"command is required: up, down, status, create, etc.",
		)
	}

	command := positional[0]
	commandArgs := positional[1:]

	if _, ok := supportedCommands[command]; !ok {
		return CLIArgs{}, fmt.Errorf("unsupported command %q", command)
	}

	dbNameValue := strings.TrimSpace(*dbName)
	if dbNameValue == "" {
		return CLIArgs{}, fmt.Errorf(
			"database is required: use -db=postgres-main, -db=mariadb-main or -db=clickhouse-main",
		)
	}

	if dbNameValue == "all" {
		return CLIArgs{}, fmt.Errorf(
			"database 'all' is not supported; specify one database with -db",
		)
	}

	result := CLIArgs{
		DBName:  dbNameValue,
		Command: command,
		Args:    commandArgs,
	}

	if command == "create" {
		if len(commandArgs) != 1 {
			return CLIArgs{}, fmt.Errorf(
				"create requires exactly one migration filename, e.g. create users.sql",
			)
		}

		filename := commandArgs[0]
		ext := strings.ToLower(filepath.Ext(filename))

		switch ext {
		case "", ".sql":
			result.MigrationType = MigrationSQL
		//case ".go":
		//	result.MigrationType = MigrationGo
		default:
			return CLIArgs{}, fmt.Errorf(
				"unsupported migration extension %q: use .sql",
				ext,
			)
		}

		name := strings.TrimSuffix(filename, filepath.Ext(filename))
		if name == "" || name == "." || name == ".." ||
			filepath.Base(name) != name {
			return CLIArgs{}, fmt.Errorf(
				"invalid migration filename %q",
				filename,
			)
		}

		// Goose receives the migration name without the extension.
		result.Args = []string{name}
	}

	return result, nil
}
