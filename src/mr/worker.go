package mr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
	"time"
)
import "log"
import "net/rpc"
import "hash/fnv"

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }
//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}


//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	// uncomment to send the Example RPC to the master.
	for {
		task := GetTask()
		switch task.Flag{
		case 1:
			time.Sleep(time.Second)
			continue
		case 2:
			return
		}
		switch task.Task.Type_ {
		case mapTask:
			MapTask(mapf,task.Task)
			TaskCompare(task.Task.Type_, task.Task.Id)
		case reduceTask:
			ReduceTask(reducef, task.Task)
			TaskCompare(task.Task.Type_, task.Task.Id)
		default:
			panic("worker panic")
		}
		time.Sleep(time.Second)
	}
}

func MapTask(mapf func(string, string) []KeyValue,task Task) {
	file, err := os.Open(task.Filename)
	if err != nil {
		log.Fatalf("cannot open %v", task.Filename)
		return
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", task.Filename)
		return
	}
	file.Close()
	kva := mapf(task.Filename, string(content))

	fileCount := task.NReduce
	files := make(map[int]*os.File, fileCount)
	encoders := make([]*json.Encoder, fileCount)

	for i := 0;i<fileCount;i++ {
		finalName := intermediateFilename(task.Id, i, task.NReduce)
		intermediateFile, err := os.Create(finalName)
		if err != nil{
			panic("Create file error : " + task.Filename)
		}
		defer intermediateFile.Close()
		files[i] = intermediateFile
		e := syscall.Flock(int(intermediateFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if e != nil {
			return
		}
		defer syscall.Flock(int(intermediateFile.Fd()), syscall.LOCK_UN)
		encoders[i] = json.NewEncoder(intermediateFile)
	}

	//// TODO 可以提前执行一部分 reduce 工作
	//for i := range kva {
	//	reduceNum := ihash(kva[i].Key) % fileCount
	//	fmt.Fprintf(files[reduceNum], "%v %v\n", kva[i].Key, kva[i].Value)
	//}
	for i := range kva {
		reduceNum := ihash(kva[i].Key) % fileCount
		encoders[reduceNum].Encode(kva[i])
	}
}

func ReduceTask(reducef func(string, []string) string,task Task){

	outputFilename := "mr-out-reduce-"+strconv.Itoa(task.Id)+".txt"
	outputFile, err := os.Create(outputFilename)
	if err != nil{
		panic("Create file error : " + outputFilename)
	}
	defer outputFile.Close()
	e := syscall.Flock(int(outputFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if e != nil {
		return
	}
	defer syscall.Flock(int(outputFile.Fd()), syscall.LOCK_UN)

	data := make(map[string][]string)
	m := task.Files
	files := make(map[int]*os.File, m)
	for i := 0;i < m;i++{
		finalName := intermediateFilename(i, task.Id, task.NReduce)
		file, err := os.Open(finalName)
		if err != nil {
			log.Fatalf("cannot open %v", finalName)
			return
		}
		defer file.Close()
		files[i] = file
		// 处理中间文件数据
		decoder := json.NewDecoder(file)
		for {
			kv := new(KeyValue)
			err = decoder.Decode(kv)
			if err != nil {
				break
			}
			data[kv.Key] = append(data[kv.Key], kv.Value)
		}
	}
	for key := range data {
		value := reducef(key, data[key])
		outputFile.WriteString(fmt.Sprintf("%v %v\n", key, value))
	}
}

func GetTask() GetTaskReply{
	args := ExampleArgs{}
	reply := GetTaskReply{}
	call("Master.GetTask", &args, &reply)
	return reply
}

func TaskCompare(type_ int, id int){
	args := TaskCompareArgs{
		Type:type_,
		ID:id,
	}
	reply := ExampleReply{}
	call("Master.TaskCompare", &args, &reply)
}
//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

func intermediateFilename(i int,j int, n int) string {
	return "mr-inter-" + strconv.Itoa(i*n+j)
}