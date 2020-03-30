package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
)


type StaticRouteConfiguration struct {
	Destination	string	`json:"destination"`
	Nexthop		string	`json:"nexthop"`
}

type Configuration struct {
	IPAddress string 							`json:"ip_address" default:"10.0.0.1/24"`
	IfaceName string 							`json:"interface_name" default:"spig0"`
	PrivateSeedFile	string						`json:"private_seed_file" default:"/etc/spigot/secrets/seed"`
	AuthorizedKeys	[]string					`json:"authorized_keys"`
	StaticRoutes	[]StaticRouteConfiguration	`json:"static_routes"`
}

func getDefaultConfiguration() (*Configuration) {
	cfg := &Configuration{}

	t := reflect.TypeOf(cfg).Elem()
	v := reflect.ValueOf(cfg).Elem()
	for i := 0 ; i < t.NumField() ; i++ {
		tF := t.Field(i)
		def, ok := tF.Tag.Lookup("default")
		if ok {
			v.Field(i).Set(reflect.ValueOf(def))
		}
	}

	return cfg
}

func GetConfiguration() (*Configuration, error) {
	cfg := getDefaultConfiguration()

	fd, err := os.Open("/etc/spigot/config.json")
	if err == nil {
		b, err := ioutil.ReadAll(fd)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(b, cfg)
	}

	return cfg, nil
}