package main

import (
	"fmt"
	"github.com/gizak/termui"
	"github.com/gizak/termui/extra"
	"github.com/golang/glog"
	"strings"
	"time"
)


type TableTabElement struct {
	Table *termui.Table
}

func NewTableTabElement(width int) (*TableTabElement) {
	table := termui.NewTable()
	table.FgColor = termui.ColorBlack
	table.BgColor = termui.ColorDefault
	rows := [][] string {}
	table.Rows = rows
	table.Width = width

	return &TableTabElement{Table: table}
}

func (this *TableTabElement) UpdateThread(listOfMemorySegments *[]TaskMemorySegment) {
	//reset rows
	rows := [][] string {}
	this.Table.Rows = rows

	header := [] string {"stackStart", "stackStop", "task ID", "write Count", "read Count", "Type", "Path"}
	this.Table.Rows = append(this.Table.Rows, header)


	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		if(segment.frameType == "JavaThread") {
			row := [] string{Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), StringfyInteger(segment.taskID),
				StringfyUinteger64(segment.WriteCount), StringfyUinteger64(segment.ReadCount), segment.frameType, segment.Path}
			this.Table.Rows = append(this.Table.Rows, row)
		}

	}
}

func (this *TableTabElement) UpdateMmap(listOfMemorySegments *[]TaskMemorySegment) {
	//reset rows
	rows := [][] string {}
	this.Table.Rows = rows

	header := [] string {"stackStart", "stackStop", "RSS", "Size", "Perm", "Type", "Path"}
	this.Table.Rows = append(this.Table.Rows, header)


	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		if(segment.frameType == "mmap") {
			row := [] string{Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), StringfyUinteger64(segment.Rss), StringfyUinteger64(segment.Size),
				segment.framePerm, segment.frameType, segment.Path}
			this.Table.Rows = append(this.Table.Rows, row)
		}

	}
}

func (this *TableTabElement) UpdateOthers(listOfMemorySegments *[]TaskMemorySegment) {
	//reset rows
	rows := [][] string {}
	this.Table.Rows = rows

	header := [] string {"stackStart", "stackStop", "RSS", "Size", "Type", "Path"}
	this.Table.Rows = append(this.Table.Rows, header)


	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		if(segment.frameType != "mmap" && segment.frameType != "JavaThread") {
			row := [] string{Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), StringfyUinteger64(segment.Rss), StringfyUinteger64(segment.Size),
				segment.frameType, segment.Path}
			this.Table.Rows = append(this.Table.Rows, row)
		}

	}
}

func associateKernelThreadAndJavaThread(pid int32, listOfKernelThreads *[]KernelThread, mapOfJavaThreads map[int]JavaThread, listOfMemorySegments *[]ProcessMemorySegment)(*[]TaskMemorySegment) {
	var foundSegments = make(map[int]*TaskMemorySegment)
	var listOfTaskSegments []TaskMemorySegment

	//Copy ProcessMemorySegment to TaskMemorySegment
	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]
		listOfTaskSegments = append(listOfTaskSegments, NewTaskMemorySegment(segment))
	}

	//associate KernelThreads and ProcessMemorySegment
	for i := 0; i < len(*listOfKernelThreads); i++ {
		kthread := (*listOfKernelThreads)[i]

		for j := 0; j < len(listOfTaskSegments); j++ {
			//call by reference of ProcessMemorySegment
			segment := &((listOfTaskSegments)[j])
			if kthread.startStack >= segment.stackStart && kthread.startStack <= segment.stackStop {
				foundSegments[kthread.tid] = segment
				break
			}
		}
	}
	glog.V(0).Infof("associated memory segments: %v\n", len(foundSegments))


	//associate ProcessMemorySegment and JavaThread
	for tid, segment := range foundSegments {
		jthread, ok := mapOfJavaThreads[tid]
		if ok {
			glog.V(0).Infof("Found java thread (%v) : %v\n", tid, jthread)
			segment.frameType = "JavaThread"
			segment.Path = jthread.threadname
			segment.taskID = jthread.nid

			ioStat, err := GetThreadIoStat(pid, int32(segment.taskID))
			if err != nil {
				glog.Warningf("GetThreadIoStat Cause: [%s]", err)
				continue
			}
			segment.WriteCount = ioStat.WriteCount
			segment.ReadCount = ioStat.ReadCount
			segment.WriteBytes = ioStat.WriteBytes
			segment.ReadBytes = ioStat.ReadBytes


		} else {
			glog.Warningf("java thread (%v) NOT found\n", tid)

		}

	}

	return &listOfTaskSegments
}

