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
var DeviceCount = 1
var Devices []models.Device
var Trails []models.Trail

func TestMain2(t *testing.T) {
	log.Print("Inside Test Main2")
	// Create N-no.of devices(based on the value of "var DeviceCount"
	setUp2(t)
	// PUT /markgarbage/devices/unclaimed : which marks all unclaimed devices as garbage
	MarkAllUnClaimedDevicesAsGrabage(t, DeviceCount)
	// 2nd call:PUT /markgarbage/devices/unclaimed : to make sure that there is no devices left to mark
	MarkAllUnClaimedDevicesAsGrabage(t, 0)
	// Delete all created devices
	tearDown2(t)

}

func setUp2(t *testing.T) bool {
	db.Connect()
	//1.Login with user/user & Obtain Access token
	login(t)
	//2.Create all devices with UTOKEN, API call: POST /devices
	CreateAllDevices(t)
	//3.Update device timecreated field to less than PANTAHUB_GC_UNCLAIMED_EXPIRY
	UpdateAllDevicesTimeCreated(t)

	return true
}
func tearDown2(t *testing.T) bool {
	//Delete all devices
	if !DeleteAllDevices(t) {
		return false
	}

	return true

}

func CreateAllDevices(t *testing.T) bool {
	for i := 0; i < DeviceCount; i++ {
		device := createDevice(t)
		Devices = append(Devices, device)
		log.Print("Created device:" + device.ID.Hex())
	}
	return true
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
func DeleteAllDevices(t *testing.T) bool {
	for _, device := range Devices {
		if !DeleteDevice(t, device) {
			t.Errorf("Something went wrong while deleting device:" + device.ID.Hex())
			t.Fail()
			return false
		}
		log.Print("Deleted device:" + device.ID.Hex())
	}
	Devices = []models.Device{}
	return true
}
func DeleteDevice(t *testing.T, device models.Device) bool {
	db := db.Session
	c := db.C("pantahub_devices")
	//log.Print("Device id:" + device.ID)
	err := c.Remove(bson.M{"_id": device.ID})
	if err != nil {
		t.Errorf("Error on Removing: " + err.Error())
		t.Fail()
		return false
	}

	return true
}

func UpdateAllDevicesTimeCreated(t *testing.T) bool {
	for _, device := range Devices {
		if !UpdateDeviceTimeCreated(t, &device) {
			t.Errorf("Something went wrong while updating device timestamp:" + device.ID.Hex())
			t.Fail()
			return false
		}
	}
	return true
}
func UpdateDeviceTimeCreated(t *testing.T, device *models.Device) bool {

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
		return false
	}
	return true
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

	log.Print(strconv.Itoa(devicesMarked) + " Devices Marked as Garbage")

	return true

}
func login(t *testing.T) bool {
	response := map[string]interface{}{}
	APIEndPoint := BaseAPIUrl + "/auth/login"

	res, err := resty.R().SetBody(map[string]string{
		"username": "user1",
		"password": "user1",
	}).Post(APIEndPoint)

	if err != nil {
		t.Errorf("internal error calling test server " + err.Error())
		t.Fail()
		return false
	}
	if res.StatusCode() != 200 {
		t.Errorf("login without username/password must yield 401")
		t.Fail()
		return false
	}
	err = json.Unmarshal(res.Body(), &response)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
		return false
	}
	UTOKEN = response["token"].(string)
	return true
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
