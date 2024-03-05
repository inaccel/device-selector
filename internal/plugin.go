package internal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/inaccel/daemon/pkg/plugin"
	"github.com/inaccel/device-selector/pkg/lspci"
	"github.com/inaccel/device-selector/pkg/sysfs/bus/pci"
	"github.com/sirupsen/logrus"
	"github.com/u-root/u-root/pkg/kmodule"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	devicepluginv1beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	pluginregistrationv1 "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/kubevirt/pkg/util"
)

type pciHostDevicePlugin struct {
	ctx  context.Context
	path string

	pciHostDevice kubevirtv1.PciHostDevice
	plugin.Plugin
}

func NewPciHostDevicePlugin(ctx context.Context, pciHostDevice kubevirtv1.PciHostDevice) plugin.New {
	return func() plugin.Plugin {
		return newPciHostDevicePlugin(ctx, pciHostDevice)
	}
}

func newPciHostDevicePlugin(ctx context.Context, pciHostDevice kubevirtv1.PciHostDevice) plugin.Plugin {
	ctx, cancel := context.WithCancel(ctx)

	pciHostDevicePlugin := &pciHostDevicePlugin{
		ctx:  ctx,
		path: filepath.Join("/var/lib/kubelet/plugins_registry", pciHostDevice.PCIVendorSelector+".sock"),
	}

	pciHostDevicePlugin.pciHostDevice = pciHostDevice

	pciHostDevicePlugin.Plugin = plugin.Base(func() {
		if listener, err := listen(pciHostDevicePlugin.path); err == nil {
			go func() {
				<-ctx.Done()

				listener.Close()
			}()

			server := grpc.NewServer()
			devicepluginv1beta1.RegisterDevicePluginServer(server, pciHostDevicePlugin)
			pluginregistrationv1.RegisterRegistrationServer(server, pciHostDevicePlugin)

			server.Serve(listener)
		} else {
			logrus.Error(err)
		}
	}, cancel)

	return pciHostDevicePlugin
}

func (plugin pciHostDevicePlugin) Allocate(ctx context.Context, request *devicepluginv1beta1.AllocateRequest) (*devicepluginv1beta1.AllocateResponse, error) {
	response := &devicepluginv1beta1.AllocateResponse{}

	for _, modulename := range []string{
		"vfio_iommu_type1",
		"vfio_pci",
	} {
		if err := kmodule.Probe(modulename, ""); err != nil {
			return nil, err
		}
	}
	envKey := util.ResourceNameToEnvVar(kubevirtv1.PCIResourcePrefix, plugin.pciHostDevice.ResourceName)
	for _, containerRequest := range request.ContainerRequests {
		var envValue string
		var devices []*devicepluginv1beta1.DeviceSpec
		for _, devicesID := range containerRequest.DevicesIDs {
			for _, pciDevice := range lspci.ListAll() {
				if pciDevice.Slot == devicesID {
					if pciDevice.Driver != "vfio-pci" {
						if pciDevice.Driver != "" {
							if err := pci.Driver(pciDevice.Driver).Unbind(pciDevice.Slot); err != nil {
								return nil, err
							}
						}
						if err := pci.Device(pciDevice.Slot).DriverOverride("vfio-pci"); err != nil {
							return nil, err
						}
						if err := pci.Driver("vfio-pci").Bind(pciDevice.Slot); err != nil {
							return nil, err
						}
					}
					if pciDevice.IOMMUGroup != "" {
						devices = append(devices, &devicepluginv1beta1.DeviceSpec{
							ContainerPath: fmt.Sprintf("/dev/vfio/%s", pciDevice.IOMMUGroup),
							HostPath:      fmt.Sprintf("/dev/vfio/%s", pciDevice.IOMMUGroup),
							Permissions:   "mrw",
						})
					}
					envValue = envValue + pciDevice.Slot
				}
			}
			envValue = envValue + ","
		}
		response.ContainerResponses = append(response.ContainerResponses, &devicepluginv1beta1.ContainerAllocateResponse{
			Envs: map[string]string{
				envKey: envValue,
			},
			Devices: append(devices, &devicepluginv1beta1.DeviceSpec{
				ContainerPath: "/dev/vfio/vfio",
				HostPath:      "/dev/vfio/vfio",
				Permissions:   "mrw",
			}),
		})
	}

	return response, nil
}

func (plugin pciHostDevicePlugin) GetDevicePluginOptions(ctx context.Context, _ *devicepluginv1beta1.Empty) (*devicepluginv1beta1.DevicePluginOptions, error) {
	options := &devicepluginv1beta1.DevicePluginOptions{}

	return options, nil
}

func (plugin pciHostDevicePlugin) GetInfo(ctx context.Context, request *pluginregistrationv1.InfoRequest) (*pluginregistrationv1.PluginInfo, error) {
	response := &pluginregistrationv1.PluginInfo{
		Type: pluginregistrationv1.DevicePlugin,
		Name: plugin.pciHostDevice.ResourceName,
		SupportedVersions: []string{
			"v1beta1",
		},
	}

	return response, nil
}

func (plugin pciHostDevicePlugin) GetPreferredAllocation(ctx context.Context, request *devicepluginv1beta1.PreferredAllocationRequest) (*devicepluginv1beta1.PreferredAllocationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (plugin pciHostDevicePlugin) ListAndWatch(_ *devicepluginv1beta1.Empty, server devicepluginv1beta1.DevicePlugin_ListAndWatchServer) error {
	response := &devicepluginv1beta1.ListAndWatchResponse{}

	for _, pciDevice := range lspci.ListAll() {
		if pciDevice.Vendor+":"+pciDevice.Device == plugin.pciHostDevice.PCIVendorSelector {
			response.Devices = append(response.Devices, &devicepluginv1beta1.Device{
				ID:     pciDevice.Slot,
				Health: devicepluginv1beta1.Healthy,
			})
		}
	}

	if err := server.Send(response); err != nil {
		return err
	}

	<-plugin.ctx.Done()

	return nil
}

func (plugin pciHostDevicePlugin) NotifyRegistrationStatus(ctx context.Context, request *pluginregistrationv1.RegistrationStatus) (*pluginregistrationv1.RegistrationStatusResponse, error) {
	response := &pluginregistrationv1.RegistrationStatusResponse{}

	if !request.PluginRegistered {
		logrus.Error(request.Error)
	}

	return response, nil
}

func (plugin pciHostDevicePlugin) PreStartContainer(ctx context.Context, request *devicepluginv1beta1.PreStartContainerRequest) (*devicepluginv1beta1.PreStartContainerResponse, error) {
	response := &devicepluginv1beta1.PreStartContainerResponse{}

	return response, nil
}
