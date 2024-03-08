package main

import (
	"io"
	"log"
	"os"

	"github.com/inaccel/daemon/pkg/plugin"
	"github.com/inaccel/device-selector/internal"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var version string

func main() {
	app := &cli.App{
		Name:    "device-selector",
		Version: version,
		Usage:   "A self-sufficient runtime for accelerators.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug output",
			},
		},
		Before: func(context *cli.Context) error {
			log.SetOutput(io.Discard)

			logrus.SetFormatter(new(logrus.JSONFormatter))

			if context.Bool("debug") {
				logrus.SetLevel(logrus.DebugLevel)
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			kube, err := config.GetConfig()
			if err != nil {
				return err
			}
			api, err := client.New(kube, client.Options{})
			if err != nil {
				return err
			}

			if err := kubevirtv1.AddToScheme(api.Scheme()); err != nil {
				return err
			}

			kubeVirt := &kubevirtv1.KubeVirt{}
			if err := api.Get(context.Context, client.ObjectKey{
				Namespace: os.Getenv("KUBE_VIRT_NAMESPACE"),
				Name:      os.Getenv("KUBE_VIRT_NAME"),
			}, kubeVirt); err != nil {
				return err
			}

			new := []plugin.New{
				internal.NewHook(context.Context),
			}
			if kubeVirt.Spec.Configuration.PermittedHostDevices != nil {
				for _, pciHostDevice := range kubeVirt.Spec.Configuration.PermittedHostDevices.PciHostDevices {
					if pciHostDevice.ExternalResourceProvider {
						new = append(new, internal.NewPciHostDevicePlugin(context.Context, pciHostDevice))
					}
				}
			}

			plugin.Handle(new...)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
