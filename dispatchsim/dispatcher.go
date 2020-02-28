package dispatchsim

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Dispatcher struct {
	E                    *Environment
	Response             chan DriverMatchingResult
	DriverAgents         map[int]*DriverAgent
	DriverAgentsResponse map[int]int // 0 - havent response, 1 - accepted, 2 - rejected
	MatchingLimit        int
	MatchingDrivers      chan DriverAgent
	RejectedDrivers      chan DriverAgent
	MatchingTasks        chan Task
	NoOfTaskTaken        int
}

func SetupDispatcher(e *Environment) Dispatcher {
	return Dispatcher{
		E:                    e,
		Response:             make(chan DriverMatchingResult, 10000),
		DriverAgents:         make(map[int]*DriverAgent),
		DriverAgentsResponse: make(map[int]int),
		MatchingLimit:        10,
		MatchingDrivers:      make(chan DriverAgent, 1000),
		RejectedDrivers:      make(chan DriverAgent, 1000),
		MatchingTasks:        make(chan Task, 1000),
		NoOfTaskTaken:        0,
	}
}

type DriverMatchingResult struct {
	Accept bool
	Id     int
}

func (dis *Dispatcher) dispatcher3(e *Environment) {
	fmt.Printf("[Dispatcher %d]Awaiting to start\n", e.Id)
	<-dis.E.S.StartDispatchers
	fmt.Printf("[Dispatcher %d]Started\n", e.Id)
	ticker := time.Tick(time.Duration(dis.E.S.DispatcherParameters.DispatchInterval) * time.Millisecond) // default is 500
	for {
	K:
		select {
		case <-e.S.Stop: // main stop
			fmt.Printf("[Dispatcher %d]Stop by main\n", e.Id)
			return
		case <-e.Stop: // stop by environment
			fmt.Printf("[Dispatcher %d]Stop\n", e.Id)
			return
		case <-ticker:
			fmt.Printf("[>Dispatcher]Started\n")
			//startWhole := time.Now()
			// get tasks from environment
			//start := time.Now()
			tasks := dis.GetValuableTasks2(e.TaskQueue, dis.MatchingLimit)
			//elapsed := time.Since(start)
			//log.Printf("Getting tasks %s", elapsed)
			if len(tasks) == 0 {
				break K
			}

			roamingDrivers := make([]*DriverAgent, 0)

			//e.DriverAgentMutex.Lock()
			//start2 := time.Now()
			mmRepFat := e.S.GetMinMaxReputationFatigue()
			//elapsed2 := time.Since(start2)
			//log.Printf("Get Min Max %s", elapsed2)

			//start3 := time.Now()
			// get all roaming drivers
			for _, v := range e.DriverAgents {
				if v.Status == Roaming && v.Valid {
					roamingDrivers = append(roamingDrivers, v)
					v.Status = Allocating // change roaming to allocating to prevent double task allocation when migrating
				}
			}
			//e.DriverAgentMutex.Unlock()
			//elapsed3 := time.Since(start3)
			//log.Printf("Change to allocation %s", elapsed3)

			noOfRoamingDrivers := len(roamingDrivers)

			//start4 := time.Now()
			// sort drivers according to ranking index
			sort.SliceStable(roamingDrivers, func(i, j int) bool {
				return roamingDrivers[i].GetRankingIndex(&mmRepFat) > roamingDrivers[j].GetRankingIndex(&mmRepFat)
			})
			//elapsed4 := time.Since(start4)
			//log.Printf("Sort drivers %s", elapsed4)
			for _, d := range roamingDrivers {
				fmt.Printf("[Dispatcher]Driver %d with ranking index %v and total earning %v\n", d.Id, d.GetRankingIndex(&mmRepFat), d.TotalEarnings)
			}
			noOfTasks := len(tasks)

			//fmt.Printf("[Dispatcher]Intial - Drivers:%v, Tasks:%v\n", noOfRoamingDrivers, noOfTasks)
			if noOfTasks > noOfRoamingDrivers {

				// cut tasks
				extraTasks := tasks[noOfRoamingDrivers:]
				tasks = tasks[:noOfRoamingDrivers]
				//fmt.Printf("[Dispatcher]tasks>drivers - Drivers:%v, Tasks:%v\n", len(noOfRoamingDrivers), len(tasks))
				go func() {
					// we need to push away the task back to queue (goroutine)
					for i := 0; i < len(extraTasks); i++ {
						e.TaskQueue <- extraTasks[i]
						dis.NoOfTaskTaken--
					}
				}()
			} else if noOfTasks < noOfRoamingDrivers {

				// cut drivers
				extraDrivers := roamingDrivers[noOfTasks:]
				roamingDrivers = roamingDrivers[:noOfTasks]

				go func() {
					// we need to push the drivers to roaming
					for i := 0; i < len(extraDrivers); i++ {
						extraDrivers[i].Status = Roaming
					}
				}()
			}

			noOfRoamingDrivers = len(roamingDrivers)
			noOfTasks = len(tasks)
			//fmt.Printf("[Dispatcher]Left - Drivers:%v, Tasks:%v\n", noOfRoamingDrivers, noOfTasks)

			// for _, d := range roamingDrivers {
			// 	fmt.Printf("[Dispatcher]Driver %d with ranking index of %v\n", d.Id, d.GetRankingIndex(&mmRepFat))
			// }

			if noOfTasks != noOfRoamingDrivers {
				panic("NOT EQUAL!")
			}

			//go func() {
			for d := 0; d < noOfRoamingDrivers; d++ {
				_, _, distance, waypoints := dis.E.S.RN.G_GetWaypoint(roamingDrivers[d].CurrentLocation, tasks[d].StartCoordinate)
				_, _, _, waypoints2 := dis.E.S.RN.G_GetWaypoint(tasks[d].StartCoordinate, tasks[d].EndCoordinate)
				if distance != math.Inf(1) {
					fmt.Printf("[Dispatcher]Task %v with value %v, distance %v to Driver %d with ranking index of %v\n", tasks[d].Id, tasks[d].FinalValue, tasks[d].Distance, roamingDrivers[d].Id, roamingDrivers[d].GetRankingIndex(&mmRepFat))
					sdw := &StartDestinationWaypoint{
						StartLocation:       roamingDrivers[d].CurrentLocation,
						DestinationLocation: tasks[d].StartCoordinate,
						Waypoint:            waypoints,
					}
					sdw2 := &StartDestinationWaypoint{
						Waypoint: waypoints2,
					}
					roamingDrivers[d].Request <- Message{Task: tasks[d], StartDestinationWaypoint: *sdw, StartDestinationWaypoint2: *sdw2}
					//fmt.Printf("[Dispatcher](Done)Task %v to Driver %d\n", tasks[d].Id, roamingDrivers[d].Id)
				} else {
					//fmt.Printf("[Dispatcher]Task %v to Driver %d (rejected)\n", tasks[d].Id, roamingDrivers[d].Id)
					roamingDrivers[d].Valid = false // turn valid to false for driver, - this driver is in an island
					e.TaskQueue <- tasks[d]
					dis.NoOfTaskTaken--
					//fmt.Printf("[Dispatcher](done)Task %v to Driver %d (rejected)\n", tasks[d].Id, roamingDrivers[d].Id)
				}
			}
			fmt.Printf("[<Dispatcher]Ended\n")
			//}()

			//elapsedWhole := time.Since(startWhole)
			//log.Printf("Dispatcher final end %s", elapsedWhole)
		}
	}

	panic("unreachable")
}

