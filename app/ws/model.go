package ws

type Message struct {
	Subject string `json:"subject"`
	Message string `json:"message,omitempty"`
	Error   bool `json:"error,omitempty"`
}

type RouterConfigureRequest struct {
	*Message
	Api      string `json:"api"`
	Ssid     *string `json:"wlan_ssid"`
	Password *string `json:"wlan_password"`
}

func (r *RouterConfigureRequest) WlanSsid() *string {
	return r.Ssid
}

func (r *RouterConfigureRequest) WlanPassword() *string {
	return r.Password
}

func (r *RouterConfigureRequest) ConfigApi() string {
	return r.Api
}

type RouterInfoRequest struct {
	*Message
	Api string `json:"api"`
}

type RouterInfoResponse struct {
	*Message
	Model    string `json:"model"`
	Mac      string `json:"mac"`
	Customer string `json:"customer"`
}
