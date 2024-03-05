package pci

import (
	"os"
	"path/filepath"
	"strings"
)

func Drivers() ([]Driver, error) {
	dirEntries, err := os.ReadDir("/sys/bus/pci/drivers")
	if err != nil {
		return nil, err
	}
	var drivers []Driver
	for _, dirEntry := range dirEntries {
		drivers = append(drivers, Driver(dirEntry.Name()))
	}
	return drivers, nil
}

type Driver string

func (driver Driver) Bind(s string) error {
	name, err := filepath.EvalSymlinks(filepath.Join(driver.Path(), "bind"))
	if err != nil {
		return err
	}
	f, err := os.OpenFile(name, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(s); err != nil {
		return err
	}
	return nil
}

func (driver Driver) Path() string {
	if strings.Contains(string(driver), string(filepath.Separator)) {
		return string(driver)
	}
	return filepath.Join("/sys/bus/pci/drivers", string(driver))
}

func (driver Driver) String() string {
	if !strings.Contains(string(driver), string(filepath.Separator)) {
		return string(driver)
	}
	return filepath.Base(string(driver))
}

func (driver Driver) Unbind(s string) error {
	name, err := filepath.EvalSymlinks(filepath.Join(driver.Path(), "unbind"))
	if err != nil {
		return err
	}
	f, err := os.OpenFile(name, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(s); err != nil {
		return err
	}
	return nil
}
