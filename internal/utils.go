package internal

import (
	"net"
	"os"
	"path/filepath"
)

func listen(path string) (net.Listener, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}

	if err := os.RemoveAll(path); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(path, os.ModePerm); err != nil {
		listener.Close()

		return nil, err
	}

	return listener, nil
}
