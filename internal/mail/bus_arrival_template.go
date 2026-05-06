package mail

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed BusArrivalTemplate.html
var busArrivalTemplateHTML string

type BusArrivalEmailData struct {
	SiteName      string
	LicensePatent string
	Anden         string
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
