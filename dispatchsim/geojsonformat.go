package dispatchsim

import "time"

type GeoJSONFormat struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string      `json:"type"`
	Geometry   interface{} `json:"geometry"`
	Properties Properties  `json:"properties"`
}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}
type Geometry2 struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

type Properties struct {
	Information interface{} `json:"information"` // accepts DriverFormat or Task Format
}

type DriverFormat struct {
	Id            int          `json:"id"`
	EnvironmentId int          `json:"environment_id"`
	Status        DriverStatus `json:"status"`
	CurrentTask   string       `json:"current_task_id"`
	Type          string       `json:"type"`
}

type TaskFormat struct {
	Id            string    `json:"id"`
	EnvironmentId int       `json:"environment_id"`
	Type          string    `json:"type"`
	StartPosition []float64 `json:"start_coordinates"`
	EndPosition   []float64 `json:"end_coordinates"`
	WaitStart     time.Time `json:"start_wait"`
	WaitEnd       time.Time `json:"end_wait"`
	TaskEnd       time.Time `json:"task_end"`
	Value         float64   `json:"value"`
	Distance      float64   `json:"distance"`
}

// Virus
type DriverVirusFormat struct {
	Id          int          `json:"id"`
	Virus       Virus        `json:"virus"`
	Status      DriverStatus `json:"status"`
	CurrentTask string       `json:"current_task_id"`
	Type        string       `json:"type"`
}

// Virus
type TaskVirusFormat struct {
	Id            string    `json:"id"`
	Type          string    `json:"type"`
	Virus         Virus     `json:"virus"`
	StartPosition []float64 `json:"start_coordinates"`
	EndPosition   []float64 `json:"end_coordinates"`
	Value         float64   `json:"value"`
	Distance      float64   `json:"distance"`
}
