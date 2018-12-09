package main

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/go-resty/resty"
	"gitlab.com/pantacor/pantahub-gc/db"
)

// TestMain2 : Mark Trails as Garbages that lost their parent devices
func TestMain3(t *testing.T) {
	log.Print("Inside Test Main3")
	// Create N-no.of devices(based on the value of "var DeviceCount"
	setUp3(t)
	// PUT markgarbage/device/<DEVICE_ID>
	MarkAllDevicesAsGarbage(t)
	tearDown3(t)
}

func setUp3(t *testing.T) bool {
	db.Connect()
	//1.Login with user/user & Obtain Access token
	login(t)
	//2.Create all devices with UTOKEN, API call: POST /devices
	CreateAllDevices(t)

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

	//log.Print(response["device"])
	id := response["device"].(map[string]interface{})["id"]
	//log.Print("Devce ID:")
	//log.Print(id)

	// check if status==1 and device id=deviceID
	if status != 1 || id != deviceID {
		t.Errorf("Error on marking device as garbage")
		t.Fail()
		return false
	}
	return true

}
