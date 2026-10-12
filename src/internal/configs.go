package internal

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/drone/envsubst/v2"
	"github.com/goccy/go-yaml"
)

type DBConfig struct {
	Dialect string `yaml:"dialect"`
	DSN     string `yaml:"dsn"`
}
type MigrationsConfig struct {
	DirName   string `yaml:"dir_name"`
	TableName string `yaml:"table_name"`
}
type Config struct {
	Migrations MigrationsConfig    `yaml:"migrations"`
	Databases  map[string]DBConfig `yaml:"databases"`
}

func LoadConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	processed, err := envsubst.EvalEnv(string(content))
	if err != nil {
		return nil, fmt.Errorf("substitute environment variables: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(processed), &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

var (
	databaseNamePattern  = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
	tableNamePattern     = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	unresolvedEnvPattern = regexp.MustCompile(
		`\$\{[^}]*\}|\$[a-zA-Z_][a-zA-Z0-9_]*`,
	)
)

var supportedDialects = map[string]struct{}{
	"postgres":   {},
	"mysql":      {},
	"clickhouse": {},
	//"mssql":      {},
	//"sqlite3":    {},
	//"turso":      {},
	//"vertica":    {},
	//"ydb":        {},
}

func (cfg *Config) Validate() error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	var problems []string

	// Validate migrations settings.
	if strings.TrimSpace(cfg.Migrations.DirName) == "" {
		problems = append(problems,
			"migrations.dir_name must not be empty")
	}

	tableName := strings.TrimSpace(cfg.Migrations.TableName)
	if tableName == "" {
		problems = append(problems,
			"migrations.table_name must not be empty")
	} else if !tableNamePattern.MatchString(tableName) {
		problems = append(problems,
			"migrations.table_name must contain only ASCII letters, digits and underscores, and must not start with a digit")
	}

	// Validate databases.
	if len(cfg.Databases) == 0 {
		problems = append(problems,
			"databases must contain at least one database")
	}

	for name, db := range cfg.Databases {
		if !databaseNamePattern.MatchString(name) {
			problems = append(problems,
				fmt.Sprintf("databases.%q has an invalid name", name))
		}

		dialect := strings.TrimSpace(db.Dialect)
		if dialect == "" {
			problems = append(problems,
				fmt.Sprintf("databases.%s.dialect must not be empty", name))
		} else if _, ok := supportedDialects[dialect]; !ok {
			problems = append(problems,
				fmt.Sprintf(
					"databases.%s.dialect %q is not supported; allowed: postgres, mysql, clickhouse",
					name,
					dialect,
				))
		}

		dsn := strings.TrimSpace(db.DSN)
		if dsn == "" {
			problems = append(problems,
				fmt.Sprintf("databases.%s.dsn must not be empty", name))
		} else if unresolvedEnvPattern.MatchString(dsn) {
			problems = append(problems,
				fmt.Sprintf(
					"databases.%s.dsn contains an unresolved environment variable",
					name,
				))
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf(
			"invalid configuration:\n- %s",
			strings.Join(problems, "\n- "),
		)
	}

	return nil
}

func (c *Config) HasDatabase(name string) bool {
	_, ok := c.Databases[name]
	return ok
}
