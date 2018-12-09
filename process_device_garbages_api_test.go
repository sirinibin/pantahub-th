package main

import (
	"encoding/json"
	"log"
	"strconv"
	"testing"

	"github.com/go-resty/resty"
)

// TestMain4 : Process Device Garbages
func TestMain4(t *testing.T) {
	log.Print("Test:Process Device Garbages")
	// Create N-no.of devices(based on the value of "var DeviceCount"
	setUp4(t)

	MarkAllDevicesAsGarbage(t)

	// Process device Garbages : PUT processgarbages/devices
	status,
		deviceProcessed,
		trailsMarkedAsGarbage,
		trailsWithErrors := ProcessDeviceGarbages(t)
	log.Print(" Case 1:Processing garbage devices with no trails")
	if status == 0 &&
		deviceProcessed == 0 &&
		trailsMarkedAsGarbage == 0 &&
		trailsWithErrors == DeviceCount {
		log.Print(" Case 1:Passed")

	} else {
		t.Errorf(" Case 1,Error:Devices should be processed is:0, But actually processed is:" + strconv.Itoa(deviceProcessed))
		t.Errorf(" Case 1,Error:Trails should be marked as garbage is:0, But actually marked is:" + strconv.Itoa(trailsMarkedAsGarbage))
		t.Errorf(" Case 1,Error:Trails with errors should be " + strconv.Itoa(DeviceCount) + ", But actual errors are:" + strconv.Itoa(trailsWithErrors))
		t.Fail()
	}

	//Create trails for all devices
	CreateAllTrails(t)

	log.Print(" Case 2:Processing garbage devices with trails")
	// Process device Garbages : PUT processgarbages/devices
	status,
		deviceProcessed,
		trailsMarkedAsGarbage,
		trailsWithErrors = ProcessDeviceGarbages(t)

	if status == 1 &&
		deviceProcessed == DeviceCount &&
		trailsMarkedAsGarbage == DeviceCount &&
		trailsWithErrors == 0 {
		log.Print(" Case 2:Passed\n\n")
	} else {
		t.Errorf(" Case 2,Error:Devices should be processed is:" + strconv.Itoa(DeviceCount) + ", But actually processed is:" + strconv.Itoa(deviceProcessed))
		t.Errorf(" Case 2,Error:Trails should be marked as garbage is:" + strconv.Itoa(DeviceCount) + ", But actually marked is:" + strconv.Itoa(trailsMarkedAsGarbage))
		t.Errorf(" Case 2,Error:Trails with errors should be 0, But actually errors are:" + strconv.Itoa(trailsWithErrors))

		t.Fail()
	}

	tearDown4(t)
}
func setUp4(t *testing.T) bool {
	//db.Connect()
	ClearOldData(t)
	//1.Login with user/user & Obtain Access token
	login(t)

	//2.Create all devices with UTOKEN, API call: POST /devices
	CreateAllDevices(t)

	return true
}
func tearDown4(t *testing.T) bool {

	//Delete all devices
	if !DeleteAllDevices(t) {
		return false
	}
	//Delete all trails
	if !DeleteAllTrails(t) {
		return false
	}

	return true

}

// ProcessDeviceGarbages : Process Device Garbages
func ProcessDeviceGarbages(t *testing.T) (int, int, int, int) {
	response := map[string]interface{}{}
	APIEndPoint := GCAPIUrl + "/processgarbages/devices"
	res, err := resty.R().Put(APIEndPoint)
	if err != nil {
		t.Errorf("internal error calling test server: " + err.Error())
		t.Fail()
	}
	err = json.Unmarshal(res.Body(), &response)

	//log.Print(response)

	status := int(response["status"].(float64))
	deviceProcessed := int(response["device_processed"].(float64))
	trailsMarkedAsGarbage := int(response["trails_marked_as_garbage"].(float64))
	trailsWithErrors := int(response["trails_with_errors"].(float64))

	return status,
		deviceProcessed,
		trailsMarkedAsGarbage,
		trailsWithErrors

}

func ClearOldData(t *testing.T) bool {
	DeleteAllDevices(t)
	DeleteAllTrails(t)
	return true
}
