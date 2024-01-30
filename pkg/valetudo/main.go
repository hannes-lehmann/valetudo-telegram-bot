package valetudo

import (
	"io"
	"net/http"

	"github.com/r3labs/sse/v2"
)

type ValetudoClient struct {
	Url string
}

func Init(url string) ValetudoClient {
	return ValetudoClient{Url: url}
}

func (client *ValetudoClient) GetRobotState() (*RobotState, error) {
	res, err := http.Get(client.Url + "/api/v2/robot/state")

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return ParseRobotState(body)
}

func (client *ValetudoClient) GetRobotStateAttributes() (*[]RobotStateAttribute, error) {
	res, err := http.Get(client.Url + "/api/v2/robot/state/attributes")

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return ParseRobotStateAttributes(body)
}

func (client *ValetudoClient) Start() error {
	err := client.PushRequest("PUT", "/api/v2/robot/capabilities/BasicControlCapability", BasicControlCapabilityRequest{
		Action: "start",
	})

	return err
}

func (client *ValetudoClient) Home() error {
	err := client.PushRequest("PUT", "/api/v2/robot/capabilities/BasicControlCapability", BasicControlCapabilityRequest{
		Action: "home",
	})

	return err
}

func (client *ValetudoClient) Pause() error {
	err := client.PushRequest("PUT", "/api/v2/robot/capabilities/BasicControlCapability", BasicControlCapabilityRequest{
		Action: "pause",
	})

	return err
}

func (client *ValetudoClient) Stop() error {
	err := client.PushRequest("PUT", "/api/v2/robot/capabilities/BasicControlCapability", BasicControlCapabilityRequest{
		Action: "stop",
	})

	return err
}

func (client *ValetudoClient) CleanMapSegments(segmentIds []string, iterations int) error {
	request := MapSegmentationCapabilityPutRequest{
		Action:     "start_segment_action",
		SegmentIds: segmentIds,
	}

	err := client.PushRequest("PUT", "/api/v2/robot/capabilities/MapSegmentationCapability", request)

	return err
}

func (client *ValetudoClient) ListenToStateChanges(callback func(*RobotState, error)) error {
	sseClient := sse.NewClient(client.Url + "/api/v2/robot/state/sse")
	sseClient.Subscribe("messages", func(msg *sse.Event) {
		state, err := ParseRobotState(msg.Data)

		if err != nil {
			callback(nil, err)
			return
		}

		callback(state, nil)
	})

	return nil
}

func (client *ValetudoClient) ListenToStateAttributesChanges(callback func(*[]RobotStateAttribute, error)) error {
	sseClient := sse.NewClient(client.Url + "/api/v2/robot/state/attributes/sse")
	sseClient.Subscribe("messages", func(msg *sse.Event) {
		state, err := ParseRobotStateAttributes(msg.Data)

		if err != nil {
			callback(nil, err)
			return
		}

		callback(state, nil)
	})

	return nil
}

func (client *ValetudoClient) GetRobotCapabilities() (*[]string, error) {
	result := []string{}
	err := client.GetRequest("/api/v2/robot/capabilities", &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (client *ValetudoClient) GetFanSpeedControlCapabilityPresets() (*[]string, error) {
	return client.getRobotCapabilityPresets("FanSpeedControlCapability")
}

func (client *ValetudoClient) SetFanSpeedControlCapabilityPreset(preset string) error {
	return client.setRobotCapabilityPreset("FanSpeedControlCapability", preset)
}

func (client *ValetudoClient) GetWaterUsageControlCapabilityPresets() (*[]string, error) {
	return client.getRobotCapabilityPresets("WaterUsageControlCapability")
}

func (client *ValetudoClient) SetWaterUsageControlCapabilityPreset(preset string) error {
	return client.setRobotCapabilityPreset("WaterUsageControlCapability", preset)
}

func (client *ValetudoClient) GetOperationModeControlCapabilityPresets() (*[]string, error) {
	return client.getRobotCapabilityPresets("OperationModeControlCapability")
}

func (client *ValetudoClient) SetOperationModeControlCapabilityPreset(preset string) error {
	return client.setRobotCapabilityPreset("OperationModeControlCapability", preset)
}
