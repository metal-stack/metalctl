package cmd

import (
	"io/ioutil"
	"os"
	"os/exec"
)

func edit(id string, getFunc func(id string) ([]byte, error), updateFunc func(filename string) error) error {
	editor, ok := os.LookupEnv("EDITOR")
	if !ok {
		editor = "vi"
	}

	tmpfile, err := ioutil.TempFile("", "metalctl*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())
	content, err := getFunc(id)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tmpfile.Name(), content, os.ModePerm)
	if err != nil {
		return err
	}
	editCommand := exec.Command(editor, tmpfile.Name())
	editCommand.Stdout = os.Stdout
	editCommand.Stdin = os.Stdin
	editCommand.Stderr = os.Stderr
	err = editCommand.Run()
	if err != nil {
		return err
	}
	return updateFunc(tmpfile.Name())
}
