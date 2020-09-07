package main

import (
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"log"
	natstransforms "nats-export/internal/config"
	"os"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"

	"github.com/nats-io/stan.go"
)

const (
	serviceKey = "NATSexport"
)

func main() {
	os.Setenv("EDGEX_SECURITY_SECRET_STORE", "false")

	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	config, err := natstransforms.LoadConfig(edgexSdk)
	if err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("Failed to load AWS MQTT configurations: %v\n", err))
		os.Exit(-1)
	}

	stanConn, err := stan.Connect(
		config.StanClusterId, "edgex_client2", stan.Pings(60, 2*60),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatal("Connection lost, reason: " + reason.Error())
		}),
		stan.NatsURL(config.StanAddr),
	)
	if err != nil {
		edgexSdk.LoggingClient.Error("on stan connect ", err.Error())
		os.Exit(-1)
	}
	defer stanConn.Close()

	deviceNamesCleaned := util.DeleteEmptyAndTrim(strings.FieldsFunc(config.DeviceNames, util.SplitComma))
	edgexSdk.LoggingClient.Debug(fmt.Sprintf("Device names read %s\n", deviceNamesCleaned))

	err = edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(deviceNamesCleaned).FilterByDeviceName,
		makeStanSender(stanConn, config.StanTopic),
	)
	if err != nil {
		edgexSdk.LoggingClient.Error("on start pipeline ", err.Error())
		os.Exit(-1)
	}

	if err = edgexSdk.MakeItRun(); err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}
}

type stanMessage struct {
	Device    string      `json:"device"`
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Timestamp int64       `json:"timestamp"`
}

func makeStanSender(conn stan.Conn, subj string) func(*appcontext.Context, ...interface{}) (bool, interface{}) {
	return func(edgexContext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		if len(params) < 1 {
			return false, nil
		}

		if _, ok := params[0].(models.Event); !ok {
			return false, nil
		}

		for _, r := range params[0].(models.Event).Readings {
			msg := stanMessage{
				Device:    r.Device,
				Name:      r.Name,
				Value:     r.Value,
				Timestamp: r.Origin,
			}

			marshaledMsg, _ := json.Marshal(msg)
			_, err := conn.PublishAsync(subj, marshaledMsg, func(_ string, err error) {
				if err != nil {
					edgexContext.LoggingClient.Error("stan publish error:", err.Error())
				}
			})
			if err != nil {
				edgexContext.LoggingClient.Error("stan publish error:", err.Error())
			}
		}

		edgexContext.Complete([]byte{})
		return false, nil
	}
}
