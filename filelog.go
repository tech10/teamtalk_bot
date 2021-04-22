package main

import (
	"errors"
	"os"
	"path/filepath"
)

func dir_create(dir string) error {
	if len(dir) > 1 {
		//Check directory existance and create it.
		_, derr := os.Stat(dir)
		exists := os.IsExist(derr)
		if !exists {
			err := os.MkdirAll(dir, 0755)
			return err
		}
	}
	return nil
}

func file_write(fname, data string) error {
	if data == "" {
		return errors.New("Empty data.")
	}
	if err := dir_create(filepath.Dir(fname)); err != nil {
		return err
	}
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	n, err := file.WriteString(data)
	if n < len(data) || err != nil {
		if err != nil {
			return err
		}
		return errors.New("Incomplete data write.")
	}
	return nil
}
