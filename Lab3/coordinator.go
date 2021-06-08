package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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

const SUCCESS = "+OK\r\n"
const FAIL = "-ERROR\r\n"

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

func eraseconn(p int) {
	connParticipant[p] = nil
	//todo
	//打开一个时钟，间隔连接
}

// c为用户输入，ot为反馈信息
func heartBeatsCheck(c chan command, ot chan string) {
	tick := time.NewTicker(time.Millisecond * 30) //30ms
	for _, cn := range connParticipant {
		cn.SetDeadline(time.Now().Add(time.Microsecond * 20)) //20ms 超时
	}
	tmp := make([][]byte, len(heartbeatsCnt))
	for i := 0; i < len(heartbeatsCnt); i++ {
		tmp[i] = make([]byte, 1024)
	}
	alive := len(heartbeatsCnt) //当前活着的participants
	for {
		<-tick.C //计时器到达
		var cmd command
		cmd.cmdType = heartBeats
		info := "*0\r\n"
		if len(c) > 0 {
			cmd := <-c
			info = cmd2RESPArr(cmd)
		}
		//第一阶段发包
		for i, _ := range heartbeatsCnt {
			if connParticipant[i] == nil {
				continue
			}
			_, err := connParticipant[i].Write([]byte(info))
			if err != nil {
				eraseconn(i)
				alive--
				println(err.Error())
			}
		}
		for i, _ := range heartbeatsCnt {
			if connParticipant[i] == nil {
				continue
			}
			_, err := connParticipant[i].Read(tmp[i])
			if err != nil {
				heartbeatsCnt[i]++
				eraseconn(i)
				alive--
				println(err.Error())
			}
			println(heartbeatsCnt[i])
		}
		if cmd.cmdType != heartBeats { //set等操作，2阶段提交
			acpcnt := 0
			for i, _ := range heartbeatsCnt {
				if connParticipant[i] == nil {
					continue
				}
				if string(tmp[i]) == SUCCESS {
					acpcnt++
				}
			}
			if alive == 0 {
				ot <- FAIL
				continue
			}
			if acpcnt == alive { //二阶段
				info = cmd2RESPArr(command{commit, []string{}, "", ""})
			} else {
				info = cmd2RESPArr(command{rollback, []string{}, "", ""})
			}
			for i, _ := range heartbeatsCnt {
				if connParticipant[i] == nil {
					continue
				}
				_, err := connParticipant[i].Write([]byte(info))
				if err != nil {
					println("unexpected error:" + err.Error())
				}
			}
			if acpcnt == alive {
				if cmd.cmdType == get {
					ot <- cmd.value
				} else if cmd.cmdType == set {
					ot <- SUCCESS
				} else if cmd.cmdType == del {
					ot <- string(tmp[0])[1 : len(tmp[0])-2]
				}
			} else {
				ot <- FAIL
			}
		}
	}
}

//指令的解析与封装
/*
	set key val
	get key
	del key_{0} key_{1}... key_{n}
*/
const (
	set        = 0
	get        = 1
	del        = 2
	heartBeats = 3
	prepare    = 4
	commit     = 5
	rollback   = 6
)

type command struct {
	cmdType int
	key     []string
	value   string
	taskid  string
}

func parseCmd(RESPArraysStr string) command {
	//解析RESP Arrays格式指令
	//*4 $3 SET $7 CS06142 $5 Cloud $9 Computing
	//小字符串直接split

	RESPArraysTmp := strings.Split(RESPArraysStr, "\r\n")
	RESPArraysTmp = RESPArraysTmp[:len(RESPArraysTmp)-1]
	for i, ele := range RESPArraysTmp {
		println(i, ":", ele)
	}
	var RESPArrays []string
	arraySize, _ := strconv.Atoi(RESPArraysTmp[0][1:])
	if arraySize == 0 {
		//心跳包：空包
		heartBeatsPacket := command{cmdType: heartBeats}
		return heartBeatsPacket
	}
	var i = 2
	for {
		if i > arraySize*2 {
			break
		}
		RESPArrays = append(RESPArrays, RESPArraysTmp[i])
		i = i + 2
	}
	cmd := command{}
	cmdTypeStr := RESPArrays[0]
	if cmdTypeStr == "SET" {
		setCmd := command{cmdType: set}
		setCmd.key = append(setCmd.key, RESPArrays[1])
		setCmd.value = RESPArrays[2]
		cmd = setCmd
	}
	if cmdTypeStr == "GET" {
		getCmd := command{cmdType: get}
		getCmd.key = append(getCmd.key, RESPArrays[1])
		cmd = getCmd
	}
	if cmdTypeStr == "DEL" {
		delCmd := command{cmdType: del}
		for i = 1; i < arraySize; i++ {
			delCmd.key = append(delCmd.key, RESPArrays[i])
		}
		cmd = delCmd
	}
	return cmd
}
func cmd2RESPArr(cmd command) string {
	//封装指令为RESP Arrays
	/*
		set        = 0    set key value
		get        = 1    get key
		del        = 2    del key1 key2 keyi
	*/

	valueStrLen := len(cmd.value)
	var valueSplit []string
	if valueStrLen > 0 {
		valueSplitTmp := strings.Split(cmd.value, " ")
		for _, val := range valueSplitTmp {
			valueSplit = append(valueSplit, val)
		}

	}
	RESPArraysSize := 1 + len(cmd.key) + len(valueSplit)
	RESPArraysStr := "*" + strconv.Itoa(RESPArraysSize) + "\r\n"
	//append cmdType
	RESPArraysStr = RESPArraysStr + "$" + strconv.Itoa(len(getCmdStr(cmd.cmdType))) + "\r\n" + getCmdStr(cmd.cmdType) + "\r\n"
	//append key[]
	for _, key := range cmd.key {
		RESPArraysStr = RESPArraysStr + "$" + strconv.Itoa(len(key)) + "\r\n" + key + "\r\n"
	}
	for _, val := range valueSplit {
		RESPArraysStr = RESPArraysStr + "$" + strconv.Itoa(len(val)) + "\r\n" + val + "\r\n"
	}
	return RESPArraysStr
}

//响应客户端请求的工作协程
var debugClientHandle = true

func clientHandle(conn net.Conn) {
	if debugClientHandle {
		fmt.Println("receive client", conn)
	}
	defer conn.Close() //函数/协程结束时关闭conn

	cmdlist := make(chan command, 10000) //任务队列
	status := make(chan string, 10000)   //用户操作是否成功
	go heartBeatsCheck(cmdlist, status)  //开启心跳
	for {
		cmdRESPArrByte := make([]byte, 1024)
		n, err := conn.Read(cmdRESPArrByte)
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		cmdRESPArrStr := string(cmdRESPArrByte[:n])
		cmd := parseCmd(cmdRESPArrStr)
		// cmdType := cmd.cmdType
		// if cmdType == set {

		// } else if cmdType == get {

		// } else if cmdType == del {

		// }
		cmdlist <- cmd
		res := <-status
		bk := str2RESPArr(res)
		fmt.Println(res, bk)
		conn.Write([]byte(bk))
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
	//直接连到服务器
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
