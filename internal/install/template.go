package install

import (
	"bytes"
	"fmt"
	"text/template"
)

type Identity struct {
	Hostname  string
	MAC       string
	PublicKey string
}

func Render(source string, identity Identity) (string, error) {
	t, err := template.New("kickstart").Option("missingkey=error").Parse(source)
	if err != nil {
		return "", fmt.Errorf("parse kickstart template: %w", err)
	}
	var output bytes.Buffer
	if err := t.Execute(&output, identity); err != nil {
		return "", fmt.Errorf("render kickstart template: %w", err)
	}
	return output.String(), nil
}
