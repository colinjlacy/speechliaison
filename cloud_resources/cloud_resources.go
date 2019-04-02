package cloud_resources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"speechLiason/errors"
	"time"
)

var deviceUrlSegments = [2]string{"https://api.amazonalexa.com/v1/devices/", "/settings/address"}
var timeout = time.Duration(5 * time.Second)

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

func GetDeviceAddress(token, deviceId string) (DeviceAddress, error) {
	req, err := http.NewRequest(http.MethodGet, buildAddressUrl(deviceId), nil)
	if err != nil {
		return DeviceAddress{}, errors.SystemError{JobName: "", UserId: "", Context: "GetDeviceAddress", Log: fmt.Sprintf("could not create a request object to get device address: %s", err)}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	c := http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return DeviceAddress{}, errors.SystemError{JobName: "", UserId: "", Context: "GetDeviceAddress", Log: fmt.Sprintf("could not get device address: %s", err)}
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		b, _ := ioutil.ReadAll(resp.Body)
		return DeviceAddress{}, errors.SystemError{JobName: "", UserId: "", Context: "GetDeviceAddress", Log: fmt.Sprintf("could not get device address: %s, %s", resp.StatusCode, b)}
	}
	var data DeviceAddress
	err = json.NewDecoder(resp.Body).Decode(&data)
	data.PromptedLocation = data.City + ", " + data.StateOrRegion + " " + data.PostalCode
	return data, err
}

func buildAddressUrl(deviceId string) string {
	return deviceUrlSegments[0] + deviceId + deviceUrlSegments[1]
}
