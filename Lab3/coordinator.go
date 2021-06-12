package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

var cmdlist = make(chan command, 10000) //任务队列
var status = make(chan string, 10000)   //用户操作是否成功
var alive int                           //活着的participants个数
var isalive [10]bool                    //true only if p is the latest version

//不断连接p
func try_dail(p int) {
	for {
		time.Sleep(time.Second * 1) //每隔1s拨号一次
		cn, err := net.Dial("tcp", participantIPPortArr[p])
		if err != nil {
			continue
		}
		connParticipant[p] = cn
		break
	}
	for i, _ := range participantIPPortArr {
		if isalive[i] {
			data_recover(p, i)
			break
		}
	}
}

/*
总共4个阶段的恢复
1. synData IP:PORT 发送给to
2. 从to接收 synData ACK
3. 从to接收 synData FIN
4. synData FIN_ACK 发送给to
*/
func data_recover(to, p int) {
	cn := connParticipant[to]
	cmd := command{cmdType: syndata, value: participantIPPortArr[p]}
	info := cmd2RESPArr(cmd)
	fmt.Println("!!!important: ", cn.RemoteAddr(), " recovering, sync with "+info)
	cn.Write([]byte(info))
	recv := make([]byte, 1024)
	n, err := cn.Read(recv)
	if err != nil {
		fmt.Println("WTF? recovering read error :", err.Error())
		eraseconn(to)
		return
	}
	cmd = parseCmd(string(recv[:n]))
	if cmd.cmdType != syndata || cmd.value != "ACK" {
		fmt.Println("UNKNOW ERROR in recovering, expect syndata ACK, found ", getCmdStr(cmd.cmdType)+" "+cmd.value)
		eraseconn(to)
		return
	}
	n, err = cn.Read(recv)
	if err != nil {
		fmt.Println("WTF? recovering read error in step 2:", err.Error())
		eraseconn(to)
		return
	}
	cmd = parseCmd(string(recv[:n]))
	if cmd.cmdType != syndata || cmd.value != "FIN" {
		fmt.Println("UNKNOW ERROR in recovering, expect syndata FIN, found ", getCmdStr(cmd.cmdType)+" "+cmd.value)
		eraseconn(to)
		return
	}
	cmd = command{cmdType: syndata, value: "FIN_ACK"}
	cn.Write([]byte(cmd2RESPArr(cmd)))
	recoverOK(to)
}

//初次启动的同步
func init_participant() {
	fmt.Println("Start init participants...")
	cmd := command{cmdType: synTargetGet}
	var id [3]uint64
	recv := make([]byte, 1024)
	info := cmd2RESPArr(cmd)
	p := 0
	for i, cn := range connParticipant {
		if cn == nil {
			continue
		}
		cn.Write([]byte(info))
		n, err := cn.Read(recv)
		if err != nil {
			fmt.Println("What? in init_participant, read error " + err.Error())
			eraseconn(i)
			return
		}
		num, _ := strconv.ParseUint(parseCmd(string(recv[:n])).value, 10, 64)
		id[i] = num
		if id[i] > id[p] {
			p = i
		}
	}
	for i, cn := range connParticipant {
		fmt.Println(participantIPPortArr[i], ":id=", id[i])
		if cn == nil {
			continue
		}
		if id[i] != id[p] {
			alive--
			data_recover(i, p)
		}
	}
}

//删除一个participant，死亡
func eraseconn(p int) {
	fmt.Println(participantIPPortArr[p], "has dead!!!!!! dead info**")
	connParticipant[p] = nil
	isalive[p] = false
	alive--
	if alive < 0 {
		fmt.Println("!!!important: eraseconn error, find alive < 0")
	}
	//todo
	//打开一个时钟，间隔连接
	go try_dail(p)
}

//一个participant恢复连接，并且完成同步
func recoverOK(p int) {
	isalive[p] = true
	alive++
}

