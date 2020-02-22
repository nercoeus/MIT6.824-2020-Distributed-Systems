package mr

import (
	"log"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

const(
	initState = 0
	inProgress = 1
	complete = 2
)

const(
	newMaster = iota
	completeMap
	completeReduce
)

const(
	mapTask = 1
	reduceTask = 2
)

type Master struct {
	mapTask []Task
	intermediateFile map[int][]string
	reduceTask []Task
	nReduce int
	masterState int
	end bool
}

type Task struct {
	Type_    int
	Id       int
	Filename string
	State    int
	NReduce  int
	Files    int
}

// Your code here -- RPC handlers for the worker to call.
func (m *Master) GetTask(_ *ExampleArgs, reply *GetTaskReply) error {
	switch m.masterState {
	case newMaster:
		// map task
		for i, task := range m.mapTask {
			if task.State == initState {
				reply.Task.Type_ = task.Type_
				reply.Task.Id = task.Id
				reply.Task.Filename = task.Filename
				reply.Task.State = task.State
				reply.Task.NReduce = task.NReduce
				reply.Flag = 0

				m.mapTask[i].State = inProgress
				return nil
			}
		}
		reply.Flag = 1
	case completeMap:
		// reduce task
		for i, task := range m.reduceTask {
			if task.State == initState {
				reply.Task.Type_ = task.Type_
				reply.Task.Id = task.Id
				reply.Task.Filename = task.Filename
				reply.Task.State = task.State
				reply.Task.NReduce = task.NReduce
				reply.Task.Files = task.Files
				reply.Flag = 0

				m.reduceTask[i].State = inProgress
				return nil
			}
		}
		reply.Flag = 1
	case completeReduce:
		reply.Flag = 2
	}

	return nil
}

func (m *Master)TaskCompare(args *TaskCompareArgs, reply *ExampleReply) error{
	if args.Type == mapTask{
		for i,task := range m.mapTask{
			if task.Id == args.ID{
				m.mapTask[i].State = complete
			}
		}
	} else {
		for i, task := range m.reduceTask {
			if task.Id == args.ID {
				m.reduceTask[i].State = complete
			}
		}
	}
	m.checkTasksComplete()
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	_ = rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockName := masterSock()
	_ = os.Remove(sockName)
	l, e := net.Listen("unix", sockName)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	return m.end
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{
		nReduce:nReduce,
		mapTask: []Task{},
		masterState: newMaster,
		end:false,
	}

	for i,file := range files{
		m.mapTask = append(m.mapTask, Task{
			Type_:    mapTask,
			Id:       i,
			Filename: file,
			State:    initState,
			NReduce:  m.nReduce,
		})
	}

	go m.server()
	return &m
}

func (m *Master)checkTasksComplete() {
	switch m.masterState {
	case newMaster:
		for _, j := range m.mapTask {
			if j.State != complete {
				return
			}
		}
		m.masterState = completeMap
		// 创建 reduce 任务
		for i := 0; i < m.nReduce; i++ {
			m.reduceTask = append(m.reduceTask, Task{
				Type_:    reduceTask,
				Id:       i,
				Filename: "",
				State:    initState,
				NReduce:  m.nReduce,
				Files:    len(m.mapTask),
			})
		}
	case completeMap:
		for _, j := range m.reduceTask {
			if j.State != complete {
				return
			}
		}
		m.masterState = completeReduce
		m.end = true
	case completeReduce:
		m.end = true
	default:
	}
}