// This function is called when dispatching - could be wrong.
func (dis *Dispatcher) ComputeDriversRegret(drivers []*DriverAgent) {
	// reintialize map
	dis.DriverAgents = make(map[int]*DriverAgent)
	dis.DriverAgentsResponse = make(map[int]int)

	// map array to map
	for k := 0; k < len(drivers); k++ {
		dis.DriverAgents[drivers[k].Id] = drivers[k]
		dis.DriverAgentsResponse[drivers[k].Id] = 0
	}

K:
	for {
		select {
		case r := <-dis.Response:
			fmt.Printf("[ComputeDriversRegret]Response: %v\n", r)
			if r.Accept {
				dis.DriverAgentsResponse[r.Id] = 1 // We know that the driver accepts the task
			} else {
				dis.DriverAgentsResponse[r.Id] = 2 // We know that the driver rejects the task
			}
			var driverCount = 0
			for _, v := range dis.DriverAgentsResponse {
				if v == 0 {
					break
				}
				driverCount++
				if driverCount == len(dis.DriverAgentsResponse) {
					break K
				}
			}
		}
	}

	for k, _ := range dis.DriverAgentsResponse {
		if dis.DriverAgentsResponse[k] == 1 {
			dis.DriverAgents[k].ComputeRegret()
		}
	}
	fmt.Printf("[ComputeDriversRegret]Finish computing regrets for drivers with tasks\n")
}

func (dis *Dispatcher) GetValuableTasks2(TaskQueue chan Task, limit int) []Task {
	//fmt.Printf("[GetValuableTasks2]Getting tasks (Limit:%v)\n", limit)
	tasks := make([]Task, 0)

K:
	// Grabs all tasks from the TaskQueue
	for {
		// Grab all tasks then break.
		if len(tasks) == limit { // TODO: Set toggle max value when choosing list of valuable tasks
			break K
		}
		select {
		case x := <-TaskQueue:
			//fmt.Printf("[GetValuableTasks2]Get Task %v from queue\n", x.Id)
			tasks = append(tasks, x)
			dis.NoOfTaskTaken++
		default:
			// if no more tasks in channel, break.
			break K
		}
	}

	// sort the tasks' value in descending order
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].FinalValue > tasks[j].FinalValue
	})

	//fmt.Printf("[GetValuableTasks]Finish getting tasks - %v\n", len(tasks))
	return tasks

}
