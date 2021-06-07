package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
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
var participantIPPort = "192.168.66.201:8002"

func readConfig() {
	/*读取配置文件
	设置mode的值，运行在协同者或是参与者模式
	设置coordinatorIPPort与participantIPPort[]的值
	设置
	*/
	println(mode)
	println(configPath)
	println(coordinatorIPPort)
	println(participantIPPort)
}

//心跳检测
var heartbeatsCnt = [3]int{0, 0, 0}

func heartBeatsCheck() {
	for _, cnt := range heartbeatsCnt {
		println(cnt)
	}
}

const HeartBeatsResps = "*0\r\n"

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
const SUCCESS = "+OK\r\n"
const FAIL = "-ERROR\r\n"

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
	//大字符串读取？
	/*var RESPArray []string //字符串序列 e.g. []arr={set,key,val}
	var err error
	var i =0
	for{
		if RESPArraysStr[i]=='*'||RESPArraysStr[i]=='$'{
			var eleSize int
			sizeStr:=""
			sizeBeginIdx:=i+1
			chType:=RESPArraysStr[i]
			for{
				if RESPArraysStr[i:i+2]=="\r\n" {
					sizeStr=RESPArraysStr[sizeBeginIdx:i]
					eleSize,err =strconv.Atoi(sizeStr)
					break
				}
				i=i+1
			}
			if chType=='$'{
				i=i+2
			}
		}

		i=i+1
		if i>= len(RESPArraysStr){
			break
		}
	}*/
	return cmd
}
func getCmdStr(cmdType int) string {
	if cmdType == set {
		return "SET"
	}
	if cmdType == get {
		return "GET"
	}
	if cmdType == del {
		return "DEL"
	}
	return ""
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
func str2RESPArr(str string) string {
	stringArr := strings.Split(str, " ")
	RESPArraysStr := "*" + strconv.Itoa(len(stringArr)) + "\r\n"
	for _, ele := range stringArr {
		RESPArraysStr = RESPArraysStr + "$" + strconv.Itoa(len(ele)) + "\r\n" + ele + "\r\n"
	}
	return RESPArraysStr
}

//存储数据的内存空间
var database map[string]string

//响应coordinator请求的工作协程
var debugClientHandle = true

//两阶段提交

func coordinatorHandle(conn net.Conn) {
	if debugClientHandle {
		fmt.Println("receive from coordinator", conn)
	}
	defer conn.Close() //函数/协程结束时关闭conn
	//监听消息
	//心跳包接收
	//1.准备阶段：接收来自协同者进程的请求报文，携带者要执行的指令
	cmdRESPArrByte := make([]byte, 1024)
	n, err := conn.Read(cmdRESPArrByte)
	if err != nil {
		fmt.Println("read error:", err)
		return
	}
	cmdRESPArrStr := string(cmdRESPArrByte[:n])
	cmd := parseCmd(cmdRESPArrStr)
	/*
		1.心跳包
		2.set key val
		3.get key
		4.del key[]
	*/
	cmdType := cmd.cmdType
	if cmdType == heartBeats {
		//响应心跳包
		conn.Write([]byte(HeartBeatsResps))
		return
	}
	var valueArr []string
	var isExistArr []bool
	if cmdType == set {
		//set key val
		val, isExist := database[cmd.key[0]]
		valueArr = append(valueArr, val)
		isExistArr = append(isExistArr, isExist)
	}

	if cmdType == get {
		val, isExist := database[cmd.key[0]]
		valueArr = append(valueArr, val)
		isExistArr = append(isExistArr, isExist)

	}
	if cmdType == del {
		for _, key := range cmd.key {
			val, isExist := database[key]
			valueArr = append(valueArr, val)
			isExistArr = append(isExistArr, isExist)
		}
	}
	//2.prepare ack阶段：响应准备阶段的请求，开始投票
	//目前没有加锁，且由于是严格串行，不存在资源冲突，直接返回ACK:"prepare 1 taskid"
	//conn.Write([]byte( str2RESPArr("prepare 1 "+cmd.taskid)  ) )
	conn.Write([]byte(str2RESPArr("prepare 1")))
	//3.commit or rollback
	//4.commit or rollback ack
	n, err = conn.Read(cmdRESPArrByte)
	if err != nil {
		fmt.Println("read error:", err)
		return
	}
	cmt_rbkRESPArrStr := string(cmdRESPArrByte[:n])
	cmt_rbk := parseCmd(cmt_rbkRESPArrStr)
	cmt_rbkType := cmt_rbk.cmdType
	if cmt_rbkType == commit {
		if cmdType == set {
			database[cmd.key[0]] = cmd.value
			conn.Write([]byte(SUCCESS))
		}
		if cmdType == get {
			if isExistArr[0] {
				//get的值存在
				//conn.Write([]byte(str2RESPArr(valueArr[0]+" "+cmd.taskid)))
				conn.Write([]byte(str2RESPArr(valueArr[0])))
			} else {
				//conn.Write([]byte(str2RESPArr("nil"+" "+cmd.taskid)))
				conn.Write([]byte(str2RESPArr("nil")))
			}
		}
		if cmdType == del {
			delNum := 0 //删除key的总数
			for i, isExist := range isExistArr {
				if isExist {
					delete(database, cmd.key[i])
					delNum += 1
				}
			}
			conn.Write([]byte(":" + strconv.Itoa(delNum) + "\r\n"))
		}
	}
	if cmt_rbkType == rollback {
		//什么也不做
		conn.Write([]byte(SUCCESS))
	}

}
func testparseCmd() {

	RESPArraysStrTest := []string{"*3\r\n$3\r\nSET\r\n$5\r\nkeynm\r\n$7\r\nvaluenm\r\n",
		"*2\r\n$3\r\nGET\r\n$8\r\nnoneitem\r\n",
		"*2\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n",
		"*3\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n",
	}

	cmd := parseCmd(RESPArraysStrTest[0])
	println("cmdType:", cmd.cmdType)
	for i, key := range cmd.key {
		println("key", i, " : ", key)
	}
	println("value : ", cmd.value)
}
func testStr2RESPArr() {
	testStr := []string{"prepare 0", "prepare 1", "prepare",
		"commit 1", "commit",
		"rollback 1", "rollback",
	}
	i := 6
	println(testStr[i], " ==str2RESPArr==> ", str2RESPArr(testStr[i]))
}
func testCmd2RESPArr() {
	cmdArr := []command{}
	cmdArr = append(cmdArr, command{cmdType: set, key: []string{"key1"}, value: "val1"})
	cmdArr = append(cmdArr, command{cmdType: get, key: []string{"key1"}})
	cmdArr = append(cmdArr, command{cmdType: del, key: []string{"key1", "key2", "key3"}})

	i := 1
	println(cmd2RESPArr(cmdArr[i]))
}
func main() {
	database = make(map[string]string)
	readConfig()
	//绑定IP:Port
	l, err := net.Listen("tcp", participantIPPort)
	if err != nil {
		fmt.Println("coordinator listen error:", err)
		return
	}

	//监听端口，accept客户端的连接请求
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("coordinatorIPPort accept error:", err)
			return
		}
		go coordinatorHandle(conn)
	}

}
