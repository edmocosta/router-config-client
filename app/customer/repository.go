package customer

import (
	"net/http"
	"fmt"
	"errors"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"os"
	"io"
)

var (
	ErrServerApi = errors.New("An error was occurred while trying to get your account data. Try again in a few minutes")
)

type Config struct {
	Nome     string  `json:"nome"`
	UserName *string `json:"username"`
	Password *string `json:"password"`
}

type GetConfigFileReq struct {
	Mac           string `json:"mac"`
	WlanSsid      string `json:"wlan_ssid"`
	WlanPassword  string `json:"wlan_password"`
	RouterModel   string `json:"router_model"`
	RouterVersion string `json:"router_version"`
}

func GetConfigFile(api string, p GetConfigFileReq) (*os.File, error) {
	v, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/config", api), bytes.NewBuffer(v))
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return nil, ErrServerApi
	}

	cfgFile, err := ioutil.TempFile("", "router_config")
	if err != nil {
		return nil, err
	}

	defer cfgFile.Close()

	if _, err = io.Copy(cfgFile, resp.Body); err != nil {
		return nil, err
	}

	return cfgFile, nil
}

func FindConfigByMac(api string, mac string) (*Config, error) {
	resp, err := http.DefaultClient.Get(fmt.Sprintf("%s/config?mac=%s", api, mac))
	if err != nil {
		return nil, ErrServerApi
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, errors.New("Unable to find your regitry/hardware configuration.")
	}

	cfg := &Config{}
	json.NewDecoder(resp.Body).Decode(cfg)
	return cfg, nil
}
