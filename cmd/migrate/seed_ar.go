package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	citiesFile    = "data/cities.json"
	terminalsFile = "data/terminals.json"
)

type cityRow struct {
	PostalCode string `json:"postal_code"`
	Name       string `json:"name"`
}

// terminalRow es una terminal de data/terminals.json. ExternalName es su nombre
// en backend-terminales y se usa para obtener el external_terminal_id.
type terminalRow struct {
	PostalCode   string `json:"postal_code"`
	Name         string `json:"name"`
	ExternalName string `json:"external_name"`
}

// externalTerminal es un ítem de GET /terminal/ de backend-terminales.
type externalTerminal struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"terminal"`
}

// seedArgentina inserta las ciudades y terminales de los JSON de data/ y les
// asigna el external_terminal_id de backend-terminales. Si no puede obtener las
// terminales de backend-terminales no modifica la base.
//
// Las ciudades existentes se dejan como están. Una terminal se crea solo si su
// ciudad no tenía ninguna; si ya existe (mismo código postal y nombre) solo se
// le completa el external_terminal_id cuando no lo tiene.
func (m migrator) seedArgentina() error {
	var cities []cityRow
	if err := readJSON(filepath.Join(m.dir, citiesFile), &cities); err != nil {
		return err
	}
	var terminals []terminalRow
	if err := readJSON(filepath.Join(m.dir, terminalsFile), &terminals); err != nil {
		return err
	}

	external, err := fetchExternalTerminals(m.terminalsURL)
	if err != nil {
		return fmt.Errorf("no se pudieron cargar las terminales de backend-terminales (%s): %w", m.terminalsURL, err)
	}
	externalIDs := make(map[string]uuid.UUID, len(external))
	for _, t := range external {
		externalIDs[normalizeName(t.Name)] = t.UUID
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

		// Las ciudades que ya tenían terminales no reciben nuevas: pueden estar
		// cargadas con otro nombre.
		var withTerminals []string
		if err := tx.Raw(`SELECT DISTINCT postal_code FROM bus_terminal`).Scan(&withTerminals).Error; err != nil {
			return err
		}
		hadTerminals := make(map[string]bool, len(withTerminals))
		for _, pc := range withTerminals {
			hadTerminals[pc] = true
		}

		var (
			newTerminals int
			linked       int
			skipped      []string // la ciudad ya tenía otra terminal
			notFound     []string // sin coincidencia en backend-terminales
			taken        []string // el external_terminal_id ya lo usa otra terminal
		)
		for _, t := range terminals {
			var row struct {
				UUID       uuid.UUID
				ExternalID *uuid.UUID `gorm:"column:external_terminal_id"`
			}
			err := tx.Raw(`SELECT uuid, external_terminal_id FROM bus_terminal
				WHERE postal_code = ? AND upper(name) = upper(?) LIMIT 1`, t.PostalCode, t.Name).Scan(&row).Error
			if err != nil {
				return err
			}
			if row.UUID == uuid.Nil {
				if hadTerminals[t.PostalCode] {
					skipped = append(skipped, t.Name)
					continue
				}
				if err := tx.Raw(`INSERT INTO bus_terminal (postal_code, name) VALUES (?, ?)
					RETURNING uuid, external_terminal_id`, t.PostalCode, t.Name).Scan(&row).Error; err != nil {
					return fmt.Errorf("terminal %s (%s): %w", t.Name, t.PostalCode, err)
				}
				newTerminals++
			}

			if row.ExternalID != nil {
				linked++
				continue
			}
			id, ok := externalIDs[normalizeName(t.ExternalName)]
			if t.ExternalName == "" || !ok {
				notFound = append(notFound, t.Name)
				continue
			}
			var inUse bool
			if err := tx.Raw(`SELECT EXISTS (SELECT 1 FROM bus_terminal WHERE external_terminal_id = ?)`, id).
				Scan(&inUse).Error; err != nil {
				return err
			}
			if inUse {
				taken = append(taken, t.Name)
				continue
			}
			if err := tx.Exec(`UPDATE bus_terminal SET external_terminal_id = ? WHERE uuid = ?`, id, row.UUID).Error; err != nil {
				return fmt.Errorf("terminal %s (%s): %w", t.Name, t.PostalCode, err)
			}
			linked++
		}

		log.Printf("ciudades: %d nuevas de %d", newCities, len(cities))
		log.Printf("terminales: %d nuevas de %d", newTerminals, len(terminals))
		log.Printf("external_terminal_id: %d cargados, %d sin cargar (backend-terminales devolvió %d terminales)",
			linked, len(notFound)+len(taken), len(external))
		logNames("sin coincidencia en backend-terminales", notFound)
		logNames("su external_terminal_id ya lo usa otra terminal", taken)
		logNames("omitidas porque la ciudad ya tenía otra terminal", skipped)
		return nil
	})
}

func fetchExternalTerminals(baseURL string) ([]externalTerminal, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(strings.TrimRight(baseURL, "/") + "/terminal/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /terminal/ respondió %s", resp.Status)
	}
	var terminals []externalTerminal
	if err := json.NewDecoder(resp.Body).Decode(&terminals); err != nil {
		return nil, fmt.Errorf("respuesta inválida de GET /terminal/: %w", err)
	}
	if len(terminals) == 0 {
		return nil, fmt.Errorf("GET /terminal/ no devolvió terminales")
	}
	return terminals, nil
}

var accents = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

// normalizeName compara nombres sin distinguir mayúsculas, tildes ni espacios extra.
func normalizeName(s string) string {
	return strings.Join(strings.Fields(accents.Replace(strings.ToLower(s))), " ")
}

func logNames(title string, names []string) {
	if len(names) == 0 {
		return
	}
	log.Printf("%s (%d):", title, len(names))
	for _, n := range names {
		log.Printf("  - %s", n)
	}
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
