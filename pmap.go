package main

import (
	"context"
	"github.com/golang/glog"
	"github.com/shirou/gopsutil/process"
	"io/ioutil"
	"strconv"
	"strings"
)

type ProcessMemorySegment struct {
	//Type embedded from process.MemoryMapsStat
	process.MemoryMapsStat
	stackStart   uint64 `json:"startStack"`
	stackStop    uint64 `json:"stackStop"`
	framePerm	 string `json:"framePerm"`
	frameType    string `json:"frameType"`
}

// MemoryMaps get memory maps from /proc/(pid)/smaps
func GetProcessMemoryMaps(grouped bool, pid int32) (*[]ProcessMemorySegment, error) {
	return GetProcessMemoryMapsWithContext(context.Background(), grouped, pid)
}

func GetProcessMemoryMapsWithContext(ctx context.Context, grouped bool, pid int32) (*[]ProcessMemorySegment, error) {
	var ret []ProcessMemorySegment
	smapsPath := "/proc/" + strconv.Itoa(int(pid)) + "/smaps"
	contents, err := ioutil.ReadFile(smapsPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")

	// function of parsing a block
	getBlock := func(first_line []string, block []string) (ProcessMemorySegment, error) {
		m := ProcessMemorySegment{}
		if len(first_line) > 3 {
			var stacks = strings.Split(first_line[0], "-")
			m.stackStart, err = strconv.ParseUint(stacks[0], 16, 64)
			if err != nil {
				glog.Errorf("Parsing stackStart failed! - ", err)
				return m, err
			}
			m.stackStop, _ = strconv.ParseUint(stacks[1], 16, 64)
			if err != nil {
				glog.Errorf("Parsing stackStart failed!")
				return m, err
			}
			m.framePerm = first_line[1]
			m.Path = first_line[len(first_line)-1]

			if(strings.HasPrefix(m.Path, "/")) {
				m.frameType = "mmap"
			} else {
				m.frameType = m.Path
			}
		}

		for _, line := range block {
			if strings.Contains(line, "VmFlags") {
				continue
			}
			field := strings.Split(line, ":")
			if len(field) < 2 {
				continue
			}
			v := strings.Trim(field[1], " kB") // remove last "kB"
			t, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return m, err
			}

			switch field[0] {
			case "Size":
				m.Size = t
			case "Rss":
				m.Rss = t
			case "Pss":
				m.Pss = t
			case "Shared_Clean":
				m.SharedClean = t
			case "Shared_Dirty":
				m.SharedDirty = t
			case "Private_Clean":
				m.PrivateClean = t
			case "Private_Dirty":
				m.PrivateDirty = t
			case "Referenced":
				m.Referenced = t
			case "Anonymous":
				m.Anonymous = t
			case "Swap":
				m.Swap = t
			}
		}
		return m, nil
	}

	blocks := make([]string, 16)
	for _, line := range lines {

		field := strings.Split(line, " ")
		if strings.HasSuffix(field[0], ":") == false {
			// new block section
			if len(blocks) > 0 {
				g, err := getBlock(field, blocks)
				if err != nil {
					return &ret, err
				}
				ret = append(ret, g)
			}
			// starts new block
			blocks = make([]string, 16)
		} else {
			blocks = append(blocks, line)
		}
	}

	return &ret, nil
}

