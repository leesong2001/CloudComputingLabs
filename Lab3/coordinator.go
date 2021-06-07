package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

/*
add virt_nics
ifconfig lo:0 192.168.66.101/24
ifconfig lo:1 192.168.66.201/24
ifconfig lo:2 192.168.66.202/24
ifconfig lo:3 192.168.66.203/24

coordinator_config:
"mode coordinator\n"\
"coordinator_info 192.168.66.101:8001\n"\
"participant_info 192.168.66.201:8002\n"\
"participant_info 192.168.66.202:8003\n"\
"participant_info 192.168.66.203:8004\n" > ${coordinator_config_path}
*/

//基本配置信息
var mode = "coordinator"
var configPath = "./src/coordinator.conf"
var coordinatorIPPort = "175.10.105.61:8001"

//var participantIPPortArr = [3]string{"192.168.66.201:8002", "192.168.66.202:8003", "192.168.66.203:8004"}
var participantIPPortArr []string
var connParticipant []net.Conn

func readConfig() {
	/*读取配置文件
	设置mode的值，运行在协同者或是参与者模式
	设置coordinatorIPPort与participantIPPort[]的值
	设置
	*/
	f, err := os.Open(configPath)
	if err != nil {
		print(err.Error())
		return
	}
	defer f.Close()
	input := bufio.NewScanner(f)
	for input.Scan() {
		st := input.Text()
		if len(st) >= 4 && st[:4] == "mode" {
			mode = st[5:]
		} else if len(st) >= 16 && st[:16] == "coordinator_info" {
			coordinatorIPPort = st[17:]
		} else if len(st) >= 16 && st[:16] == "participant_info" {
			participantIPPortArr = append(participantIPPortArr, st[17:])
		}
	}
	println(mode)
	println(configPath)
	println(coordinatorIPPort)
	for _, participantIPPort := range participantIPPortArr {
		println(participantIPPort)
	}
}

//心跳检测
var heartbeatsCnt = [3]int{0, 0, 0}

func heartBeatsCheck() {
	for _, cnt := range heartbeatsCnt {
		println(cnt)
	}
}

//指令的解析与封装
/*
	set key val
	get key
	del key_{0} key_{1}... key_{n}
*/
const (
	set = 0
	get = 1
	del = 2
)

type command struct {
	cmdType int
	key     []string
	value   string
}

func parseCmd(RESPArraysStr string) command {
	//解析RESP Arrays格式指令

	//e.g. del cmd:
	cmd := command{cmdType: del}
	cmd.key = append(cmd.key, "k1", "k2", "k3")
	return cmd
}
func cmd2RESPArr(command) string {
	//封装指令为RESP Arrays
	RESPArraysStr := ""

	return RESPArraysStr
}

//响应客户端请求的工作协程
var debugClientHandle = true

func clientHandle(conn net.Conn) {
	if debugClientHandle {
		fmt.Println("receive client", conn)
	}
	defer conn.Close() //函数/协程结束时关闭conn
	for {
		cmdRESPArrByte := make([]byte, 1024)
		n, err := conn.Read(cmdRESPArrByte)
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		cmdRESPArrStr := string(cmdRESPArrByte[:n])
		cmd := parseCmd(cmdRESPArrStr)
		cmdType := cmd.cmdType
		if cmdType == set {

		}
		if cmdType == get {

		}
		if cmdType == del {

		}
	}
}

func main() {
	readConfig()
	//绑定IP:Port
	l, err := net.Listen("tcp", coordinatorIPPort)
	defer l.Close()
	if err != nil {
		fmt.Println("coordinator listen error:", err)
		return
	}
	for _, i := range participantIPPortArr {
		cn, err := net.Dial("tcp", i)
		if err != nil {
			fmt.Printf("link to %s failed: %s\n", i, err.Error())
			continue
		}
		connParticipant = append(connParticipant, cn)
		defer cn.Close()
	}
	//监听端口，accept客户端的连接请求
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("coordinatorIPPort accept error:", err)
			return
		}
		go clientHandle(conn)
	}
}
