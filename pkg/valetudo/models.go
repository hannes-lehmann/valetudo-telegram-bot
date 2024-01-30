package valetudo

type RobotStateRemainingAttribute struct {
	Value *int    `json:"value"`
	Unit  *string `json:"unit"`
}

type RobotStateAttribute struct {
	Class       string                        `json:"__class"`
	Type        *string                       `json:"type"`
	SubType     *string                       `json:"subType"`
	Value       *string                       `json:"value"`
	CustomValue *string                       `json:"customValue"`
	Attached    *bool                         `json:"attached,omitempty"`
	Level       *int                          `json:"level,omitempty"`
	Flag        *string                       `json:"flag,omitempty"`
	Remaining   *RobotStateRemainingAttribute `json:"remaining,omitempty"`
}

type RobotStateMap struct {
	Size      RobotStateMapSize     `json:"size"`
	PixelSize int                   `json:"pixelSize"`
	Layers    []RobotStateMapLayer  `json:"layers"`
	Entities  []RobotStateMapEntity `json:"entities"`
}

type RobotStateMapEntity struct {
	Class    string                      `json:"__class"`
	Metadata RobotStateMapEntityMetadata `json:"metaData"`
	Type     string                      `json:"type"`
	Points   *[]int                      `json:"points,omitempty"`
}

type RobotStateMapEntityMetadata struct {
	Angle *float64 `json:"angle,omitempty"`
}

type RobotStateMapSize struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type RobotStateMapLayer struct {
	Type             string                       `json:"type"`
	Metadata         RobotStateMapLayerMetadata   `json:"metaData"`
	Dimensions       RobotStateMapLayerDimensions `json:"dimensions"`
	Pixels           []int                        `json:"pixels"`
	CompressedPixels []int                        `json:"compressedPixels"`
}

type RobotStateMapLayerDimensions struct {
	X RobotStateMapDimensionData `json:"x"`
	Y RobotStateMapDimensionData `json:"y"`
}

type RobotStateMapDimensionData struct {
	Min int `json:"min"`
	Max int `json:"max"`
	Mid int `json:"mid"`
	Avg int `json:"avg"`
}

type RobotStateMapLayerMetadata struct {
	Area      *int    `json:"area"`
	SegmentId *string `json:"segmentId"`
	Active    *bool   `json:"active"`
	Name      *string `json:"name"`
}

type RobotState struct {
	Attributes []RobotStateAttribute `json:"attributes"`
	Map        RobotStateMap         `json:"map"`
}

type MapSegmentationCapabilityPutRequest struct {
	Action      string   `json:"action"`
	SegmentIds  []string `json:"segment_ids"`
	Iterations  *int     `json:"iterations,omitempty"`
	CustomOrder *bool    `json:"custom_order,omitempty"`
}

type BasicControlCapabilityRequest struct {
	Action string `json:"action"`
}
