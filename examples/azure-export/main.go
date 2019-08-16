package main

import (
	"fmt"
	"os"

	azureTransforms "github.com/IOTechSystems/app-functions-azure/pkg/transforms"
	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
)

const (
	serviceKey = "AzureExport"
)

var counter int

func main() {
	// 1) First thing to do is to create an instance of the EdgeX SDK and initialize it.
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	// 2) Since our DeviceNameFilter Function requires the list of device names we would
	// like to search for, we'll go ahead and define that now.
	deviceName := []string{"Random-Integer-Generator01"}

	// 3) This is our pipeline configuration, the collection of functions to
	// execute every time an event is triggered.
	edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(deviceName).FilterByDeviceName,
		azureTransforms.NewConversion().TransformToAzure,
		azureTransforms.NewMQTTSender(edgexSdk).MQTTSend,
	)

	// 5) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	err := edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}
