package router

import "net"

type ConfigParams interface {
	WlanSsid() *string
	WlanPassword() *string
	ConfigApi() string
}

type Configurator interface {
	Model() string
	Detected() bool
	WanMacAddress() (net.HardwareAddr, error)
	Configure(p ConfigParams) error
}
