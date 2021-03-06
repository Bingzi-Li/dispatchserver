package dispatchsim

// type RecieveFormat2 struct {
// 	Command       int `json:"command"`
// 	SecondCommand int `json:"second_command"`
// 	Data          interface{}
// }

// for adjusting settings
// 0 0
// pause

// 0 1
type SettingsFormat struct {
	TaskParameters       TaskParametersFormat       `json:"task_parameters"`
	DispatcherParameters DispatcherParametersFormat `json:"dispatcher_parameters"`
	VirusParameters      VirusParameters            `json:"virus_parameters"`
	DriverParameters     DriverParamatersFormat     `json:"driver_parameters"`
}

type TaskParametersFormat struct {
	TaskValueType       string  `json:"task_value_type"`
	ValuePerKM          float64 `json:"value_per_km"`
	PeakHourRate        float64 `json:"peak_hour_rate"`
	ReputationGivenType string  `json:"reputation_given_type"`
	ReputationValue     float64 `json:"reputation_value"`
}

type DriverParamatersFormat struct {
	TravellingMode string  `json:"travelling_mode"`   // node/distance
	TravelInterval int     `json:"travel_interval"`   // travel interval
	SpeedKmPerHour float64 `json:"speed_km_per_hour"` // distance
}

type DispatcherParametersFormat struct {
	DispatchInterval  int     `json:"dispatcher_interval"`
	SimilarReputation float64 `json:"similar_reputation"`
}

// ---- future improvement ----
// 1
// generating environment

// 2
// 0
// not in use

// 2
// 1
// sendIntializationDriver

type IntializationDriverFormat struct {
	StartLocation       string
	DestinationLocation string
	Waypoint            string
}

// 2 2
// sendWaypointsDriver
type SendWaypointDriverFormat struct {
	Waypoint string
}

// 2 3
type SendGenerateResultDriver struct {
}

// 3 1
type CorrectedLocation struct {
	StartCoordinate LatLng
	EndCoordinate   LatLng
	Distance        float64 // distance between the StartCoordinate and EndCoordinate
}

type LngLat struct {
	Lng float64
	Lat float64
}
