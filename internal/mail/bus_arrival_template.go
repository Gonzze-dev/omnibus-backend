package mail

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed BusArrivalTemplate.html
var busArrivalTemplateHTML string

// BusArrivalEmailData se pasa al template del mail de llegada del colectivo.
// AppURL y MapsURL son opcionales: si están vacíos no se muestra el botón.
type BusArrivalEmailData struct {
	SiteName      string
	LicensePatent string
	TerminalName  string
	Anden         string
	AppURL        string
	MapsURL       string
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
