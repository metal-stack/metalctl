package applier

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v3"
)

// Applier can be used to apply entities
type Applier[C any, U any, R any] struct {
	from string
	fs   afero.Fs
}

// Appliable must be implemented in order to apply entities
type Appliable[C any, U any, R any] interface {
	// Create tries to create the entity with the given request, if it already exists it does NOT return an error but nil for both return arguments.
	// if the creation was successful it returns the success response.
	Create(rq C) (*R, error)
	// Update tries to update the entity with the given request.
	// if the update was successful it returns the success response.
	Update(rq U) (R, error)
}

func NewApplier[C any, U any, R any](from string) (*Applier[C, U, R], error) {
	fs := afero.NewOsFs()

	switch from {
	case "":
		return nil, fmt.Errorf("from must not be empty")
	case "-":
	default:
		exists, err := afero.Exists(fs, from)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("file does not exist: %s", from)
		}
	}

	return &Applier[C, U, R]{
		from: from,
		fs:   fs,
	}, nil
}

func (a *Applier[C, U, R]) Apply(appliable Appliable[C, U, R]) ([]R, error) {
	docs, err := readYAML[C](a.fs, a.from)
	if err != nil {
		return nil, err
	}

	result := []R{}

	for index := range docs {
		createDoc := docs[index]

		created, err := appliable.Create(createDoc)
		if err != nil {
			return nil, fmt.Errorf("error creating entity: %w", err)
		}

		if created != nil {
			result = append(result, *created)
			continue
		}

		updateDoc, err := readYAMLIndex[U](a.fs, a.from, index)
		if err != nil {
			return nil, err
		}

		updated, err := appliable.Update(updateDoc)
		if err != nil {
			return nil, fmt.Errorf("error updating entity: %w", err)
		}

		result = append(result, updated)
	}

	return result, nil
}

// readFrom will either read from stdin (-) or a file path an marshall from yaml to data
func readYAML[D any](fs afero.Fs, from string) ([]D, error) {
	reader, err := getReader(fs, from)
	if err != nil {
		return nil, err
	}

	var docs []D

	dec := yaml.NewDecoder(reader)

	for {
		data := new(D)

		err := dec.Decode(&data)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("decode error: %w", err)
		}

		docs = append(docs, *data)
	}

	return docs, nil
}

// readFrom will either read from stdin (-) or a file path an marshall from yaml to data
func readYAMLIndex[D any](fs afero.Fs, from string, index int) (D, error) {
	emptyD := new(D)

	reader, err := getReader(fs, from)
	if err != nil {
		return *emptyD, err
	}

	dec := yaml.NewDecoder(reader)

	count := 0
	for {
		data := new(D)

		err := dec.Decode(data)
		if errors.Is(err, io.EOF) {
			return *emptyD, fmt.Errorf("index not found in document: %d", index)
		}
		if err != nil {
			return *emptyD, fmt.Errorf("decode error: %w", err)
		}

		if count == index {
			return *data, nil
		}

		count++
	}
}

func getReader(fs afero.Fs, from string) (io.Reader, error) {
	var reader io.Reader
	var err error
	switch from {
	case "-":
		reader = os.Stdin
	default:
		reader, err = fs.Open(from)
		if err != nil {
			return nil, fmt.Errorf("unable to open %q: %w", from, err)
		}
	}

	return reader, nil
}
