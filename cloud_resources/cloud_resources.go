package cloud_resources

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"io/ioutil"
	"net/http"
	"os"
	"speechLiason/errors"
	"time"
)

var userEmailUrl = "https://api.amazonalexa.com/v2/accounts/~current/settings/Profile.email"
var deviceUrlSegments = [2]string{"https://api.amazonalexa.com/v1/devices/", "/settings/address"}
var timeout = time.Duration(5 * time.Second)
var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

type DeviceAddress struct {
	StateOrRegion     string `json:"stateOrRegion"`
	City              string `json:"city"`
	CountryCode       string `json:"countryCode"`
	PostalCode        string `json:"postalCode"`
	AddressLine1      string `json:"addressLine1"`
	AddressLine2      string `json:"addressLine2"`
	AddressLine3      string `json:"addressLine3"`
	DistrictOrCountry string `json:"districtOrCounty"`
	PromptedLocation  string
}

type UserMapping struct {
	VoiceUserId string `json:"voiceUserId"`
	SyncUserId  string `json:"syncUserId"`
}

func GetDeviceAddress(token, deviceId string) (DeviceAddress, error) {
	req, err := http.NewRequest(http.MethodGet, buildAddressUrl(deviceId), nil)
	if err != nil {
		return DeviceAddress{}, errors.SystemError{JobName: "", UserId: "", Context: "GetDeviceAddress", Log: fmt.Sprintf("could not create a request object to get device address: %s", err)}
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	c := http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return DeviceAddress{}, errors.SystemError{JobName: "", UserId: "", Context: "GetDeviceAddress", Log: fmt.Sprintf("could not get device address: %s", err)}
	}
	defer resp.Body.Close()
	// TODO: handle permission errors
	if resp.StatusCode > 399 {
		b, _ := ioutil.ReadAll(resp.Body)
		return DeviceAddress{}, errors.SystemError{JobName: "", UserId: "", Context: "GetDeviceAddress", Log: fmt.Sprintf("could not get device address: %d, %s", resp.StatusCode, b)}
	}
	var data DeviceAddress
	err = json.NewDecoder(resp.Body).Decode(&data)
	data.PromptedLocation = data.City + ", " + data.StateOrRegion + " " + data.PostalCode
	return data, err
}

func GetUserEmail(token, voiceUserId string) (userEmail string, err error) {
	req, err := http.NewRequest(http.MethodGet, userEmailUrl, nil)
	if err != nil {
		return "", errors.SystemError{JobName: "", UserId: voiceUserId, Context: "GetUserEmail", Log: fmt.Sprintf("could not create a request object to get user email: %s", err)}
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	c := http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", errors.SystemError{JobName: "", UserId: "", Context: "GetUserEmail", Log: fmt.Sprintf("could not get user email: %s", err)}
	}
	defer resp.Body.Close()
	// TODO: handle permission errors
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode > 399 {
		return "", errors.SystemError{JobName: "", UserId: "", Context: "GetUserEmail", Log: fmt.Sprintf("could not get user email: %d, %s", resp.StatusCode, b)}
	}
	userEmail = string(b)
	return
}

func GetUserMapping(voiceUserId string) (userMapping *UserMapping, err error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("USER_MAPPINGS_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"voiceUserId": {
				S: aws.String(voiceUserId),
			},
		},
	}
	result, err := db.GetItem(input)
	if err != nil {
		return &UserMapping{}, errors.SystemError{JobName: "", UserId: voiceUserId, Context: "GetUserMapping", Log: fmt.Sprintf("could not get user mapping: %s", err)}
	}
	if result.Item == nil {
		return
	}
	userMapping = new(UserMapping)
	err = dynamodbattribute.UnmarshalMap(result.Item, userMapping)
	if err != nil {
		return &UserMapping{}, errors.SystemError{JobName: "", UserId: voiceUserId, Context: "GetUserMapping", Log: fmt.Sprintf("could not unmarshal dynamodb attributes: %s", err)}
	}
	return
}

func SaveUserMapping(voiceUserId, syncUserId string) error {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("USER_MAPPINGS_TABLE")),
		Item: map[string]*dynamodb.AttributeValue{
			"voiceUserId": {
				S: aws.String(voiceUserId),
			},
			"syncUserId": {
				S: aws.String(syncUserId),
			},
		},
	}
	_, err := db.PutItem(input)
	if err != nil {
		return errors.SystemError{UserId: voiceUserId, Context: "SaveUserMapping", Log: fmt.Sprintf("error while persisting user mapping to database: %s", err)}
	}
	return nil
}

func buildAddressUrl(deviceId string) string {
	return deviceUrlSegments[0] + deviceId + deviceUrlSegments[1]
}
