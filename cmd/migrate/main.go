// Command migrate applies the SQL migrations in migrations/ and records which
// ones ran in the omnibus_schema_migrations table.
//
// Usage:
//
//	go run ./cmd/migrate [-dir migrations] <up|down|status|baseline|seed|seed-ar|seed-platforms>
//
//	up        Empty database: applies schema.sql and marks every migration.
//	          Existing database: applies the pending migrations in order.
//	down      Reverts the last applied migration using its .down.sql.
//	status    Lists the migrations and whether they are applied.
//	baseline  Marks every migration as applied without running it
//	          (for databases created before this command existed).
//	seed      Runs seed.sql (sample data).
//	seed-platforms
//	          Runs seed_platforms.sql: the bays of TERMINAL DE GUALEGUAYCHÚ
//	          (run seed-ar first). It is idempotent.
//	seed-ar   Loads the cities and terminals of Argentina from data/cities.json
//	          and data/terminals.json, and assigns their external_terminal_id
//	          by querying backend-terminales. It is idempotent.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"

	"tesina/backend/internal/config"
	"tesina/backend/internal/database"
)

const (
	migrationsTable = "omnibus_schema_migrations"
	schemaFile      = "schema.sql"
	seedFile        = "seed.sql"
	platformsFile   = "seed_platforms.sql"
	upSuffix        = ".up.sql"
	downSuffix      = ".down.sql"
	// sentinelTable means the database already has the application schema.
	sentinelTable = "users"
)

func main() {
	dir := flag.String("dir", "migrations", "directory with the .sql files")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: migrate [-dir migrations] <up|down|status|baseline|seed|seed-ar|seed-platforms>")
		flag.PrintDefaults()
	}
	flag.Parse()

	cmd := "up"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	cfg := config.Load()
	db, err := database.OpenPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	m := migrator{db: db, dir: *dir, terminalsURL: cfg.ExternalTerminalUpstreamURL}
	if err := m.ensureTable(); err != nil {
		log.Fatal(err)
	}

	switch cmd {
	case "up":
		err = m.up()
	case "down":
		err = m.down()
	case "status":
		err = m.status()
	case "baseline":
		err = m.baseline()
	case "seed":
		err = m.execFile(seedFile, "")
	case "seed-ar":
		err = m.seedArgentina()
	case "seed-platforms":
		err = m.execFile(platformsFile, "")
	default:
		flag.Usage()
		os.Exit(2)
	}
	if err != nil {
		log.Fatal(err)
	}
}

type migrator struct {
	db           *gorm.DB
	dir          string
	terminalsURL string // backend-terminales, for the seed-ar external_terminal_ids
}

func (m migrator) ensureTable() error {
	return m.db.Exec(`CREATE TABLE IF NOT EXISTS ` + migrationsTable + ` (
		name       VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ  NOT NULL DEFAULT now()
	)`).Error
}

// migrations returns the names (without .up.sql) sorted alphabetically.
func (m migrator) migrations() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(m.dir, "*"+upSuffix))
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, strings.TrimSuffix(filepath.Base(f), upSuffix))
	}
	sort.Strings(names)
	return names, nil
}

func (m migrator) applied() (map[string]bool, error) {
	var names []string
	if err := m.db.Table(migrationsTable).Pluck("name", &names).Error; err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	return set, nil
}

func (m migrator) up() error {
	all, err := m.migrations()
	if err != nil {
		return err
	}
	done, err := m.applied()
	if err != nil {
		return err
	}

	if len(done) == 0 {
		var exists bool
		if err := m.db.Raw(`SELECT to_regclass(?) IS NOT NULL`, sentinelTable).Scan(&exists).Error; err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("the database already has tables but no migrations are recorded in %s; "+
				"if it is up to date run `migrate baseline`", migrationsTable)
		}
		log.Printf("empty database: applying %s", schemaFile)
		return m.db.Transaction(func(tx *gorm.DB) error {
			if err := execFile(tx, m.dir, schemaFile); err != nil {
				return err
			}
			return record(tx, all)
		})
	}

	pending := 0
	for _, name := range all {
		if done[name] {
			continue
		}
		pending++
		if err := m.execFile(name+upSuffix, name); err != nil {
			return err
		}
	}
	if pending == 0 {
		log.Println("no pending migrations")
	}
	return nil
}

func (m migrator) down() error {
	var last string
	err := m.db.Table(migrationsTable).Order("name DESC").Limit(1).Pluck("name", &last).Error
	if err != nil {
		return err
	}
	if last == "" {
		log.Println("no applied migrations")
		return nil
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		if err := execFile(tx, m.dir, last+downSuffix); err != nil {
			return err
		}
		return tx.Exec(`DELETE FROM `+migrationsTable+` WHERE name = ?`, last).Error
	})
}

func (m migrator) status() error {
	all, err := m.migrations()
	if err != nil {
		return err
	}
	done, err := m.applied()
	if err != nil {
		return err
	}
	for _, name := range all {
		mark := "pending"
		if done[name] {
			mark = "applied"
		}
		fmt.Printf("%-10s %s\n", mark, name)
	}
	return nil
}

func (m migrator) baseline() error {
	all, err := m.migrations()
	if err != nil {
		return err
	}
	if err := record(m.db, all); err != nil {
		return err
	}
	log.Printf("%d migrations marked as applied", len(all))
	return nil
}

// execFile runs a file in its own transaction and, if name is not empty,
// records it as an applied migration.
func (m migrator) execFile(file, name string) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		if err := execFile(tx, m.dir, file); err != nil {
			return err
		}
		if name == "" {
			return nil
		}
		return record(tx, []string{name})
	})
}

func execFile(tx *gorm.DB, dir, file string) error {
	sql, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return err
	}
	log.Printf("running %s", file)
	// Without arguments pgx uses the simple protocol, which allows several statements.
	if err := tx.Exec(string(sql)).Error; err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	return nil
}

func record(tx *gorm.DB, names []string) error {
	for _, name := range names {
		err := tx.Exec(`INSERT INTO `+migrationsTable+` (name) VALUES (?) ON CONFLICT DO NOTHING`, name).Error
		if err != nil {
			return err
		}
	}
	return nil
}
