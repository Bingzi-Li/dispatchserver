package dispatchsim

import (
	"fmt"
	"strconv"
	"time"
)

type DriverStatus int

const (
	Roaming DriverStatus = iota
	Matching
	Fetching
	Travelling
)

func (d DriverStatus) String() string {
	return [...]string{"Roaming", "Matching", "Fetching", "Travelling"}[d]
}

type DriverAgent struct {
	E                   *Environment
	Id                  int
	Name                string
	CurrentLocation     LatLng
	Waypoint            []LatLng
	NextLocation        LatLng
	DestinationLocation LatLng
	PendingTask         Task // store task during matching
	CurrentTask         Task // store task when fetching/travelling
	Status              DriverStatus
	Request             chan Message
	Recieve             chan Message // receive
	Send                chan string  // send
	TasksCompleted      int
	TaskHistory         []Task
	Motivation          float64
	Reputation          float64
	Fatigue             float64
	Regret              float64
	ChangeDestination   chan string
	ChangingWaypoint    bool
}

// Random point
//TODO: Generate Address
func CreateDriver(id int, e *Environment) DriverAgent {
	return DriverAgent{
		E:                 e,
		Id:                id,
		Name:              "John",
		Status:            Roaming,
		Request:           make(chan Message),
		Recieve:           make(chan Message),
		Send:              make(chan string),
		TasksCompleted:    0,
		TaskHistory:       make([]Task, 0),
		Motivation:        100,
		Reputation:        0,
		Fatigue:           0,
		Regret:            0,
		ChangeDestination: make(chan string),
		ChangingWaypoint:  false,
	}
}

// func (d *DriverAgent) ProcessTask() {
// 	tick := time.Tick(d.E.S.MasterSpeed * time.Millisecond)
// 	for {
// 		select {
// 		case <-d.E.Quit:
// 			//fmt.Println("Driver " + strconv.Itoa(d.Id) + " ended his day")
// 			return
// 		case <-tick:
// 			switch d.Status {
// 			case Roaming:
// 				fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
// 				d.E.S.Send <- "Driver " + strconv.Itoa(d.Id) + " is " + d.Status.String()
// 				select {
// 				case t := <-d.Request:
// 					d.Status = Matching
// 					d.PendingTask = t.Task
// 					d.CurrentTask = t.Task
// 					//print(&t)
// 					fmt.Printf("Driver %d got Task %v! \n", d.Id, t.Task.Id)
// 				default:

// 				}
// 			case Matching:
// 				fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())

// 				if d.AcceptRide() == true {
// 					d.CurrentTask = d.PendingTask
// 					d.PendingTask = Task{Id: "null"}
// 					d.Status = Fetching
// 					d.ComputeRegret()
// 				} else {
// 					// send Tasks to TaskQueue
// 					d.E.TaskQueue <- d.CurrentTask
// 					d.PendingTask = Task{Id: "null"}
// 					d.Status = Roaming
// 					// set nil to d.Task
// 				}
// 				fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
// 			case Fetching:
// 				fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
// 			K:
// 				for {
// 					reached, ok := d.ReachTaskPosition()
// 					if ok == true {
// 						if reached == true {
// 							// driver has arrived at task location
// 							break K
// 						} else {
// 							fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
// 						}
// 					}
// 				}
// 				// driver has arrived at task location
// 				d.CurrentTask.WaitEnd = time.Now()
// 				d.Status = Travelling
// 			case Travelling:

// 			F:
// 				for {
// 					reached, ok := d.ReachTaskDestination()
// 					if ok == true {
// 						if reached == true {
// 							break F
// 						} else {
// 							fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
// 						}
// 					}
// 				}
// 				// driver has completed the task
// 				d.CurrentTask.TaskEnded = time.Now()
// 				d.E.FinishQueue <- d.CurrentTask // put into FinishQueue
// 				d.CompleteTask()
// 				d.ComputeFatigue()
// 				d.ComputeMotivation()
// 				d.Status = Roaming
// 				fmt.Println("Driver " + strconv.Itoa(d.Id) + " is completed the ride")
// 				fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
// 			}
// 		}
// 	}
// }

