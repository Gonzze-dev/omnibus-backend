// Command migrate aplica las migraciones SQL de migrations/ y registra cuáles
// se ejecutaron en la tabla omnibus_schema_migrations.
//
// Uso:
//
//	go run ./cmd/migrate [-dir migrations] <up|down|status|baseline|seed|seed-ar>
//
//	up        Base vacía: aplica schema.sql y marca todas las migraciones.
//	          Base existente: aplica las migraciones pendientes en orden.
//	down      Revierte la última migración aplicada usando su .down.sql.
//	status    Lista las migraciones y si están aplicadas.
//	baseline  Marca todas las migraciones como aplicadas sin ejecutarlas
//	          (para bases creadas antes de existir este comando).
//	seed      Ejecuta seed.sql (datos de ejemplo).
//	seed-ar   Carga ciudades y terminales de Argentina desde data/cities.json
//	          y data/terminals.json. Es idempotente: no duplica ni pisa datos.
package main

import (
	"encoding/json"
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
	citiesFile      = "data/cities.json"
	terminalsFile   = "data/terminals.json"
	upSuffix        = ".up.sql"
	downSuffix      = ".down.sql"
	// sentinelTable indica que la base ya tiene el esquema de la aplicación.
	sentinelTable = "users"
)

func main() {
	dir := flag.String("dir", "migrations", "directorio con los archivos .sql")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "uso: migrate [-dir migrations] <up|down|status|baseline|seed|seed-ar>")
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
	case "seed-ar":
		err = m.seedArgentina()
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

// seedRow es una ciudad o una terminal de los JSON de data/.
type seedRow struct {
	PostalCode string `json:"postal_code"`
	Name       string `json:"name"`
}

// seedArgentina inserta las ciudades y terminales de los JSON de data/.
// Las ciudades existentes (mismo código postal) se dejan como están y solo se
// agregan terminales a ciudades que todavía no tienen ninguna.
func (m migrator) seedArgentina() error {
	var cities []seedRow
	if err := readJSON(filepath.Join(m.dir, citiesFile), &cities); err != nil {
		return err
	}
	var terminals []seedRow
	if err := readJSON(filepath.Join(m.dir, terminalsFile), &terminals); err != nil {
		return err
	}

	return m.db.Transaction(func(tx *gorm.DB) error {
		newCities := 0
		for _, c := range cities {
			res := tx.Exec(`INSERT INTO city (postal_code, name) VALUES (?, ?) ON CONFLICT (postal_code) DO NOTHING`,
				c.PostalCode, c.Name)
			if res.Error != nil {
				return fmt.Errorf("ciudad %s (%s): %w", c.Name, c.PostalCode, res.Error)
			}
			newCities += int(res.RowsAffected)
		}

		// Las ciudades que ya tenían terminales se saltean: pueden estar cargadas
		// con otro nombre (p. ej. vinculadas a backend-terminales).
		var withTerminals []string
		if err := tx.Raw(`SELECT DISTINCT postal_code FROM bus_terminal`).Scan(&withTerminals).Error; err != nil {
			return err
		}
		skip := make(map[string]bool, len(withTerminals))
		for _, pc := range withTerminals {
			skip[pc] = true
		}

		newTerminals := 0
		for _, t := range terminals {
			if skip[t.PostalCode] {
				continue
			}
			if err := tx.Exec(`INSERT INTO bus_terminal (postal_code, name) VALUES (?, ?)`,
				t.PostalCode, t.Name).Error; err != nil {
				return fmt.Errorf("terminal %s (%s): %w", t.Name, t.PostalCode, err)
			}
			newTerminals++
		}

		log.Printf("ciudades: %d nuevas de %d; terminales: %d nuevas de %d",
			newCities, len(cities), newTerminals, len(terminals))
		return nil
	})
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
