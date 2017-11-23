package ws

import (
	"golang.org/x/net/websocket"
	"fmt"
	"encoding/json"
	"github.com/edmocosta/router-config-client/app/router/detector"
	"github.com/edmocosta/router-config-client/app/router"
	"github.com/edmocosta/router-config-client/app/customer"
	"os"
)

const (
	SubjectCheckClient         = "check_client"
	SubjectCheckClientResponse = "check_client_response"
	SubjectConfigure           = "configure"
	SubjectConfigureResponse   = "configure_response"
	SubjectRouterInfo          = "router_info"
	SubjectRouterInfoResponse  = "router_info_response"
	SubjectClose               = "close"
)

func ConfigHandler(ws *websocket.Conn) {
	for {
		var msg []byte
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			fmt.Println("Error while receiving a message from websocket connection!")
			break
		}

		jsonMsg := Message{}
		if err := json.Unmarshal(msg, &jsonMsg); err != nil {
			continue
		}

		switch jsonMsg.Subject {
		case SubjectCheckClient:
			msg, _ := json.Marshal(&Message{Subject: SubjectCheckClientResponse})
			websocket.Message.Send(ws, string(msg))
		case SubjectRouterInfo:
			routerInfoMsg := &RouterInfoRequest{}
			if err := json.Unmarshal(msg, routerInfoMsg); err != nil {
				continue
			}
			getRouterInfo(ws, routerInfoMsg)
		case SubjectConfigure:
			routerConfigureMsg := &RouterConfigureRequest{}
			if err := json.Unmarshal(msg, routerConfigureMsg); err != nil {
				continue
			}
			configureRouter(ws, routerConfigureMsg)
		case SubjectClose:
			os.Exit(0)
		}
	}
}

func detectRouterConfigurator(ws *websocket.Conn, subject string) router.Configurator {
	configurator := detector.Detect()
	if configurator == nil {
		msg, _ := json.Marshal(&Message{Subject: subject, Error: true, Message: "Router not detected. " +
			"Please check if your router is connected to the same network that your computer."})
		websocket.Message.Send(ws, string(msg))
		return nil
	}
	return configurator
}

func getRouterInfo(ws *websocket.Conn, message *RouterInfoRequest) {
	configurator := detectRouterConfigurator(ws, SubjectRouterInfoResponse)
	if configurator == nil {
		return
	}

	mac, err := configurator.WanMacAddress()
	if err != nil {
		msg, _ := json.Marshal(&Message{Subject: SubjectRouterInfoResponse, Error: true, Message: err.Error()})
		websocket.Message.Send(ws, string(msg))
		return
	}

	config, err := customer.FindConfigByMac(message.Api, mac.String())
	if err != nil {
		msg, _ := json.Marshal(&Message{Subject: SubjectRouterInfoResponse, Error: true, Message: err.Error()})
		websocket.Message.Send(ws, string(msg))
		return
	}

	response := &RouterInfoResponse{Message: &Message{Subject: SubjectRouterInfoResponse},
		Mac: mac.String(),
		Model: configurator.Model(),
		Customer: config.Nome}

	msg, _ := json.Marshal(response)
	websocket.Message.Send(ws, string(msg))
}

func configureRouter(ws *websocket.Conn, message *RouterConfigureRequest) {
	configurator := detectRouterConfigurator(ws, SubjectConfigureResponse)
	if configurator == nil {
		return
	}

	err := configurator.Configure(message)
	if err != nil {
		msg, _ := json.Marshal(&Message{Subject: SubjectConfigureResponse, Error: true, Message: err.Error()})
		websocket.Message.Send(ws, string(msg))
	}

	msg, _ := json.Marshal(&Message{Subject: SubjectConfigureResponse, Message: "Router configured! " +
		"You router will be rebooted, wait a few minutes and try to connect to the internet."})
	websocket.Message.Send(ws, string(msg))
}
