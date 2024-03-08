package internal

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
	"github.com/inaccel/daemon/pkg/plugin"
	"github.com/inaccel/device-selector/pkg/lspci"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kubevirt.io/kubevirt/pkg/hooks/info"
	kubevirthooksv1alpha2 "kubevirt.io/kubevirt/pkg/hooks/v1alpha2"
)

type hook struct {
	ctx  context.Context
	path string

	plugin.Plugin
}

func NewHook(ctx context.Context) plugin.New {
	return func() plugin.Plugin {
		return newHook(ctx)
	}
}

func newHook(ctx context.Context) plugin.Plugin {
	ctx, cancel := context.WithCancel(ctx)

	hook := &hook{
		ctx:  ctx,
		path: filepath.Join("/var/run/kubevirt-hooks/inaccel.sock"),
	}

	hook.Plugin = plugin.Base(func() {
		if listener, err := listen(hook.path); err == nil {
			go func() {
				<-ctx.Done()

				listener.Close()
			}()

			server := grpc.NewServer()
			kubevirthooksv1alpha2.RegisterCallbacksServer(server, hook)
			info.RegisterInfoServer(server, hook)

			server.Serve(listener)
		} else {
			logrus.Error(err)
		}
	}, cancel)

	return hook
}

func (hook hook) Info(ctx context.Context, params *info.InfoParams) (*info.InfoResult, error) {
	result := &info.InfoResult{
		Name: "inaccel",
		HookPoints: []*info.HookPoint{
			{
				Name: info.OnDefineDomainHookPointName,
			},
		},
		Versions: []string{
			"v1alpha2",
		},
	}

	return result, nil
}

func (hook hook) OnDefineDomain(ctx context.Context, params *kubevirthooksv1alpha2.OnDefineDomainParams) (*kubevirthooksv1alpha2.OnDefineDomainResult, error) {
	result := &kubevirthooksv1alpha2.OnDefineDomainResult{}

	xml := etree.NewDocument()
	if err := xml.ReadFromBytes(params.DomainXML); err != nil {
		return nil, err
	}
	for index, hostdev := range xml.FindElements("domain/devices/hostdev") {
		if hostdev.SelectAttrValue("type", "") == "pci" && hostdev.FindElement("address") == nil {
			domain := strings.TrimPrefix(hostdev.FindElement("source/address").SelectAttrValue("domain", ""), "0x")
			bus := strings.TrimPrefix(hostdev.FindElement("source/address").SelectAttrValue("bus", ""), "0x")
			slot := strings.TrimPrefix(hostdev.FindElement("source/address").SelectAttrValue("slot", ""), "0x")

			for _, pciDevice := range lspci.ListAll() {
				if strings.HasPrefix(pciDevice.Slot, fmt.Sprintf("%s:%s:%s.", domain, bus, slot)) {
					function := strings.TrimPrefix(pciDevice.Slot, fmt.Sprintf("%s:%s:%s.", domain, bus, slot))

					hostdevCopy := hostdev.Copy()

					address := hostdevCopy.CreateElement("address")
					address.CreateAttr("type", "pci")
					address.CreateAttr("domain", "0x0000")
					address.CreateAttr("bus", fmt.Sprintf("0x%02x", index+1))
					address.CreateAttr("slot", "0x00")
					address.CreateAttr("function", "0x"+function)
					if function == "0" {
						address.CreateAttr("multifunction", "on")
					}

					hostdevCopy.FindElement("alias").CreateAttr("name", hostdevCopy.FindElement("alias").SelectAttrValue("name", "")+"-"+function)

					hostdevCopy.FindElement("source/address").CreateAttr("function", "0x"+function)

					xml.FindElement("domain/devices").InsertChildAt(hostdev.Index(), hostdevCopy)
				}
			}
			xml.FindElement("domain/devices").RemoveChild(hostdev)
		}
	}
	xml.IndentTabs()
	domainXML, err := xml.WriteToBytes()
	if err != nil {
		return nil, err
	}
	result.DomainXML = domainXML

	return result, nil
}

func (hook hook) PreCloudInitIso(ctx context.Context, params *kubevirthooksv1alpha2.PreCloudInitIsoParams) (*kubevirthooksv1alpha2.PreCloudInitIsoResult, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
