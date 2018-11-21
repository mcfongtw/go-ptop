package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"os"
	"strconv"
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

	tuiLoop(pid)

	//TODO: reoorg logger configuration, i.e. default log directory location etc
	glog.Flush()
}

func printUsage() {
	fmt.Fprintf(os.Stdout, "ptop <pid>\n")
}



