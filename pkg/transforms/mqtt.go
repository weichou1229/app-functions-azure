package transforms

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	sdkTransforms "github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	serviceKey         = "AzureExport"
	appConfigIoTHub    = "IoTHub"
	appConfigIoTDevice = "IoTDevice"
	appConfigMQTTCert  = "MQTTCert"
	appConfigMQTTKey   = "MQTTKey"
	appConfigTokenPath = "TokenPath"
	appConfigVaultHost = "VaultHost"
	appConfigVaultPort = "VaultPort"
	appConfigCertPath  = "CertPath"
	mqttPort           = 8883
	vaultToken         = "X-Vault-Token"
)

// global logger
var log logger.LoggingClient

type certCollect struct {
	Pair certPair `json:"data"`
}

type certPair struct {
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

type auth struct {
	Token string `json:"root_token"`
}

func getAppSetting(settings map[string]string, name string) string {
	value, ok := settings[name]

	if ok {
		return value
	} else {
		log.Error(fmt.Sprintf("ApplicationName application setting %s not found", name))
		return ""
	}
}

func retrieveKeyPair(tokenPath string, vaultHost string, vaultPort string, certPath string) (*sdkTransforms.KeyCertPair, error) {
	a := auth{}
	content, err := ioutil.ReadFile(tokenPath)

	if err == nil {
		err = json.Unmarshal(content, &a)
		if err == nil {
			// we hae a.Token here
			s := sling.New().Set(vaultToken, a.Token)
			vaultUrl := fmt.Sprintf("https://%s:%s/", vaultHost, vaultPort)
			req, err := s.New().Base(vaultUrl).Get(certPath).Request()

			if err == nil {
				res, err := getNewClient(true).Do(req)

				if err == nil {
					defer res.Body.Close()

					cc := certCollect{}
					json.NewDecoder(res.Body).Decode(&cc)

					pair := &sdkTransforms.KeyCertPair{
						KeyPEMBlock:  []byte(cc.Pair.Key),
						CertPEMBlock: []byte(cc.Pair.Cert),
					}

					log.Info("Successfully loaded key/cert pair from Vault")

					return pair, nil
				} else {
					log.Error("Client request failed", err.Error())
					return nil, err
				}
			} else {
				log.Error("Failed to create request", err.Error())
				return nil, err
			}
		}
	} else {
		log.Error("Failed to read token file", err.Error())
		return nil, err
	}

	// won't reach here
	return nil, nil
}

func getNewClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	return &http.Client{Timeout: 10 * time.Second, Transport: tr}
}

func NewMQTTSender(sdk *appsdk.AppFunctionsSDK) *sdkTransforms.MQTTSender {
	log = sdk.LoggingClient

	var iotHub, iotDevice, mqttCert, mqttKey, tokenPath, vaultHost, vaultPort, certPath string

	appSettings := sdk.ApplicationSettings()

	if appSettings != nil {
		iotHub = getAppSetting(appSettings, appConfigIoTHub)
		iotDevice = getAppSetting(appSettings, appConfigIoTDevice)
		mqttCert = getAppSetting(appSettings, appConfigMQTTCert)
		mqttKey = getAppSetting(appSettings, appConfigMQTTKey)
		tokenPath = getAppSetting(appSettings, appConfigTokenPath)
		vaultHost = getAppSetting(appSettings, appConfigVaultHost)
		vaultPort = getAppSetting(appSettings, appConfigVaultPort)
		certPath = getAppSetting(appSettings, appConfigCertPath)
	} else {
		log.Error("No application-specific settings found")
		return nil
	}

	// Generate Azure-specific host, user amd topic
	host := fmt.Sprintf("%s.azure-devices.net", iotHub)
	user := fmt.Sprintf("%s/%s/?api-version=2018-06-30", host, iotDevice)
	topic := fmt.Sprintf("devices/%s/messages/events/", iotDevice)

	addressable := models.Addressable{
		Address:   host,
		Port:      mqttPort,
		Protocol:  "tls",
		Publisher: iotDevice, // must be the same as the device name
		User:      user,
		Password:  "",
		Topic:     topic,
	}

	// Retrieve key/cert pair from Vault
	pair, err := retrieveKeyPair(tokenPath, vaultHost, vaultPort, certPath)

	// Fall back to local key/cert files
	if err != nil {
		pair = &sdkTransforms.KeyCertPair{
			KeyFile:  mqttKey,
			CertFile: mqttCert,
		}
	}

	mqttConfig := sdkTransforms.NewMqttConfig()
	mqttSender := sdkTransforms.NewMQTTSender(log, addressable, pair, mqttConfig)

	return mqttSender
}
