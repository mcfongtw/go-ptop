package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"os"
	"strconv"
	"time"
)

const DEFAULT_PROFILE_INTERVAL_IN_SECOND = 10

func main() {

	args := os.Args

	if len(args) < 2 {
		printUsage()
		return
	}

	/*
	  Ref: https://github.com/openshift/autoheal/pull/31/commits/d6f3c88cccea70c14b151f9163d267224aeb2acc
	  This is needed to make `glog` believe that the flags have already been parsed, otherwise every log messages is prefixed by an error message stating the the flags haven't been
	  parsed.
	*/
	flag.CommandLine.Parse([]string{})

	var parsedPid,_ = strconv.ParseInt(args[1], 10, 32)
	var pid = int32(parsedPid)

	mainLoop(pid)

	glog.Flush()
}

func printUsage() {
	fmt.Fprintf(os.Stdout, "pmap <pid>\n")
}

func mainLoop(pid int32) {
	for {
		fmt.Println("=======================================================================================")
		fmt.Printf("Current Time: %v\n", time.Now().String())
		go pmap(pid)
		time.Sleep(DEFAULT_PROFILE_INTERVAL_IN_SECOND * time.Second)
	}
}

func pmap(pid int32) {
	var jstackResp, err = GetJavaThreadDump(pid)

	if(err != nil) {
		glog.Fatalf("GetJavaThreadDump Cause: [%s]", err)
	}

	////////////////////////////////////

	mapOfJavaThread := parseJavaThreadInfo(jstackResp)

	for key, jthread := range mapOfJavaThread {
		glog.V(0).Infof("key : %d, val: %s\n", key, jthread)

	}
	////////////////////////////////////

	listOfMemoryInfo, err := GetProcessMemoryMaps(false, pid)

	if err != nil {
		glog.Fatalf("GetProcessMemoryMaps Cause: [%s]", err)
	}

	//for i := 0; i < len(*listOfMemoryInfo); i++ {
	//	mmap := (*listOfMemoryInfo)[i]
	//	glog.Infof("Ref: %v, RSS : %v \t PSS : %v \t anon : %v \t size %v \t Stack Start : %v \t Stack Stop : %v \t Path: %v\n", mmap.Rss, mmap.Pss, mmap.Anonymous, mmap.Referenced, mmap.Size, Stringify64BitAddress(mmap.stackStart), Stringify64BitAddress(mmap.stackStop), mmap.Path)
	//}

	////////////////////////////////////

	listOfKernelThreads, err := GetListOfKernelThreadsFromJStack(pid, mapOfJavaThread)

	if err != nil {
		glog.Fatalf("GetListOfKernelThreadsFromJStack Cause: [%s]", err)
	}

	for i := 0; i < len(*listOfKernelThreads); i++ {
		kthread := (*listOfKernelThreads)[i]
		glog.V(0).Infof("tid : %v, start stack : %v\n", kthread.tid, Stringify64BitAddress(kthread.startStack))

	}

	///////////////////////////////////////

	mappedSegments := associateKernelThreadAndJavaThread(listOfKernelThreads, mapOfJavaThread, listOfMemoryInfo)

	printMemorySegments(listOfMemoryInfo, mappedSegments)
}

func associateKernelThreadAndJavaThread(listOfKernelThreads *[]KernelThread, mapOfJavaThreads map[int]JavaThread, listOfMemorySegments *[]ProcessMemorySegment) (map[uint64]JavaThread) {
	var foundSegments = make(map[int]ProcessMemorySegment)

	for i := 0; i < len(*listOfKernelThreads); i++ {
		kthread := (*listOfKernelThreads)[i]

		for j := 0; j < len(*listOfMemorySegments); j++ {
			segment := (*listOfMemorySegments)[j]
			if kthread.startStack >= segment.stackStart && kthread.startStack <= segment.stackStop {
				foundSegments[kthread.tid] = segment
				break
			}
		}
	}
	//Kernel Thread id : ProcessMemorySegment
	//log.Printf("foundSegments len: %v\n", len(foundSegments))

	var mappedJavaThreadStacks = make(map[uint64]JavaThread)
	for tid, segment := range foundSegments {
		jthread, ok := mapOfJavaThreads[tid]
		if ok {
			glog.V(0).Infof("Found java thread (%v) : %v\n", tid, jthread)

			mappedJavaThreadStacks[segment.stackStart] = jthread
		} else {
			glog.Warningf("java thread (%v) NOT found\n", tid)

		}

	}

	return mappedJavaThreadStacks
}

func printMemorySegments(listOfMemorySegments *[]ProcessMemorySegment, mappedJavaThreadStacks map[uint64]JavaThread) {
	fmt.Printf("[%-18s : %-18s] %9s %9s %9s %-30s\n", "START ADDR", "STOP ADDR", "PSS", "RSS", "DIRTY", "PATH")
	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]


		jthread, ok := mappedJavaThreadStacks[segment.stackStart]
		if ok {
			fmt.Printf("[%-18v : %-18v] %9v %9v %9v [%-30v]\n", Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), segment.Pss, segment.Rss, segment.PrivateDirty, jthread.threadname)
			//mappedJavaThreadStacks[segment] = jthread
		} else {
			fmt.Printf("[%-18v : %-18v] %9v %9v %9v %-30v\n", Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), segment.Pss, segment.Rss, segment.PrivateDirty, segment.Path)
		}

	}
}