func printMemorySegments(listOfMemorySegments *[]TaskMemorySegment) {
	fmt.Printf("[%-18s : %-18s] %9s %9s %9s %9s %9s %9s %9s %-10s %-30s\n", "START ADDR", "STOP ADDR", "PSS", "RSS", "DIRTY", "RD BYTES", "WRT BYTES", "RD CNT", "WRT CNT", "TYPE", "DATA")
	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		if(strings.HasPrefix(segment.Path, "/")){
			fmt.Printf("[%-18v : %-18v] %9v %9v %9v %9v %9v %9v %9v [%-10s] %-30v\n", Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), segment.Pss, segment.Rss, segment.PrivateDirty, segment.ReadBytes, segment.WriteBytes, segment.ReadCount, segment.WriteCount, segment.frameType, segment.Path)
		} else {
			fmt.Printf("[%-18v : %-18v] %9v %9v %9v %9v %9v %9v %9v [%-10s] %-30v\n", Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), segment.Pss, segment.Rss, segment.PrivateDirty, segment.ReadBytes, segment.WriteBytes, segment.ReadCount, segment.WriteCount, segment.frameType, segment.Path)
		}

	}
}

const PARAGRAPH = "[%s] Press q to quit, Press j or k to switch tabs"

func tuiLoop(pid int32) {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	//////////////////////////////////////////////////////////////////////////////

	header := termui.NewPar("[" + time.Now().String() + "] Press q to quit, Press j or k to switch tabs")
	header.Height = 1
	header.Width = 50
	header.Border = false
	header.TextBgColor = termui.ColorBlue
	pTicker := time.NewTicker(time.Second)
	pTickerCount := 1
	go func() {
		for {
			if pTickerCount%2 == 0 {
				header.TextFgColor = termui.ColorRed
			} else {
				header.TextFgColor = termui.ColorWhite
			}
			header.Text = "[" + time.Now().String() + "] Press q to quit, Press j or k to switch tabs"

			pTickerCount++
			<-pTicker.C
		}
	}()

	//////////////////////////////////////////////////////////////////////////////

	termWidth := 300


	tabpane := extra.NewTabpane()
	tabpane.Y = 1
	tabpane.Width = 30
	tabpane.Border = false

	//////////////////////////////////////////////
	tabThread := extra.NewTab("Thread")
	threadTabElem := NewTableTabElement(termWidth)
	tabThread.AddBlocks(threadTabElem.Table)


	tabMmap := extra.NewTab("MMap")
	mmapTabElem := NewTableTabElement(termWidth)
	tabMmap.AddBlocks(mmapTabElem.Table)


	tabOthers := extra.NewTab("Others")
	othersTabElem := NewTableTabElement(termWidth)
	tabOthers.AddBlocks(othersTabElem.Table)
	/////////////////////////////////////////////

	tabpane.SetTabs(*tabThread, *tabMmap, *tabOthers)
	termui.Render(header, tabpane)
	///////////////////////////////////////////////////////////////////////////////

	termui.Handle("q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("j", func(termui.Event) {
		tabpane.SetActiveLeft()
		termui.Render(header, tabpane)
	})

	termui.Handle("k", func(termui.Event) {
		tabpane.SetActiveRight()
		termui.Render(header, tabpane)
	})

	drawTicker := time.NewTicker(time.Second)
	drawTickerCount := 10
	go func() {
		for {
			listOfMemorySegments := ptop(pid)

			threadTabElem.UpdateThread(listOfMemorySegments)

			mmapTabElem.UpdateMmap(listOfMemorySegments)

			othersTabElem.UpdateOthers(listOfMemorySegments)

			termui.Render(header, tabpane)

			drawTickerCount++
			<-drawTicker.C
		}
	}()


	termui.Loop()
}

func ptop(pid int32) (*[]TaskMemorySegment) {
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

	listOfMemorySegment, err := GetProcessMemoryMaps(false, pid)

	if err != nil {
		glog.Fatalf("GetProcessMemoryMaps Cause: [%s]", err)
	}

	//for i := 0; i < len(*listOfMemorySegment); i++ {
	//	mmap := (*listOfMemorySegment)[i]
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

	listOfTaskSegment := associateKernelThreadAndJavaThread(pid, listOfKernelThreads, mapOfJavaThread, listOfMemorySegment)

	//printMemorySegments(listOfTaskSegment)

	return listOfTaskSegment
}