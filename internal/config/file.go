package config

import (
	"errors"
	"os"
	"path/filepath"
)

func Read(path string) (spec Spec, err error) {
	f, err := os.Open(path)
	if err != nil {
		return Spec{}, err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	spec, err = Parse(f)
	if err != nil {
		return Spec{}, err
	}
	base, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return Spec{}, err
	}
	resolve := func(value string) string {
		if value == "" || filepath.IsAbs(value) {
			return value
		}
		return filepath.Join(base, value)
	}
	spec.Media.ISO = resolve(spec.Media.ISO)
	for i, value := range spec.Media.ExtraISOs {
		spec.Media.ExtraISOs[i] = resolve(value)
	}
	spec.Install.Template = resolve(spec.Install.Template)
	return spec, nil
}
