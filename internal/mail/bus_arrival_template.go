package mail

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed BusArrivalTemplate.html
var busArrivalTemplateHTML string

// BusArrivalEmailData is passed to the bus arrival email template.
// AppURL and MapsURL are optional: if empty, the button is not shown.
// LogoURL is optional: if empty, the isotype is not shown.
type BusArrivalEmailData struct {
	SiteName      string
	LicensePlate string
	TerminalName  string
	Bay         string
	AppURL        string
	MapsURL       string
	LogoURL       string
}

func RenderBusArrivalHTML(data BusArrivalEmailData) (string, error) {
	tmpl, err := template.New("bus_arrival").Parse(busArrivalTemplateHTML)
	if err != nil {
		return "", fmt.Errorf("mail: parse bus arrival template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("mail: execute bus arrival template: %w", err)
	}
	return buf.String(), nil
}
