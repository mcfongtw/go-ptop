package main

import (
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type JavaThread struct {
	threadname	 string

	nid 		 int

	tid 		 string

	stackPtr	 uint64
}

const THREAD_REGEX = `\"(?P<threadName>[^\"]+)\".*tid=(?P<tid>0x[0-9a-f]+).*nid=(?P<nid>0x[0-9a-f]+).*\[(?P<stackPtr>0x[0-9a-f]+)\]`

func GetJavaThreadDump(targetPid int32) (string, error) {
	var path string = fmt.Sprintf("/tmp/.java_pid%d", targetPid)
	var exist, _ = checkFileExists(path)

	if(!exist) {
		err := startServer(targetPid, path)

		if( err != nil) {
			return "", err
		}
	}

	var transportType = "unix" // or "unixgram" or "unixpacket"
	var laddr = net.UnixAddr{path, transportType}
	socket, err := net.DialUnix(transportType, nil, &laddr)
	//socket, err := net.Dial("unix", path)
	if err != nil {
		glog.Errorf("Dial error", err)
		return "" , err
	}


	sendString(socket,"1")
	sendString(socket, "threaddump")
	var args[3]string
	args[0] = ""
	sendString(socket, args[0])
	args[1] = ""
	sendString(socket, args[1])
	args[2] = "  "
	sendString(socket, args[2])

	glog.V(3).Infof("Asked for dump, waiting for reply...\n")


	res := readString(socket)

	socket.Close()

	return res, nil
}

func startServer(pid int32, udsPath string) (error) {
	glog.V(3).Infof("Socket file does not exist. Asking process to start server...\n")

	var path string = fmt.Sprintf("/proc/%d/cwd/.attach_pid%d", pid, pid)

	file, _ := os.OpenFile(path, os.O_RDWR | os.O_CREATE, 0666)
	file.Write([]byte(""))
	file.Close()

	proc, err := searchProcessByPid(pid)

	if err != nil {
		glog.Errorf("proc [%d] cannot be found! Cause: [%s]", pid, err)
		return err
	}

	proc.SendSignal(unix.SIGQUIT)


	waitForSocketCreation(udsPath, 1 * time.Second, 60)

	return nil
}

func waitForSocketCreation(path string, waitPeriod time.Duration, maxTimeout int64) (bool) {
	glog.V(3).Infof("Waiting for existence of %s...\n", path)
	var result bool = false
	var tick int64 = 0

	for tick < maxTimeout {
		glog.V(3).Info("Current Unix Time: %s\n", time.Now().Unix())
		var exist, _ = checkFileExists(path)
		if(exist) {
			result = true
			break
		} else {
			tick += int64(waitPeriod)
			time.Sleep(waitPeriod)
		}
	}

	return result
}

func checkFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err != nil, err
}



func sendString(socket net.Conn, message string) {
	nBytes, error := socket.Write([]byte(message + "\x00"))

	if error != nil {
		glog.V(3).Infof("Write error:", error)
		return
	} else {
		glog.V(0).Infof("Client sent %s (%d bytes)\n", message, nBytes)
	}
}

func readString(socket net.Conn)(res string) {
	var result string = ""

	buf := make([]byte, 4096)

	for {
		n, err := socket.Read(buf[:])
		if err != nil {
			return result
		}

		packet := string(buf[0:n])
		//log.Printf("Client got: %s", packet)
		result += packet
	}

	return result
}

func parseJavaThreadInfo(jstackOutput string) (map[int]JavaThread) {
	lines := strings.Split(jstackOutput, "\n")
	var result = make(map[int]JavaThread)
	for _, line := range lines {
		params := ParseRegexByGroup(THREAD_REGEX, line)
		if len(params) > 0 {

			assembleJavaThreadInfo := func (paramsMap map[string]string) (JavaThread, error) {
				jthread := JavaThread{}

				jthread.threadname = paramsMap["threadName"]
				jthread.tid = paramsMap["tid"]
				nid, err := strconv.ParseUint(paramsMap["nid"], 0, 16)
				if err != nil {
					glog.V(3).Infof("Parsing jthread.tid has failed: ", err)
					return jthread, err
				}
				jthread.nid = int(nid)
				jthread.stackPtr, err = strconv.ParseUint(paramsMap["stackPtr"], 0, 64)
				if err != nil {
					glog.V(3).Infof("Parsing jthread.stackPtr has failed: ", err)
					return jthread, err
				}
				return jthread, nil
			}

			glog.V(0).Infof("%s\n", line)

			javaThread,_ := assembleJavaThreadInfo(params)

			result[javaThread.nid] = javaThread
		}
	}

	return result
}