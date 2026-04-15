//go:build integration

package helpers

import (
	"bytes"
	"embed"
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

// Second / Millisecond re-export time constants so test files need only one import.
const (
	Second      = time.Second
	Millisecond = time.Millisecond
)

// Fixture is a generic test scenario loaded from a multi-document YAML file.
type Fixture[T any] struct {
	Scenario string `yaml:"scenario"`
	Input    T      `yaml:"input"`
	Expected any    `yaml:"expected"`
}

// LoadFixtures decodes all YAML documents in the embedded file at path into a slice.
// Panics if the file is missing or any document fails to parse — fixture errors are
// programming errors that should never reach the test runner.
func LoadFixtures[T any](fs embed.FS, path string) []Fixture[T] {
	data, err := fs.ReadFile(path)
	if err != nil {
		panic("load fixtures " + path + ": " + err.Error())
	}
	return LoadFixturesFromBytes[T](data)
}

// LoadFixturesFromBytes decodes all YAML documents in data into a slice.
func LoadFixturesFromBytes[T any](data []byte) []Fixture[T] {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var fixtures []Fixture[T]
	for {
		var f Fixture[T]
		if err := dec.Decode(&f); err != nil {
			if err == io.EOF {
				break
			}
			panic("decode fixture: " + err.Error())
		}
		fixtures = append(fixtures, f)
	}
	return fixtures
}
