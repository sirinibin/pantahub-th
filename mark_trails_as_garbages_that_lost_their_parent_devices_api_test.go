package main

import (
	"encoding/json"
	"log"
	"strconv"
	"testing"

	"github.com/go-resty/resty"
	"gitlab.com/pantacor/pantahub-gc/db"
	"gitlab.com/pantacor/pantahub-gc/models"
	"gopkg.in/mgo.v2/bson"
)

// TestMain1 : Mark Trails as Garbages that lost their parent devices
func TestMain1(t *testing.T) {
	log.Print("Test:Mark Trails as Garbages that lost their parent devices")
	// Create N-no.of devices(based on the value of "var DeviceCount"
	setUp1(t)
	// PUT markgarbage/trails : Mark Trails as Garbages that lost their parent devices
	log.Print(" Case 1:Mark all created trails as garbages that lost their parent device")
	status, trailsMarked := MarkTrailsAsGarbage(t)
	// check if trails_marked=len(Trails)
	if status == 1 && trailsMarked == len(Trails) {
		log.Print(" Case 1:Passed")
	} else {
		t.Errorf(" Case 1,Error:Trails should be marked is:" + strconv.Itoa(len(Trails)) + ", But Trails actually marked is:" + strconv.Itoa(trailsMarked))
		t.Fail()
	}
	//log.Print(strconv.Itoa(trailsMarked) + " Trails Marked as Garbage")

	// PUT markgarbage/trails : to make sure that there is no trails left to mark
	log.Print(" Case 2:Mark trails when there is no trails left that lost their parent device")

	status, trailsMarked = MarkTrailsAsGarbage(t)
	// check if trails_marked=0
	if status == 1 || trailsMarked == 0 {
		log.Print(" Case 2:Passed\n\n")
	} else {
		t.Errorf(" Case 2,Error:Trails should be marked is:0, But Trails actually marked is:" + strconv.Itoa(trailsMarked))
		t.Fail()
	}
	//log.Print(strconv.Itoa(trailsMarked) + " Trails Marked as Garbage")

	tearDown1(t)

}

func setUp1(t *testing.T) bool {
	//db.Connect()
	ClearOldData(t)

	//1.Login with user/user & Obtain Access token
	login(t)
	//2.Create all devices with UTOKEN, API call: POST /devices
	CreateAllDevices(t)
	//2.Create all trails with DTOKENS, API call: POST /trails/
	CreateAllTrails(t)
	//3.Delete All devices so all trails becomes parentless
	DeleteAllDevices(t)

	return true
}
func tearDown1(t *testing.T) bool {

	//Delete all trails
	if !DeleteAllTrails(t) {
		return false
	}

	return true

}

// MarkTrailsAsGarbage : Mark Trails as Garbages that lost their parent devices
func MarkTrailsAsGarbage(t *testing.T) (int, int) {
	response := map[string]interface{}{}
	APIEndPoint := GCAPIUrl + "/markgarbage/trails"
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
	trailsMarked := int(response["trails_marked"].(float64))

	return status, trailsMarked

}

func CreateAllTrails(t *testing.T) bool {

	for _, device := range Devices {
		//log.Print("Creating trail:" + device.ID.Hex())
		trail := createTrail(t, device)
		Trails = append(Trails, trail)
		//log.Print("Created trail:" + trail.ID.Hex())
	}

	return true
}

// createTrail : Create a trail
func createTrail(t *testing.T, device models.Device) models.Trail {
	//response := map[string]interface{}{}
	APIEndPoint := BaseAPIUrl + "/trails/"
	loginResponse, result := loginDevice(t, device.Prn, device.Secret)
	DTOKEN := ""
	if result {
		DTOKEN = loginResponse["token"].(string)
	}

	res, err := resty.R().SetAuthToken(DTOKEN).Post(APIEndPoint)

	if err != nil {
		t.Errorf("internal error calling test server " + err.Error())
		t.Fail()
	}

	trail := models.Trail{}
	err = json.Unmarshal(res.Body(), &trail)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}
	//log.Print("Trail create response:")
	//log.Print(trail)

	if res.StatusCode() != 200 {
		//log.Print(response)
		t.Fail()
	}
	return trail
}

func loginDevice(t *testing.T, username string, password string) (map[string]interface{}, bool) {
	response := map[string]interface{}{}
	APIEndPoint := BaseAPIUrl + "/auth/login"

	res, err := resty.R().SetBody(map[string]string{
		"username": username,
		"password": password,
	}).Post(APIEndPoint)

	err = json.Unmarshal(res.Body(), &response)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
		return response, false
	}

	if err != nil {
		t.Errorf("internal error calling test server " + err.Error())
		t.Fail()
		return response, false
	}
	if res.StatusCode() != 200 {
		t.Errorf("login without username/password must yield 401")
		t.Fail()
		return response, false
	}

	//UTOKEN = response["token"].(string)
	return response, true
}
func DeleteAllTrails(t *testing.T) bool {
	db := db.Session
	c := db.C("pantahub_trails")
	_, err := c.RemoveAll(bson.M{})
	if err != nil {
		t.Errorf("Error on Removing trail: " + err.Error())
		t.Fail()
		return false
	}
	Trails = []models.Trail{}
	return true

	/*
		for _, trail := range Trails {
			if !DeleteTrail(t, trail) {
				t.Errorf("Something went wrong while deleting trail:" + trail.ID.Hex())
				t.Fail()
				return false
			}
			log.Print("Deleted trail:" + trail.ID.Hex())
		}
	*/

}
func DeleteTrail(t *testing.T, trail models.Trail) bool {
	db := db.Session
	c := db.C("pantahub_trails")
	//log.Print("Device id:" + device.ID)
	err := c.Remove(bson.M{"_id": trail.ID})
	if err != nil {
		t.Errorf("Error on Removing trail: " + err.Error())
		t.Fail()
		return false
	}

	return true
}
