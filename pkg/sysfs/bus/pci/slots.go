package pci

import (
	"os"
	"path/filepath"
	"strings"
)

func Slots() ([]Slot, error) {
	dirEntries, err := os.ReadDir("/sys/bus/pci/slots")
	if err != nil {
		return nil, err
	}
	var slots []Slot
	for _, dirEntry := range dirEntries {
		slots = append(slots, Slot(dirEntry.Name()))
	}
	return slots, nil
}

type Slot string

func (slot Slot) Address() (string, error) {
	name, err := filepath.EvalSymlinks(filepath.Join(slot.Path(), "address"))
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (slot Slot) Path() string {
	if strings.Contains(string(slot), string(filepath.Separator)) {
		return string(slot)
	}
	return filepath.Join("/sys/bus/pci/slots", string(slot))
}

func (slot Slot) String() string {
	if !strings.Contains(string(slot), string(filepath.Separator)) {
		return string(slot)
	}
	return filepath.Base(string(slot))
}
