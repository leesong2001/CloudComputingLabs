package main

import (
	"fmt"
	"net"
	"strconv"
)

//两阶段提交
func synData(IP_Port string) {
	/*
		协调者--参与者 要求同步恢复
		参与者-参与者 请求同步
	*/
	conn, err := net.Dial("tcp", IP_Port)
	if err != nil {
		fmt.Printf("link to targetParticipant%s failed: %s\n", IP_Port, err.Error())
		return
	}
	defer conn.Close()
	//1.待同步参与者向目标参与者请求同步，req: synget
	conn.Write([]byte(cmd2RESPArr(command{cmdType: synget})))

	/*
		2.loop:
			2.1 目标参与者回复待同步参与者某个key：使用之前的标准set格式  set key val
			2.2 待同步参与者向目标参与者回复ACK：+OK
	*/
	database = make(map[string]string)
	for {
		synSetCmdByte := make([]byte, 1024)
		length, e := conn.Read(synSetCmdByte)
		if e != nil {
			fmt.Println("read error:", err)
			return
		}
		synSetCmdStr := string(synSetCmdByte[:length])
		synSetCmd := parseCmd(synSetCmdStr)
		database[synSetCmd.key[0]] = synSetCmd.value
		//2.2 待同步参与者向目标参与者回复ACK：+OK
		conn.Write([]byte(SUCCESS))
	}
}

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
			5.syndata
			6.synget
			7.synTargetGet
		*/
		cmdType := cmd.cmdType
		if cmdType == heartBeats {
			//响应心跳包
			conn.Write([]byte(HeartBeatsResps))
			continue
		}
		if cmdType == syndata {
			//协调者要求参与者与目标参与者进行数据同步
			/*RESP Arrays 格式 ：req: syndata IP:Port
			req ack:  syndata ACK
			fin: syndata FIN
			fin ack: syndata FIN_ACK
			*/
			IP_Port := cmd.value
			syndataACK := command{cmdType: syndata, value: "ACK"}
			conn.Write([]byte(cmd2RESPArr(syndataACK)))
			synData(IP_Port)
			//fin: syndata FIN
			conn.Write([]byte(cmd2RESPArr(command{cmdType: syndata, value: "FIN"})))
			//fin ack: syndata FIN_ACK
			FIN_ACKByte := make([]byte, 1024)
			length, e := conn.Read(FIN_ACKByte)
			if e != nil {
				fmt.Println("73 read error:", err)
				return
			}
			FIN_ACKStr := string(FIN_ACKByte[:length])
			FIN_ACK := parseCmd(FIN_ACKStr).value
			if FIN_ACK == "FIN_ACK" {
			}

			continue
		}
		if cmdType == synget {
			//当前结点作为目标结点，响应其他待同步的参与者的同步请求
			/*弃用
			RESP Arrays 格式 ：req: synget
						    ack: synget key_1 key_2 ..key_n
			1.待同步参与者向目标参与者请求同步，req: synget
			2.目标参与者回复待同步参与者当前的key[]，ack: synget key_1 key_2 ..key_n
			3.loop:
				3.1 待同步参与者向目标参与者请求数据：使用之前的标准get格式  get key
				3.2 目标参与者回复待同步参与者所请求的key：使用之前的标准set格式  set key val
			*/

			//现行版本
			/*
				1.待同步参与者向目标参与者请求同步，req: synget
				2.loop:
					2.1 目标参与者回复待同步参与者某个key：使用之前的标准set格式  set key val
					2.2 待同步参与者向目标参与者回复ACK：+OK
			*/
			//2.1 目标参与者回复待同步参与者某个key：使用之前的标准set格式  set key val
			for k, value := range database {
				setCmd := command{cmdType: set, value: value}
				setCmd.key = append(setCmd.key, k)
				conn.Write([]byte(cmd2RESPArr(setCmd)))
				//2.2 待同步参与者向目标参与者回复ACK：+OK
				setCmdRespByte := make([]byte, 1024)
				setCmdRespByteLen, e := conn.Read(setCmdRespByte)
				if e != nil {
					fmt.Println("read error:", err)
					break
				}
				setCmdResp := string(setCmdRespByte[:setCmdRespByteLen])
				if setCmdResp == SUCCESS {
				}
			}
			break //同步完成连接断开
		}
		if cmdType == synTargetGet {
			/*
				RESP Arrays 格式 ：req: synTargetGet
								  ack: synTargetGet cntFlag(cntFlag是参与者最近一次执行命令后的计数，每次完成一次两阶段提交计数值+1)
			*/

			//直接返回 最新执行次数 计数值
			synTargetGetResp := command{cmdType: synTargetGet, value: strconv.FormatUint(finCnt2PC, 10)}
			conn.Write([]byte(cmd2RESPArr(synTargetGetResp)))
			continue
		}

		//set del get 两阶段的数据缓存
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
			finCnt2PC++ //最新状态计数变量++ 成功执行一次2PC
		}
		if cmt_rbkType == rollback {
			//什么也不做
			conn.Write([]byte(SUCCESS))
			finCnt2PC++ //最新状态计数变量++  成功执行一次2PC
		}

	}
}
