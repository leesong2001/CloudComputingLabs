package main

import (
	"fmt"
	"net"
	"strconv"
)

//两阶段提交

func coordinatorHandle(conn net.Conn) {
	if debugClientHandle {
		fmt.Println("receive conn from coordinator ", conn.LocalAddr(), " || ", conn.RemoteAddr())
	}
	defer conn.Close() //函数/协程结束时关闭conn
	//监听消息
	//心跳包接收
	//1.准备阶段：接收来自协同者进程的请求报文，携带者要执行的指令
	cmdRESPArrByte := make([]byte, 1024)
	for {
		n, err := conn.Read(cmdRESPArrByte)
		if err != nil {
			fmt.Println("268 read error:", err)
			return
		}
		cmdRESPArrStr := string(cmdRESPArrByte[:n])
		// if debugClientHandle {
		// 	print(cmdRESPArrStr)
		// }
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
			continue
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
		//为了适应CYH目前的代码 直接返回ACK:SUCCESS="+OK\r\n"
		//conn.Write([]byte( str2RESPArr("prepare 1 "+cmd.taskid)  ) )
		//conn.Write([]byte(str2RESPArr("prepare 1")))
		if debugClientHandle {
			fmt.Println("prepare ack : SUCCESS")
		}
		conn.Write([]byte((SUCCESS)))
		//3.commit or rollback
		//4.commit or rollback ack
		n, err = conn.Read(cmdRESPArrByte)
		if err != nil && debugClientHandle {
			fmt.Println("323 read error:", err)
			return
		}
		cmt_rbkRESPArrStr := string(cmdRESPArrByte[:n])
		cmt_rbk := parseCmd(cmt_rbkRESPArrStr)
		cmt_rbkType := cmt_rbk.cmdType
		if debugClientHandle {
			fmt.Println("cmdtype: " + getCmdStr(cmdType))
			fmt.Println("cmt_rbkType: " + getCmdStr(cmt_rbkType))
		}
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
			fmt.Println(participantIPPortArr[0] + ":  ")
			for i, j := range database {
				fmt.Println(i, j)
			}
		}
		if cmt_rbkType == rollback {
			//什么也不做
			conn.Write([]byte(SUCCESS))
		}
	}
}
