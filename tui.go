package main

import (
	"fmt"
	"github.com/gizak/termui"
	"github.com/gizak/termui/extra"
	"github.com/golang/glog"
	"sort"
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
	table.Block.BorderLabel = "PTOP"

	return &TableTabElement{Table: table}
}

func (this *TableTabElement) UpdateThread(listOfMemorySegments *[]TaskMemorySegment) {
	//reset rows
	rows := [][] string {}
	this.Table.Rows = rows

	header := [] string {"stackStart", "stackStop", "task ID", "Wrt Cnt", "Rd Cnt", "Wrt Byte", "Rd Byte", "Type", "Path"}
	this.Table.Rows = append(this.Table.Rows, header)


	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		row := [] string{Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), StringfyInteger(segment.taskID),
			StringfyUinteger64(segment.WriteCount), StringfyUinteger64(segment.ReadCount), StringfyUinteger64(segment.WriteBytes), StringfyUinteger64(segment.ReadBytes), segment.frameType, segment.Path}
		this.Table.Rows = append(this.Table.Rows, row)
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

		row := [] string{Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), StringfyUinteger64(segment.Rss), StringfyUinteger64(segment.Size),
			segment.framePerm, segment.frameType, segment.Path}
		this.Table.Rows = append(this.Table.Rows, row)

	}
}

func (this *TableTabElement) Update(listOfMemorySegments *[]TaskMemorySegment) {
	//reset rows
	rows := [][] string {}
	this.Table.Rows = rows

	header := [] string {"stackStart", "stackStop", "RSS", "Size", "Type", "Path"}
	this.Table.Rows = append(this.Table.Rows, header)


	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		row := [] string{Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), StringfyUinteger64(segment.Rss), StringfyUinteger64(segment.Size),
			segment.frameType, segment.Path}
		this.Table.Rows = append(this.Table.Rows, row)

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

////////////////////////////////////////////////////////////////


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

const CLOCK_TEXT = "[%s]"

func tuiLoop(pid int32) {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()


	//////////////////////////////////////////////////////////////////////////////

	clockText := termui.NewPar("")
	clockText.Text = fmt.Sprintf(CLOCK_TEXT, time.Now().String())
	clockText.Y = 1
	clockText.Height = 1 // 1 line
	clockText.Width = 100
	clockText.Border = false
	clockText.TextFgColor = termui.ColorWhite
	clockText.TextBgColor = termui.ColorBlue
	parTicker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			clockText.Text = fmt.Sprintf(time.Now().Format("2006-01-02 15:04:05 MST -07:00"))

			termui.Render(clockText)
			<-parTicker.C
		}
	}()


	keybindingText := termui.NewPar("Press <Esc> to quit, Press <Right> or <Left> to switch tabs, <Ctrl-s> to sort by Write Count")
	keybindingText.Y = 2
	keybindingText.Height = 2 // 2 line
	keybindingText.Width = 100  // 100 chars
	keybindingText.Border = false
	keybindingText.TextFgColor = termui.ColorWhite
	keybindingText.TextBgColor = termui.ColorBlue

	//////////////////////////////////////////////////////////////////////////////

	termWidth := 300

	tabpane := extra.NewTabpane()
	tabpane.Y = 4
	tabpane.Width = 50
	tabpane.Border = true

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

	tabAll := extra.NewTab("All")
	allTabElem := NewTableTabElement(termWidth)
	tabAll.AddBlocks(allTabElem.Table)
	/////////////////////////////////////////////

	tabpane.SetTabs(*tabThread, *tabMmap, *tabOthers, *tabAll)
	termui.Render(clockText, keybindingText, tabpane)
	///////////////////////////////////////////////////////////////////////////////

	termui.Handle("<Escape>", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("<Left>", func(termui.Event) {
		tabpane.SetActiveLeft()
		termui.Clear()
		termui.Render(clockText, keybindingText, tabpane)
	})

	termui.Handle("<Right>", func(termui.Event) {
		tabpane.SetActiveRight()
		termui.Clear()
		termui.Render(clockText, keybindingText, tabpane)
	})

	//TODO: remember current configuration. When next tick starts, reload config and render.

	tabpaneTicker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			listOfMemorySegments := ptop(pid)

			listOfJavaThreadSegments := filterJavaThread(listOfMemorySegments)

			//TODO: 1-1 key binding for each column?
			termui.Handle("<C-d>", func(termui.Event) {
				sort.Sort(SortedTaskMemorySegmentVector(*listOfJavaThreadSegments))
				threadTabElem.UpdateThread(listOfJavaThreadSegments)
				termui.Render(clockText, keybindingText, tabpane)
			})

			termui.Handle("<C-s>", func(termui.Event) {
				sort.Sort(WriteCountSortedTaskMemorySegmentVector{*listOfJavaThreadSegments})
				threadTabElem.UpdateThread(listOfJavaThreadSegments)
				termui.Render(clockText, keybindingText, tabpane)
			})

			threadTabElem.UpdateThread(listOfJavaThreadSegments)

			mmapTabElem.UpdateMmap(filterMmap(listOfMemorySegments))

			othersTabElem.Update(filterOthers(listOfMemorySegments))

			allTabElem.Update(listOfMemorySegments)

			termui.Render(tabpane)

			<-tabpaneTicker.C
		}
	}()


	termui.Loop()
}

//TODO: interface filter by topN element

//More efficient way to retrieve JavaThread memory segment
func filterJavaThread(listOfMemorySegments *[]TaskMemorySegment)(*[]TaskMemorySegment) {
	list := []TaskMemorySegment{}

	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		if (segment.frameType == "JavaThread") {
			list = append(list, segment)
		}
	}

	return &list
}

//More efficient way to retrieve mmap memory segment
func filterMmap(listOfMemorySegments *[]TaskMemorySegment)(*[]TaskMemorySegment) {
	list := []TaskMemorySegment{}

	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		if (segment.frameType == "mmap") {
			list = append(list, segment)
		}
	}

	return &list
}

//More efficient way to retrieve others	 memory segment
func filterOthers(listOfMemorySegments *[]TaskMemorySegment)(*[]TaskMemorySegment) {
	list := []TaskMemorySegment{}

	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		if (segment.frameType != "JavaThread" && segment.frameType != "mmap") {
			list = append(list, segment)
		}
	}

	return &list
}

//////////////////////


////////////////////////////////////////////////////////////////

//TODO: We might need a generic way to traverse all sortable columns
type SortedTaskMemorySegmentVector []TaskMemorySegment


func (vector SortedTaskMemorySegmentVector) Len() int           { return len(vector)}
func (vector SortedTaskMemorySegmentVector) Swap(i, j int)      { vector[i], vector[j] = vector[j], vector[i] }
func (vector SortedTaskMemorySegmentVector) Less(i, j int) bool { return vector[i].taskID > vector[j].taskID }


///////////
type WriteCountSortedTaskMemorySegmentVector struct {
	SortedTaskMemorySegmentVector
}

func (vector WriteCountSortedTaskMemorySegmentVector) Less(i, j int) bool { return vector.SortedTaskMemorySegmentVector[i].WriteCount > vector.SortedTaskMemorySegmentVector[j].WriteCount }

