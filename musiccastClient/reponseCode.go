package musiccastClient

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

var responseCodeMap = map[int]string{
	0:   "Successful request",
	1:   "Initializing",
	2:   "Internal Error",
	3:   "Invalid Request (A method did not exist, a method wasn’t appropriate etc.)",
	4:   "Invalid Parameter (Out of range, invalid characters etc.)",
	5:   "Guarded (Unable to setup in current status etc.)",
	6:   "6 Time Out",
	100: "Access Error",
	101: "Other Errors",
	102: "Wrong User Name",
	103: "Wrong Password",
	104: "Account Expired",
	105: "Account Disconnected/Gone Off/Shut Down",
	106: "Account Number Reached to the Limit",
	107: "Server Maintenance",
	108: "Invalid Account",
	109: "License Error",
	110: "Read Only Mode",
	111: "Max Stations",
	112: "Access Denied",
	113: "There is a need to specify the additional destination Playlist",
	114: "There is a need to create a new Playlist",
	115: "Simultaneous logins has reached the upper limit",
	200: "Linking in progress",
	201: "Unlinking in progress",
}

type returnCodeResponse struct {
	ResponseCode int `json:"response_code"`
}

type featuresResponse struct {
	ResponseCode int          `json:"response_code"`
	Distribution Distribution `json:"Distribution"`
}

type Distribution struct {
	Version          float64   `json:"version"`
	CompatibleClient []float64 `json:"compatible_client"`
}

func MapFeaturesResponse(log *logrus.Logger, response *http.Response) (*Distribution, error) {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("could not parse response")
		return nil, err
	}

	musiccastResponse := &featuresResponse{}
	err = json.Unmarshal([]byte(body), musiccastResponse)
	if err != nil {
		log.Errorf("could not parse response")
		return nil, err
	}
	if musiccastResponse.ResponseCode != 0 {
		log.Error(mapResponseCode(musiccastResponse.ResponseCode))
		return nil, fmt.Errorf(mapResponseCode(musiccastResponse.ResponseCode))
	}
	return &musiccastResponse.Distribution, nil
}

func MapStatusResponse(log *logrus.Logger, response *http.Response) error {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("could not parse response")
		return err
	}

	musiccastResponse := &returnCodeResponse{}
	err = json.Unmarshal([]byte(body), musiccastResponse)
	if err != nil {
		log.Errorf("could not parse response")
		return err
	}
	if musiccastResponse.ResponseCode != 0 {
		log.Error(mapResponseCode(musiccastResponse.ResponseCode))
		return fmt.Errorf(mapResponseCode(musiccastResponse.ResponseCode))
	}
	return nil
}

func mapResponseCode(code int) string {
	return responseCodeMap[code]
}
