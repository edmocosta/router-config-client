package intelbras

import (
	"net/http"
	"encoding/json"
	"github.com/edmocosta/router-config-client/app/router"
	"fmt"
	"net"
	"time"
	"errors"
	"io"
	"bytes"
	"github.com/edmocosta/router-config-client/app/customer"
	"mime/multipart"
	"os"
	"strings"
)

var (
	hosts          = []string{"http://10.0.0.1", "http://meuintelbras.local"}
	adminPasswords = []string{"admin", "123456", ""}
)

type Configurator struct {
	host          string
	model         string
	wanMacAddress net.HardwareAddr
	detected      bool
}

func NewConfigurator() router.Configurator {
	detected, host := checkRouter()
	if detected {
		doAdminLogin(host)
	}

	return &Configurator{host: host, detected: detected}
}

func (i *Configurator) Detected() bool {
	return i.detected
}

func (i *Configurator) Model() string {
	if i.model == "" {
		i.model = getRouterModel(i.host)
	}
	return string(i.model)
}

func (i *Configurator) WanMacAddress() (net.HardwareAddr, error) {
	if i.wanMacAddress == nil {
		var err error
		i.wanMacAddress, err = getRouterMac(i.host)
		if err != nil {
			return nil, err
		}
	}
	return i.wanMacAddress, nil
}

func (i *Configurator) Configure(p router.ConfigParams) error {

	fmt.Println(fmt.Sprintf("Configuring router %s...", i.Model()))

	if err := doAdminLogin(i.host); err != nil {
		return err
	}

	mac, _ := i.WanMacAddress()
	version := getRouterVersion(i.host)

	cfgFile, _ := customer.GetConfigFile(p.ConfigApi(), customer.GetConfigFileReq{
		Mac:           mac.String(),
		WlanSsid:      *p.WlanSsid(),
		WlanPassword:  *p.WlanPassword(),
		RouterModel:   i.Model(),
		RouterVersion: version})

	if err := uploadConfigFile(i.host, cfgFile.Name()); err != nil {
		return err
	}

	return nil
}

func uploadConfigFile(apiConfig string, filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	defer w.Close()

	fw, err := w.CreateFormFile("file", "config.img")
	if err != nil {
		return err
	}

	if _, err = io.Copy(fw, file); err != nil {
		return err
	}

	w.Close()

	req, err := newRequest("POST", apiConfig, "/form2fileConf.htm", &b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 2 * time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		fmt.Println("Router configured!")
		return nil
	}

	return fmt.Errorf("Error while uploading the configuration file to router! Code: %s", res.Status)
}

func execRequest(verb string, host string, path string, body io.Reader) (map[string]interface{}, error) {
	req, err := newRequest(verb, host, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var jsonResponse = map[string]interface{}{}
	json.NewDecoder(resp.Body).Decode(&jsonResponse)

	return jsonResponse, nil
}

func newRequest(verb string, host string, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(verb, host+path, body)
	if err != nil {
		return req, err
	}
	req.Header.Add("Referer", host)
	return req, nil
}

func doAdminLogin(host string) error {
	for _, password := range adminPasswords {
		jsonLogin := map[string]interface{}{}
		jsonLogin["username"] = "admin"
		jsonLogin["password"] = password

		body, err := json.Marshal(jsonLogin)
		loginReq, _ := newRequest("POST", host, "/v1/system/login", bytes.NewBuffer(body))
		loginReq.Header.Set("Content-Type", "application/json")

		loginResp, err := http.DefaultClient.Do(loginReq)
		if err != nil {
			continue
		}

		if loginResp.StatusCode == http.StatusOK {
			return nil
		}
	}
	return errors.New("Login failed")
}

func getRouterMac(host string) (net.HardwareAddr, error) {
	resp, _ := execRequest("GET", host, "/v1/interface/wan", nil)
	mac, ok := resp["mac"]
	if !ok {
		return nil, errors.New("Cannot get the hardware address (MAC) from your router")
	}

	return net.ParseMAC(mac.(string))
}

func getRouterModel(host string) string {
	jsonResponse, err := getDeviceInfo(host)
	if err != nil {
		return "Unknown"
	}
	val, ok := jsonResponse["model"]
	if !ok {
		return "Unknown"
	}
	return val.(string)
}

func getRouterVersion(host string) string {
	var jsonResponse, err = getDeviceInfo(host)
	if err != nil {
		return ""
	}

	val, ok := jsonResponse["version"]
	if !ok {
		return ""
	}

	return val.(string)
}

func getDeviceInfo(host string) (map[string]interface{}, error) {
	req, _ := newRequest("GET", host, "/v1/system/device", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var jsonResponse = map[string]interface{}{}
	json.NewDecoder(resp.Body).Decode(&jsonResponse)
	return jsonResponse, nil
}

func checkRouter() (bool, string) {
	for _, host := range hosts {
		device, err := getDeviceInfo(host)
		if err != nil || device == nil {
			continue
		}
		model, ok := device["model"]
		if !ok {
			continue
		}
		if strings.Contains(strings.ToUpper(model.(string)), "IWR-1000 N") {
			return true, host
		}
	}
	return false, ""
}
