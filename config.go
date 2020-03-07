package main

import (
	"io"
	"os"

	"github.com/paked/configure"
)


type Configuration struct {
	IPAddress string
	IfaceName string
}

func GetConfiguration() (*Configuration, error) {
	configFileConfig := configure.New()
	configFile := configFileConfig.String("config-file", "/etc/spigot/config.json", "Configuration file to use")
	configFileConfig.Use(
		configure.NewFlag(),
	)
	configFileConfig.Parse()

	openConfigFile := func() (io.Reader, error) {
		return os.Open(*configFile)
	}

	conf := configure.New()

	ipAddress := conf.String("ip-address", "10.0.0.1/24", "IP address for the Spigot interface")
	ifaceName := conf.String("iface-name", "spig0", "Name for the Spigot interface")

	conf.Use(
		configure.NewFlag(),
		configure.NewJSON(openConfigFile),
	)

	conf.Parse()

	c := &Configuration{
		IPAddress: *ipAddress,
		IfaceName: *ifaceName,
	}

	return c, nil
}