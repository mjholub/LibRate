package crypt

import (
	"fmt"
	"os"
)

// CreateFiles creates the temporary directory and database file
func CreateFile(path string) (dbFile *os.File, err error) {
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			dbFile, err = os.Create(path)
			if err != nil {
				return nil, fmt.Errorf("could not create database file: %v", err)
			}
			return dbFile, nil
		}
		return nil, fmt.Errorf("could not stat database file: %v", err)
	}
	return os.OpenFile(path, os.O_RDWR, 0o600)
}
