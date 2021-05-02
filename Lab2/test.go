package main

import (
	"fmt"
	"net"
)

// 每一个协程的处理，现在只是简单打印conn
func SingleThreadWork(c1 chan net.Conn) {
	for {
		conn := <-c1
		buf := make([]byte, 1024)
		conn.Read(buf) // 从conn中读取客户端发送的数据内容
		fmt.Println(string(buf))
	}
}

const ThreadNum = 5

func main() {

	listen, err := net.Listen("tcp", ":8888") // 创建用于监听的 socket
	if err != nil {
		fmt.Println("listen err=", err)
		return
	}
	fmt.Println("监听套接字，创建成功, 服务器开始监听。。。")
	defer listen.Close() // 服务器结束前关闭 listener

	connchan := make(chan net.Conn, ThreadNum*100)
	for i := 0; i < ThreadNum; i++ {
		go SingleThreadWork(connchan)
	}
	// 循环等待客户端链接
	for {
		fmt.Println("阻塞等待客户端链接...")
		conn, err := listen.Accept() // 创建用户数据通信的socket

		if err != nil {
			panic("Accept() err=  " + err.Error())
		}
		// 这里准备起一个协程，为客户端服务
		//go accept_request_thread(conn)
		connchan <- conn
	}
}

/*http://127.0.0.1:8888/api/camera/get_ptz?camera_id=1324566666789876543*/
