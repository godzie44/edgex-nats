package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const (
	topic       = "Topic"
	clusterId   = "ClusterId"
	addr        = "Addr"
	deviceNames = "DeviceNames"
)

var log logger.LoggingClient

// Config holds AWS IoT specific information
type Config struct {
	StanTopic     string
	StanClusterId string
	StanAddr      string
	DeviceNames   string
}

func getNewClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	return &http.Client{Timeout: 10 * time.Second, Transport: tr}
}

func getAppSetting(settings map[string]string, name string) string {
	value, ok := settings[name]

	if ok {
		log.Debug(value)
		return value
	}
	log.Error(fmt.Sprintf("ApplicationName application setting %s not found", name))
	return ""
}

// LoadConfig Loads the mqtt configuration necessary to connect to AWS cloud
func LoadConfig(sdk *appsdk.AppFunctionsSDK) (*Config, error) {
	if sdk == nil {
		return nil, errors.New("Invalid AppFunctionsSDK")
	}

	log = sdk.LoggingClient

	appSettings := sdk.ApplicationSettings()
	if appSettings == nil {
		return nil, errors.New("No application-specific settings found")
	}
	config := Config{}

	config.StanAddr = getAppSetting(appSettings, addr)
	config.StanTopic = getAppSetting(appSettings, topic)
	config.StanClusterId = getAppSetting(appSettings, clusterId)
	config.DeviceNames = getAppSetting(appSettings, deviceNames)

	return &config, nil
}