func (d *DriverAgent) ProcessTask2() {
	go d.Drive()

	tick := time.Tick(d.E.S.MasterSpeed * time.Millisecond)
	for {
		select {
		case <-d.E.Quit:
			//fmt.Println("Driver " + strconv.Itoa(d.Id) + " ended his day")
			return
		case <-tick:
			switch d.Status {
			case Roaming:
				fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
				d.E.S.Send <- "Driver " + strconv.Itoa(d.Id) + " is " + d.Status.String()
				select {
				case t := <-d.Request:
					fmt.Printf("Task address @ during matching: %p\n", &t)
					d.Status = Matching
					d.PendingTask = t.Task
					d.CurrentTask = t.Task
					fmt.Printf("Driver %d got Task %v! \n", d.Id, t.Task.Id)
				default:

				}
			case Matching:
				d.E.S.Send <- "Driver " + strconv.Itoa(d.Id) + " is " + d.Status.String()

				if d.AcceptRide() == true {
					d.ChangingWaypoint = true
					d.CurrentTask = d.PendingTask
					d.PendingTask = Task{Id: "null"} // bad way...
					d.Status = Fetching
					d.E.Tasks[d.CurrentTask.Id].WaitStart = time.Now()
					d.ComputeRegret()
					d.DestinationLocation = d.CurrentTask.StartPosition
					d.ChangeDestination <- "change message"
				} else {
					// send Tasks to TaskQueue
					d.E.TaskQueue <- d.CurrentTask
					d.PendingTask = Task{Id: "null"}
					d.Status = Roaming
					// set nil to d.Task
				}
				fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
			case Fetching:
				fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
				d.E.S.Send <- "Driver " + strconv.Itoa(d.Id) + " is " + d.Status.String()
				reached, ok := d.ReachTaskPosition()
				if ok == true && reached == true {
					fmt.Printf("Driver %d has reached the passenger \n", d.Id)
					// driver has arrived at task location
					d.E.Tasks[d.CurrentTask.Id].WaitEnd = time.Now()
					d.CurrentTask.WaitEnd = time.Now()
					d.E.Tasks[d.CurrentTask.Id].WaitEnd = time.Now()
					d.Status = Travelling
					d.DestinationLocation = d.CurrentTask.EndPosition
					d.E.S.Send <- "2,2," + strconv.Itoa(d.E.Id) + "," + strconv.Itoa(d.Id) + "," + latLngToString(d.CurrentLocation) + "," + latLngToString(d.CurrentTask.EndPosition)
				}
			case Travelling:
				fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
				d.E.S.Send <- "Driver " + strconv.Itoa(d.Id) + " is " + d.Status.String()
				reached, ok := d.ReachTaskDestination()
				if ok == true && reached == true {
					// driver has completed the task
					d.CurrentTask.TaskEnded = time.Now()
					d.E.Tasks[d.CurrentTask.Id].TaskEnded = time.Now()
					d.E.FinishQueue <- d.CurrentTask // put into FinishQueue
					d.CompleteTask()
					d.ComputeFatigue()
					d.ComputeMotivation()
					d.Status = Roaming
					fmt.Println("Driver " + strconv.Itoa(d.Id) + " is completed the ride")
					fmt.Printf("Driver %d is %v \n", d.Id, d.Status.String())
					d.DriveToNextPoint()
				}
			}
		}
	}
}

func (d *DriverAgent) Drive() {
	//1. Setting up start, destination and waypoint
	d.E.S.Send <- "2,1," + strconv.Itoa(d.E.Id) + "," + strconv.Itoa(d.Id)
	var generatedNode bool = false
	for {
		select {
		case <-d.E.Quit:
			return
		case x := <-d.Recieve:

			switch x.CommandSecondType {
			case 0:
				fmt.Printf("[Driver %d]Saving random point (lat:%f,lng:%f) to current location \n", d.Id, x.LatLng.Lat, x.LatLng.Lng)
				//d.CurrentLocation = x.LatLng

			case 1:
				fmt.Printf("[Driver %d]Intialization of Driver\n", d.Id)
				//1b. Recieve random current location, destination location and waypoint
				d.CurrentLocation = x.StartDestinationWaypoint.StartLocation
				d.DestinationLocation = x.StartDestinationWaypoint.DestinationLocation
				d.Waypoint = x.StartDestinationWaypoint.Waypoint[1:]

				if generatedNode == false {
					// 2. generate node
					d.E.S.Send <- "2,3," + strconv.Itoa(d.E.Id) + "," + strconv.Itoa(d.Id) + "," + latLngToString(d.CurrentLocation)
				}

			case 2:
				fmt.Printf("[Driver %d]Setting waypoint\n", d.Id)
				d.NextLocation, d.Waypoint = x.Waypoint[1], x.Waypoint[1:]
				//d.CurrentLocation = x.Waypoint[0]
				d.E.S.Send <- "2,4," + strconv.Itoa(d.E.Id) + "," + strconv.Itoa(d.Id) + "," + latLngToString(d.CurrentLocation) + "," + latLngToString(d.NextLocation)

			case 3:
				if x.Success == true {
					fmt.Printf("[Driver %d]Node generated\n", d.Id)
					generatedNode = true
					d.DriveToNextPoint()
				}
			case 4:
				if x.Success == true {
					fmt.Printf("[Driver %d]Driver moved successfully\n", d.Id)
					d.CurrentLocation = x.LocationArrived
					select {
					case <-d.ChangeDestination:
						d.E.S.Send <- "2,2," + strconv.Itoa(d.E.Id) + "," + strconv.Itoa(d.Id) + "," + latLngToString(d.CurrentLocation) + "," + latLngToString(d.CurrentTask.StartPosition) + ",4"
					default:
						d.DriveToNextPoint()
					}

				}
			case 5:
				fmt.Printf("[Driver %d]New random generated\n", d.Id)
				d.CurrentLocation = x.StartDestinationWaypoint.StartLocation
				d.DestinationLocation = x.StartDestinationWaypoint.DestinationLocation
				d.Waypoint = x.StartDestinationWaypoint.Waypoint[1:]
				d.DriveToNextPoint()
			}
		}
	}
}

