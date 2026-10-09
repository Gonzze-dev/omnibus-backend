// Command migrate aplica las migraciones SQL de migrations/ y registra cuáles
// se ejecutaron en la tabla omnibus_schema_migrations.
//
// Uso:
//
//	go run ./cmd/migrate [-dir migrations] <up|down|status|baseline|seed>
//
//	up        Base vacía: aplica schema.sql y marca todas las migraciones.
//	          Base existente: aplica las migraciones pendientes en orden.
//	down      Revierte la última migración aplicada usando su .down.sql.
//	status    Lista las migraciones y si están aplicadas.
//	baseline  Marca todas las migraciones como aplicadas sin ejecutarlas
//	          (para bases creadas antes de existir este comando).
//	seed      Ejecuta seed.sql (datos de ejemplo).
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
	upSuffix        = ".up.sql"
	downSuffix      = ".down.sql"
	// sentinelTable indica que la base ya tiene el esquema de la aplicación.
	sentinelTable = "users"
)

func main() {
	dir := flag.String("dir", "migrations", "directorio con los archivos .sql")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "uso: migrate [-dir migrations] <up|down|status|baseline|seed>")
		flag.PrintDefaults()
	}
	flag.Parse()

	cmd := "up"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	db, err := database.OpenPostgres(config.Load().DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	m := migrator{db: db, dir: *dir}
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
	default:
		flag.Usage()
		os.Exit(2)
	}
	if err != nil {
		log.Fatal(err)
	}
}

type migrator struct {
	db  *gorm.DB
	dir string
}

func (m migrator) ensureTable() error {
	return m.db.Exec(`CREATE TABLE IF NOT EXISTS ` + migrationsTable + ` (
		name       VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ  NOT NULL DEFAULT now()
	)`).Error
}

// migrations devuelve los nombres (sin .up.sql) ordenados alfabéticamente.
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
			return fmt.Errorf("la base ya tiene tablas pero no hay migraciones registradas en %s; "+
				"si está al día ejecutá `migrate baseline`", migrationsTable)
		}
		log.Printf("base vacía: aplicando %s", schemaFile)
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
		log.Println("no hay migraciones pendientes")
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
		log.Println("no hay migraciones aplicadas")
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
		mark := "pendiente"
		if done[name] {
			mark = "aplicada"
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
	log.Printf("%d migraciones marcadas como aplicadas", len(all))
	return nil
}

// execFile ejecuta un archivo en su propia transacción y, si name no está vacío,
// lo registra como migración aplicada.
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
	log.Printf("ejecutando %s", file)
	// Sin argumentos pgx usa el protocolo simple, que admite varias sentencias.
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
