package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/go-resty/resty"
	"gitlab.com/pantacor/pantahub-base/utils"
	"gitlab.com/pantacor/pantahub-gc/db"
	"gitlab.com/pantacor/pantahub-gc/models"
	"gopkg.in/mgo.v2/bson"
)

var GCAPIUrl = "http://localhost:2000"
var BaseAPIUrl = "http://localhost:12365"
var UTOKEN = ""
var DeviceCount = 3

func TestMain(t *testing.T) {
	// Create N-no.of devices(based on the value of "var DeviceCount"
	devices := setUp(t)
	// PUT /markgarbage/devices/unclaimed : which marks all unclaimed devices as garbage
	MarkAllUnClaimedDevicesAsGrabage(t, DeviceCount)
	// PUT /markgarbage/devices/unclaimed : to make sure that there is no devices left to mark
	MarkAllUnClaimedDevicesAsGrabage(t, 0)
	// Delete all created devices
	tearDown(t, devices)

}
func setUp(t *testing.T) []models.Device {
	db.Connect()
	response := map[string]interface{}{}
	//1.Login with user/user & Obtain Access token
	response = login(t)
	UTOKEN = response["token"].(string)

	var devices []models.Device
	for i := 0; i < DeviceCount; i++ {

		//2.Create a device with UTOKEN, API call: POST /devices
		device := createDevice(t)
		//3.Update device timecreated field to less than 1 min of PANTAHUB_GC_UNCLAIMED_EXPIRY
		device = UpdateDeviceTimeCreated(t, device)

		devices = append(devices, device)
		log.Print("Created device:" + device.ID.Hex())
	}

	return devices
}
func tearDown(t *testing.T, devices []models.Device) {
	//Delete all devices
	for _, device := range devices {
		if !DeleteDevice(t, device) {
			t.Errorf("Something went wrong while deleting device:" + device.ID.Hex())
			t.Fail()
		}
		log.Print("Deleted device:" + device.ID.Hex())
	}

}
func DeleteDevice(t *testing.T, device models.Device) bool {
	db := db.Session
	c := db.C("pantahub_devices")
	//log.Print("Device id:" + device.ID)
	err := c.Remove(bson.M{"_id": device.ID})
	if err != nil {
		t.Errorf("internal error calling test server: " + err.Error())
		t.Fail()
		return false
	}
	return true
}

func UpdateDeviceTimeCreated(t *testing.T, device models.Device) models.Device {

	TimeLeftForGarbaging := utils.GetEnv("PANTAHUB_GC_UNCLAIMED_EXPIRY")
	duration := ParseDuration(TimeLeftForGarbaging)
	TimeBeforeDuration := time.Now().Local().Add(-duration)
	//log.Print(TimeBeforeDuration)
	TimeBeforeDuration = TimeBeforeDuration.Local().Add(-time.Minute * time.Duration(1)) //decrease 1 min
	//log.Print(TimeBeforeDuration)
	db := db.Session
	c := db.C("pantahub_devices")
	//log.Print("Device id:" + device.ID)

	err := c.Update(
		bson.M{"_id": device.ID},
		bson.M{"$set": bson.M{
			"timecreated": TimeBeforeDuration,
		}})
	if err != nil {
		t.Errorf("internal error calling test server: " + err.Error())
		t.Fail()
	}
	return device
}
func MarkAllUnClaimedDevicesAsGrabage(t *testing.T, deviceCount int) bool {
	response := map[string]interface{}{}
	APIEndPoint := GCAPIUrl + "/markgarbage/devices/unclaimed"
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
	devicesMarked := int(response["devices_marked"].(float64))
	//to handle some already existing unclaimed devices count
	if devicesMarked > deviceCount {
		deviceCount += devicesMarked
	}

	//5.check if device_marked=deviceCount
	if devicesMarked != deviceCount {
		t.Errorf("Error:Devices should be marked is:" + strconv.Itoa(deviceCount) + ", But Devices actually marked is:" + strconv.Itoa(devicesMarked))
		t.Fail()
		return false
	}

	return true

}
func login(t *testing.T) map[string]interface{} {
	response := map[string]interface{}{}
	APIEndPoint := BaseAPIUrl + "/auth/login"

	res, err := resty.R().SetBody(map[string]string{
		"username": "user1",
		"password": "user1",
	}).Post(APIEndPoint)

	if err != nil {
		t.Errorf("internal error calling test server " + err.Error())
		t.Fail()
	}
	if res.StatusCode() != 200 {
		t.Errorf("login without username/password must yield 401")
	}
	err = json.Unmarshal(res.Body(), &response)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}
	return response
}

// createDevice : Register a Device (As User)
func createDevice(t *testing.T) models.Device {
	response := map[string]interface{}{}
	APIEndPoint := BaseAPIUrl + "/devices/"

	res, err := resty.R().SetAuthToken(UTOKEN).SetBody(map[string]string{
		"secret": "123",
	}).Post(APIEndPoint)

	if err != nil {
		t.Errorf("internal error calling test server " + err.Error())
		t.Fail()
	}
	device := models.Device{}
	err = json.Unmarshal(res.Body(), &device)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}
	if res.StatusCode() != 200 {
		log.Print(response)
		t.Fail()
	}
	return device
}

// ParseDuration : Parse Duration referece : https://stackoverflow.com/questions/28125963/golang-parse-time-duration
func ParseDuration(str string) time.Duration {
	durationRegex := regexp.MustCompile(`P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?`)
	matches := durationRegex.FindStringSubmatch(str)

	years := ParseInt64(matches[1])
	months := ParseInt64(matches[2])
	days := ParseInt64(matches[3])
	hours := ParseInt64(matches[4])
	minutes := ParseInt64(matches[5])
	seconds := ParseInt64(matches[6])

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)
	return time.Duration(years*24*365*hour + months*30*24*hour + days*24*hour + hours*hour + minutes*minute + seconds*second)
}

// ParseInt64 : ParseInt64
func ParseInt64(value string) int64 {
	if len(value) == 0 {
		return 0
	}
	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0
	}
	return int64(parsed)
}

/*
func DeleteAllDevices(t *testing.T, device models.Device) (bool, int) {
	db := db.Session
	c := db.C("pantahub_devices")
	//log.Print("Device id:" + device.ID)
	info, err := c.RemoveAll(bson.M{})
	if err != nil {
		t.Errorf("Error Deleting devices " + err.Error())
		t.Fail()
		return false, info.Removed
	}
	return true, info.Removed
}
*/
