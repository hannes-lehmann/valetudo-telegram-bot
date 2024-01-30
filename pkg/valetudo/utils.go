package valetudo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func ParseRobotState(body []byte) (*RobotState, error) {
	var state RobotState

	err := json.Unmarshal(body, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func ParseRobotStateAttributes(body []byte) (*[]RobotStateAttribute, error) {
	result := []RobotStateAttribute{}

	err := json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (client *ValetudoClient) PushRequest(method string, url string, data interface{}) error {
	requestBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, client.Url+url, bytes.NewReader(requestBytes))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		log.Println(string(body))

		return fmt.Errorf("received unexpected status code: %d", res.StatusCode)
	}

	return nil
}

func (client *ValetudoClient) GetRequest(url string, unmarshalInto any) error {
	res, err := http.Get(client.Url + url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		log.Println(string(body))

		return fmt.Errorf("received unexpected status code: %d", res.StatusCode)
	}

	err = json.Unmarshal(body, unmarshalInto)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

func (client *ValetudoClient) getRobotCapabilityPresets(capability string) (*[]string, error) {
	result := []string{}
	err := client.GetRequest("/api/v2/robot/capabilities/"+capability+"/presets", &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (client *ValetudoClient) setRobotCapabilityPreset(capability string, preset string) error {
	data := PutRobotCapabilityPresetRequest{
		Name: preset,
	}
	err := client.PushRequest("PUT", "/api/v2/robot/capabilities/"+capability+"/preset", &data)

	if err != nil {
		return err
	}

	return nil
}
