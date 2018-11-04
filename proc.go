package main

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type KernelThread struct {
	pid 		int
	tid 		int

	startStack 	uint64
}

func GetListOfKernelThreadsFromJStack(pid int32, mapOfJavaThread map[int]JavaThread)(*[]KernelThread, error) {
	var listOfKernelThreads []KernelThread

	for tid, jthread := range mapOfJavaThread {
		lwp := KernelThread{}
		lwp.pid = int(pid)
		lwp.tid = tid
		lwp.startStack = jthread.stackPtr


		listOfKernelThreads = append(listOfKernelThreads, lwp)
	}

	return &listOfKernelThreads, nil
}

func GetListOfKernelThreadsFromProcStat(pid int32) (*[]KernelThread, error) {
	proc:= getProcess(pid)
	var listOfKernelThreads []KernelThread

	threadMaps, err := proc.Threads()
	if err != nil {
		log.Panicf("Failed to get number of tasks under /proc/%d/. Cause: [%s]", pid, err)
	}

	for k, _ := range threadMaps {
		lwp := KernelThread{}
		lwp.pid = int(pid)
		lwp.tid = int(k)
		stackaddress, err := GetProcStats(pid, true, k)
		if err != nil {
			//break loop and return error immediately
			return nil, err
		}
		//lwp.startStack = "0x" + strconv.FormatUint(stackaddress, 16)
		lwp.startStack = stackaddress

		listOfKernelThreads = append(listOfKernelThreads, lwp)

	}

	return &listOfKernelThreads, nil
}

func getProcess(pid int32) (proc *process.Process) {
	proc, _ = searchProcessByPid(pid)

	return proc
}

func GetProcStats(pid int32, isLwp bool, tid int32) (uint64, error){
	var statPath string

	if isLwp {
		statPath = "/proc/" + strconv.Itoa(int(pid)) + "/task/" + strconv.Itoa(int(tid)) + "/stat"

	} else {
		statPath = "/proc/" + strconv.Itoa(int(pid)) + "/stat"
	}

	fields, err := GetProcStatFields(pid, statPath)
	if err != nil {
		return 0, err
	}

	i := 1
	for !strings.HasSuffix(fields[i], ")") {
		i++
	}

	startstack, err := strconv.ParseUint(fields[i+26], 10, 64)
	if err != nil {
		return 0, err
	}

	return startstack, nil
}

func GetProcStatFields(pid int32, statPath string) ([]string, error) {
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(string(contents))
	return fields, nil
}

func GetThreadIoStat(pid int32, tid int32) (*process.IOCountersStat, error) {
	var ioPath = "/proc/" + strconv.Itoa(int(pid)) + "/task/" + strconv.Itoa(int(tid)) + "/io"

	ioline, err := ioutil.ReadFile(ioPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(ioline), "\n")
	ret := process.IOCountersStat{}

	for _, line := range lines {
		field := strings.Fields(line)
		if len(field) < 2 {
			continue
		}
		t, err := strconv.ParseUint(field[1], 10, 64)
		if err != nil {
			return nil, err
		}
		param := field[0]
		if strings.HasSuffix(param, ":") {
			param = param[:len(param)-1]
		}
		switch param {
		case "syscr":
			ret.ReadCount = t
		case "syscw":
			ret.WriteCount = t
		case "read_bytes":
			ret.ReadBytes = t
		case "write_bytes":
			ret.WriteBytes = t
		}
	}

	return &ret, nil
}

func searchProcessByPid(target int32) (*process.Process, error) {
	listOfProcesses, _ := process.Processes()

	for _, proc := range listOfProcesses {

		if proc.Pid == target {
			return proc, nil
		}
	}

	return nil, fmt.Errorf("pid %d not found!", target)
}