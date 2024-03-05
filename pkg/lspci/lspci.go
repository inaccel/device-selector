package lspci

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/inaccel/device-selector/pkg/sysfs/bus/pci"
	"github.com/sirupsen/logrus"
)

type PCIDevice struct {
	Slot       string
	Class      string
	Vendor     string
	Device     string
	SVendor    string
	SDevice    string
	PhySlot    string
	Rev        string
	ProgIf     string
	Driver     string
	NUMANode   string
	DTNode     string
	IOMMUGroup string
}

func ListAll() []PCIDevice {
	var pciDevices []PCIDevice
	sysfsBusPciDevices, err := pci.Devices()
	if err != nil {
		logrus.Debug(err)
	}
	for _, sysfsBusPciDevice := range sysfsBusPciDevices {
		classRaw, err := sysfsBusPciDevice.Class()
		if err != nil {
			logrus.Debug(err)
			continue
		}
		deviceRaw, err := sysfsBusPciDevice.Device()
		if err != nil {
			logrus.Debug(err)
			continue
		}
		vendorRaw, err := sysfsBusPciDevice.Vendor()
		if err != nil {
			logrus.Debug(err)
			continue
		}

		slot := sysfsBusPciDevice.String()
		class := strings.TrimPrefix(classRaw, "0x")[:4]
		vendor := strings.TrimPrefix(vendorRaw, "0x")
		device := strings.TrimPrefix(deviceRaw, "0x")
		var sVendor string
		var sDevice string
		if subsystemVendorRaw, err := sysfsBusPciDevice.SubsystemVendor(); err != nil {
			logrus.Debug(err)
		} else {
			sVendor = strings.TrimPrefix(subsystemVendorRaw, "0x")
			if subsystemDeviceRaw, err := sysfsBusPciDevice.SubsystemDevice(); err != nil {
				logrus.Debug(err)
			} else {
				sDevice = strings.TrimPrefix(subsystemDeviceRaw, "0x")
			}
		}
		var phySlot string
		if sysfsBusPciSlots, err := pci.Slots(); err != nil {
			logrus.Debug(err)
		} else {
			for _, sysfsBusPciSlot := range sysfsBusPciSlots {
				address, err := sysfsBusPciSlot.Address()
				if err != nil {
					logrus.Debug(err)
					continue
				}
				if strings.HasPrefix(sysfsBusPciDevice.String(), address) {
					phySlot = sysfsBusPciSlot.String()
					break
				}
			}
		}
		var rev string
		if revisionRaw, err := sysfsBusPciDevice.Revision(); err != nil {
			logrus.Debug(err)
		} else {
			rev = strings.TrimPrefix(revisionRaw, "0x")
		}
		progIf := strings.TrimPrefix(classRaw, "0x")[4:]
		var driver string
		if driverRaw, err := sysfsBusPciDevice.Driver(); err != nil {
			logrus.Debug(err)
		} else {
			driver = filepath.Base(driverRaw)
		}
		var numaNode string
		if numaNodeRaw, err := sysfsBusPciDevice.NumaNode(); err != nil {
			logrus.Debug(err)
		} else {
			numaNode = numaNodeRaw
		}
		var dtNode string
		if ofNodeRaw, err := sysfsBusPciDevice.OfNode(); err != nil {
			logrus.Debug(err)
		} else {
			dtNode = ofNodeRaw
		}
		var iommuGroup string
		if iommuGroupRaw, err := sysfsBusPciDevice.IommuGroup(); err != nil {
			logrus.Debug(err)
		} else {
			iommuGroup = filepath.Base(iommuGroupRaw)
		}

		pciDevices = append(pciDevices, PCIDevice{
			slot,
			class,
			vendor,
			device,
			sVendor,
			sDevice,
			phySlot,
			rev,
			progIf,
			driver,
			numaNode,
			dtNode,
			iommuGroup,
		})
	}
	return pciDevices
}

func (pciDevice *PCIDevice) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Slot:\t%s\n", pciDevice.Slot)
	fmt.Fprintf(&b, "Class:\t%s\n", pciDevice.Class)
	fmt.Fprintf(&b, "Vendor:\t%s\n", pciDevice.Vendor)
	fmt.Fprintf(&b, "Device:\t%s\n", pciDevice.Device)
	if pciDevice.SVendor != "" {
		fmt.Fprintf(&b, "SVendor:\t%s\n", pciDevice.SVendor)
	}
	if pciDevice.SDevice != "" {
		fmt.Fprintf(&b, "SDevice:\t%s\n", pciDevice.SDevice)
	}
	if pciDevice.PhySlot != "" {
		fmt.Fprintf(&b, "PhySlot:\t%s\n", pciDevice.PhySlot)
	}
	if pciDevice.Rev != "" {
		fmt.Fprintf(&b, "Rev:\t%s\n", pciDevice.Rev)
	}
	if pciDevice.ProgIf != "" {
		fmt.Fprintf(&b, "ProgIf:\t%s\n", pciDevice.ProgIf)
	}
	if pciDevice.Driver != "" {
		fmt.Fprintf(&b, "Driver:\t%s\n", pciDevice.Driver)
	}
	if pciDevice.NUMANode != "" {
		fmt.Fprintf(&b, "NUMANode:\t%s\n", pciDevice.NUMANode)
	}
	if pciDevice.DTNode != "" {
		fmt.Fprintf(&b, "DTNode:\t%s\n", pciDevice.DTNode)
	}
	if pciDevice.IOMMUGroup != "" {
		fmt.Fprintf(&b, "IOMMUGroup:\t%s\n", pciDevice.IOMMUGroup)
	}
	return b.String()
}
