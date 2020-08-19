// Copyright (c) Microsoft. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// This application uses the Azure IoT Hub service SDK for Java
// For samples see: https://github.com/Azure/azure-iot-sdk-java/tree/master/service/iot-service-samples

package com.microsoft.docs.iothub.samples;

import com.microsoft.azure.sdk.iot.service.devicetwin.DeviceMethod;
import com.microsoft.azure.sdk.iot.service.devicetwin.MethodResult;
import com.microsoft.azure.sdk.iot.service.exceptions.IotHubException;

import java.io.IOException;
import java.nio.charset.Charset;
import java.util.concurrent.TimeUnit;

public class BackEndApplication {

  // Connection string for your IoT Hub
  // az iot hub show-connection-string --hub-name {your iot hub name} --policy-name service
  public static final String iotHubConnectionString = "HostName=EssaiIOT.azure-devices.net;SharedAccessKeyName=service;SharedAccessKey=pto2o31P7kuJwEEPNdcUffm4ZxUpfEF9iNPSx5wYXCI=";

  // Device to call direct method on.
  public static final String deviceId = "Coriance_Pub_Device";

  private static final String TURN_ON  = "turn-on";
  private static final String TURN_OFF = "turn-off";

  // Name of direct method and payload.
  private static String methodName = "SetTelemetryInterval";
  public static final String payload = "10"; // Number of seconds for telemetry interval.

  public static final Long responseTimeout = TimeUnit.SECONDS.toSeconds(30);
  public static final Long connectTimeout = TimeUnit.SECONDS.toSeconds(5);

  public static void main(String[] args) {
    try {
      // Create a DeviceMethod instance to call a direct method.
      DeviceMethod methodClient = DeviceMethod.createFromConnectionString(iotHubConnectionString);

      MethodResult result = null;

//      String deviceId = "";
      String payload = "10";

      if (args.length >= 2) {
//        deviceId = args[0];
        methodName = args[0];
        payload = args[1];
        System.out.println("payload: " + args[1]);
      }else{
        System.out.println("Not enough arguments(methodName, payload)");
        System.out.println("Ex: java -jar invoke-direct-method.jar Configuration '{ \"DemandWindowSize\" : \"25\", \"LineFrequency\" : \"70\" }'\n");
        return;
      }

      // Call the direct method.
      result = methodClient.invoke(deviceId, methodName, responseTimeout, connectTimeout, payload);

      System.out.println("Calling direct method " + methodName);

      if (result == null) {
        throw new IOException("Direct method invoke returns null");
      }

      // Show the acknowledgement from the device.
      System.out.println("Status: " + result.getStatus());
      System.out.println("Response: " + result.getPayload());
    } catch (IotHubException e) {
      System.out.println("IotHubException calling direct method:");
      System.out.println(e.getMessage());
    } catch (IOException e) {
      System.out.println("IOException calling direct method:");
      System.out.println(e.getMessage());
    }
    System.out.println("Done!");
  }
}
