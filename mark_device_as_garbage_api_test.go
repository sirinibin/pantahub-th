package main

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/go-resty/resty"
	"gitlab.com/pantacor/pantahub-gc/db"
)

// TestMain3 : Mark Devices as Garbages
func TestMain3(t *testing.T) {
	setUp3(t)
	log.Print("Test:Mark Devices as Garbages")
	// PUT markgarbage/device/<DEVICE_ID>
	log.Print(" Case 1:Mark all devices as Garbages")
	if MarkAllDevicesAsGarbage(t) {
		log.Print(" Case 1:Passed\n\n")
	}
	tearDown3(t)
}

func setUp3(t *testing.T) bool {
	db.Connect()
	ClearOldData(t)
	//1.Login with user/user & Obtain Access token
	login(t)
	//2.Create all devices with UTOKEN, API call: POST /devices
	CreateDevices(t)

	return true
}
func tearDown3(t *testing.T) bool {

	//Delete all trails
	if !DeleteAllDevices(t) {
		return false
	}

	return true

}

func MarkAllDevicesAsGarbage(t *testing.T) bool {

	for _, device := range Devices {
		if !MarkDeviceAsGarbage(t, device.ID.Hex()) {
			return false
		}
	}
	return true
}

// MarkDeviceAsGarbage : Mark Device as Garbage
func MarkDeviceAsGarbage(t *testing.T, deviceID string) bool {
	response := map[string]interface{}{}
	APIEndPoint := GCAPIUrl + "/markgarbage/device/" + deviceID
	res, err := resty.R().Put(APIEndPoint)
	if err != nil {
		t.Errorf("internal error calling test server: " + err.Error())
		t.Fail()
	}
	err = json.Unmarshal(res.Body(), &response)
	if res.StatusCode() != 200 {
		log.Print(response)
		t.Fail()
	}
	status := int(response["status"].(float64))
	id := response["device"].(map[string]interface{})["id"]
	// check if status==1 and device id=deviceID
	if status != 1 || id != deviceID {
		t.Errorf("Error:Expected device id:" + deviceID + ",but got:" + id.(string))
		t.Fail()
		return false
	}
	return true

}