func (d *DriverAgent) DriveToNextPoint() {
	// e.g [1,1],[2,2],[3,3],[4,4]
	// fmt.Printf("noOfWaypoints: %d\n", len(d.Waypoint))
	// fmt.Println(d.Waypoint)
	// fmt.Println(d.CurrentLocation)
	switch noOfWaypoints := len(d.Waypoint); {
	case noOfWaypoints >= 2:
		d.NextLocation = d.Waypoint[0]
		//d.E.S.Send <-  latLngToString(d.CurrentLocation)
		d.E.S.Send <- "2,4," + strconv.Itoa(d.E.Id) + "," + strconv.Itoa(d.Id) + "," + latLngToString(d.CurrentLocation) + "," + latLngToString(d.NextLocation)
		d.Waypoint = d.Waypoint[1:]
	case noOfWaypoints == 1:

		d.NextLocation = d.Waypoint[0]
		d.E.S.Send <- "2,4," + strconv.Itoa(d.E.Id) + "," + strconv.Itoa(d.Id) + "," + latLngToString(d.CurrentLocation) + "," + latLngToString(d.NextLocation)
		d.Waypoint = d.Waypoint[:0]
	case noOfWaypoints == 0:
		// Driver has reached the last point of the waypoint
		// Generate a new random destination and its waypoint
		fmt.Printf("[Driver %d]No more waypoint to drive\n", d.Id)
		if d.Status == Roaming {
			d.E.S.Send <- "2,5," + strconv.Itoa(d.E.Id) + "," + strconv.Itoa(d.Id) + "," + latLngToString(d.CurrentLocation)
		}
	}

}

// Accept rate
func (d *DriverAgent) AcceptRide() bool {
	return true
	// min := 0
	// max := 1
	// rand.Seed(time.Now().UnixNano())
	// n := min + rand.Intn(max-min+1)
	// if n == 0 {
	// 	d.CurrentTask.WaitStart = time.Now()
	// 	fmt.Println("Driver " + strconv.Itoa(d.Id) + " accepted ride")
	// 	return true
	// } else {
	// 	fmt.Println("Driver " + strconv.Itoa(d.Id) + " declined ride")
	// 	return false
	// }
}

// TODO: implement movement
func (d *DriverAgent) ReachTaskPosition() (bool, bool) {
	fmt.Printf("[Driver %d]CurrentLocation: %v, StartPosition: %v\n", d.Id, d.CurrentLocation, d.CurrentTask.StartPosition)
	if d.CurrentLocation == d.CurrentTask.StartPosition {
		return true, true
	} else {
		return false, true
	}
}

// TODO: implement movement
func (d *DriverAgent) ReachTaskDestination() (bool, bool) {
	fmt.Printf("[Driver %d]CurrentLocation: %v, EndPosition: %v\n", d.Id, d.CurrentLocation, d.CurrentTask.EndPosition)
	if d.CurrentLocation == d.CurrentTask.EndPosition {
		return true, true
	} else {
		return false, true
	}
}

// After driver completed a task, update rating, taskscompleted, fatigue and motivation
func (d *DriverAgent) CompleteTask() (bool, bool) {

	getRating := d.CurrentTask.ComputeRating()
	fmt.Printf("Task %d with value of %d gave Driver %d a rating of %f \n", d.CurrentTask.Id, d.CurrentTask.Value, d.Id, getRating)

	var updatedRating float64 = 0
	if d.TasksCompleted > 0 {
		updatedRating = (d.Reputation + getRating) / 2
	} else {
		updatedRating = d.CurrentTask.ComputeRating()
	}
	updatedTaskCompleted := d.TasksCompleted + 1

	d.TasksCompleted = updatedTaskCompleted
	d.Reputation = updatedRating
	d.TaskHistory = append(d.TaskHistory, d.CurrentTask)
	d.CurrentTask = Task{Id: "null"}

	return true, true
}

// use goroutines?
func (d *DriverAgent) ComputeFatigue() {
	d.Fatigue = d.Fatigue + 1
}

// use goroutines?
func (d *DriverAgent) ComputeMotivation() {
	d.Motivation = d.Motivation + 1
}

// use goroutines?
func (d *DriverAgent) ComputeRegret() {
	fmt.Printf("Computing Regret for Driver %d\n", d.Id)
	fmt.Printf("Regret now: %f, Value of Order: %f, Average value of Orders: %f\n", d.Regret, float64(d.CurrentTask.Value), d.E.ComputeAverageValue(d.Reputation))
	computedRegret := d.Regret - float64(d.CurrentTask.Value) + d.E.ComputeAverageValue(d.Reputation)

	if computedRegret < 0 {
		d.Regret = 0
	} else {
		d.Regret = computedRegret
	}
	fmt.Printf("Computing Regret for Driver %d. Calculated: %f\n", d.Id, d.Regret)
}

func (d *DriverAgent) GetRankingIndex() float64 {
	return (d.Motivation * (d.Reputation - d.Fatigue)) + d.Regret
	// todo : norminlize reputation and fatigue
}