// c为用户输入，ot为反馈信息
func heartBeatsCheck(c chan command, ot chan string) {
	tick := time.NewTicker(time.Millisecond * 100) //100ms一次心跳
	tmp := make([][]byte, len(heartbeatsCnt))
	var tmpsz [3]int
	for i := 0; i < len(heartbeatsCnt); i++ {
		tmp[i] = make([]byte, 1024)
	}
	for {
		<-tick.C //计时器到达
		//现在不知道有啥用
		for _, cn := range connParticipant {
			if cn == nil {
				continue
			}
			//cn.SetDeadline(time.Now().Add(time.Microsecond * 10000)) //20ms 超时
		}
		var cmd command
		cmd.cmdType = heartBeats
		info := HeartBeatsResps
		select {
		case cmd = <-c:
			info = cmd2RESPArr(cmd)
			fmt.Println("info=" + info)
		default:
		}

		//prepare 第一阶段发包
		for i, _ := range heartbeatsCnt {
			if connParticipant[i] == nil {
				continue
			}
			//给存活结点发送指令包或者是心跳空包
			_, err := connParticipant[i].Write([]byte(info))
			if err != nil {
				eraseconn(i)
				fmt.Println("heartBeatsCheck() prepare Write: ", err.Error())
			}
		}
		for i, _ := range heartbeatsCnt {
			if connParticipant[i] == nil {
				continue
			}
			sz, err := connParticipant[i].Read(tmp[i])
			tmpsz[i] = sz
			// if debugClientHandle {
			// 	fmt.Println("i:" + strconv.Itoa(i) + "  " + string(tmp[i]))
			// }
			if err != nil {
				heartbeatsCnt[i]++
				eraseconn(i)
				fmt.Println(err.Error())
			}
			//fmt.Println(heartbeatsCnt[i])
		}
		if cmd.cmdType != heartBeats { //set get del 等操作，2阶段提交
			if debugClientHandle {
				fmt.Println("cmdType: " + getCmdStr(cmd.cmdType))
			}
			acpcnt := 0
			for i, _ := range heartbeatsCnt {
				if connParticipant[i] == nil {
					continue
				}
				if debugClientHandle {
					fmt.Println("recv from par: " + string(tmp[i][:tmpsz[i]]))
				}
				if string(tmp[i][:tmpsz[i]]) == SUCCESS {
					acpcnt++
				}
			}
			if alive == 0 {
				ot <- FAIL
				continue
			}
			fmt.Println("alive:", alive, "acpcnt:", acpcnt)
			if acpcnt == alive { //二阶段；准备ack阶段收到的赞同投票数与存活节点数一致
				info = str2RESPArr(getCmdStr(commit))
			} else {
				info = str2RESPArr(getCmdStr(rollback))
			}
			for i, _ := range heartbeatsCnt {
				if connParticipant[i] == nil {
					continue
				}
				_, err := connParticipant[i].Write([]byte(info))
				if err != nil {
					fmt.Println("unexpected error:" + err.Error())
				}
			}
			//lisong 20210609 11:11 第四阶段 commit/rollback ack
			var ackInfo string
			for i, _ := range heartbeatsCnt {
				if connParticipant[i] == nil {
					continue
				}
				ackInfoLen, err := connParticipant[i].Read(tmp[i])
				if err != nil {
					heartbeatsCnt[i]++
					eraseconn(i)
					fmt.Println(err.Error())
				} else {
					//只要有一个参与者结点返回数据，认为是成功的？可能需要额外检测
					ackInfo = string(tmp[i][:ackInfoLen])
				}
				//fmt.Println(heartbeatsCnt[i])
			}
			if len(ackInfo) > 0 {
				ot <- ackInfo
			} else {
				ot <- FAIL
			}
			/*
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
				}*/
			//lisong 20210609 11:11第四阶段 commit/rollback ack

		}
	}
}
func clientHandle(conn net.Conn) {
	if debugClientHandle {
		fmt.Println("receive conn from client", conn.LocalAddr(), " || ", conn.RemoteAddr())
		for _, connParticipant_i := range connParticipant {
			if connParticipant_i != nil {
				fmt.Println("clientHandle() partcipant: localAddr ", connParticipant_i.LocalAddr(), "||remote addr ", connParticipant_i.RemoteAddr())
			}
		}
	}
	defer conn.Close() //函数/协程结束时关闭conn
	for {
		cmdRESPArrByte := make([]byte, 1024)
		n, err := conn.Read(cmdRESPArrByte)
		if err != nil {
			fmt.Println("532 read error:", err)
			return
		}
		cmdRESPArrStr := string(cmdRESPArrByte[:n])
		cmd := parseCmd(cmdRESPArrStr)
		cmdlist <- cmd
		res := <-status
		bk := res
		fmt.Println("bk:" + bk)
		conn.Write([]byte(bk))
	}
}

func start_coordinator(l net.Listener) {
	//直接连到服务器
	//监听端口，accept客户端的连接请求
	time.Sleep(time.Second * 1)
	alive = len(participantIPPortArr)
	for i, participantIPPortTmp := range participantIPPortArr {
		cn, err := net.Dial("tcp", participantIPPortTmp)
		if err != nil {
			fmt.Printf("link to participant%s failed: %s\n", participantIPPortTmp, err.Error())
			eraseconn(i)
			continue
		}
		connParticipant[i] = cn
		defer connParticipant[i].Close()
	}
	init_participant()
	go heartBeatsCheck(cmdlist, status) //开启心跳
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("coordinatorIPPort accept error:", err)
			return
		}
		fmt.Println("client dail: ", conn)
		clientHandle(conn)
	}
}
