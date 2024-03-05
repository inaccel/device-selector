package pci

import (
	"os"
	"path/filepath"
	"strings"
)

func Devices() ([]Device, error) {
	dirEntries, err := os.ReadDir("/sys/bus/pci/devices")
	if err != nil {
		return nil, err
	}
	var devices []Device
	for _, dirEntry := range dirEntries {
		devices = append(devices, Device(dirEntry.Name()))
	}
	return devices, nil
}

type Device string

func (device Device) Class() (string, error) {
	name, err := filepath.EvalSymlinks(filepath.Join(device.Path(), "class"))
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (device Device) Device() (string, error) {
	name, err := filepath.EvalSymlinks(filepath.Join(device.Path(), "device"))
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (device Device) Driver() (string, error) {
	return filepath.EvalSymlinks(filepath.Join(device.Path(), "driver"))
}

func (device Device) DriverOverride(s string) error {
	name, err := filepath.EvalSymlinks(filepath.Join(device.Path(), "driver_override"))
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

func (device Device) IommuGroup() (string, error) {
	return filepath.EvalSymlinks(filepath.Join(device.Path(), "iommu_group"))
}

func (device Device) NumaNode() (string, error) {
	name, err := filepath.EvalSymlinks(filepath.Join(device.Path(), "numa_node"))
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (device Device) OfNode() (string, error) {
	return filepath.EvalSymlinks(filepath.Join(device.Path(), "of_node"))
}

func (device Device) Path() string {
	if strings.Contains(string(device), string(filepath.Separator)) {
		return string(device)
	}
	return filepath.Join("/sys/bus/pci/devices", string(device))
}

func (device Device) Revision() (string, error) {
	name, err := filepath.EvalSymlinks(filepath.Join(device.Path(), "revision"))
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (device Device) String() string {
	if !strings.Contains(string(device), string(filepath.Separator)) {
		return string(device)
	}
	return filepath.Base(string(device))
}

func (device Device) SubsystemDevice() (string, error) {
	name, err := filepath.EvalSymlinks(filepath.Join(device.Path(), "subsystem_device"))
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (device Device) SubsystemVendor() (string, error) {
	name, err := filepath.EvalSymlinks(filepath.Join(device.Path(), "subsystem_vendor"))
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (device Device) Vendor() (string, error) {
	name, err := filepath.EvalSymlinks(filepath.Join(device.Path(), "vendor"))
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
