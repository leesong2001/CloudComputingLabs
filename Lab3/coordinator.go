package main

import (
	"fmt"
	"net"
	"time"
)

func eraseconn(p int) {
	connParticipant[p] = nil
	//todo
	//打开一个时钟，间隔连接
}

var cmdlist = make(chan command, 10000) //任务队列
var status = make(chan string, 10000)   //用户操作是否成功
var alive = 3                           //活着的participants
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
		for _, cn := range connParticipant {
			if cn == nil {
				continue
			}
			//cn.SetDeadline(time.Now().Add(time.Microsecond * 10000)) //20ms 超时
		}
		var cmd command
		cmd.cmdType = heartBeats
		info := "*0\r\n"
		if len(c) > 0 {
			cmd = <-c
			info = cmd2RESPArr(cmd)
			fmt.Println("info=" + info)
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
				alive--
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
				alive--
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
					fmt.Println("445: " + string(tmp[i][:tmpsz[i]]))
				}
				if string(tmp[i][:tmpsz[i]]) == SUCCESS {
					acpcnt++
				}
			}
			if alive == 0 {
				ot <- FAIL
				continue
			}
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
					alive--
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
	for i, participantIPPortTmp := range participantIPPortArr {
		cn, err := net.Dial("tcp", participantIPPortTmp)
		if err != nil {
			fmt.Printf("link to participant%s failed: %s\n", participantIPPortTmp, err.Error())
			connParticipant[i] = nil
			continue
		}
		connParticipant[i] = cn
		defer connParticipant[i].Close()
	}
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
