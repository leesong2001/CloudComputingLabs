package main

import (
	"fmt"
	"net"
	"os"
)

const filepath = "./record.txt"

func SynWithServer(c net.Conn, data []byte) {
	for {
		recv := make([]byte, 10)
		c.Write(data)
		_, err := c.Read(recv)
		if err != nil {
			// 服务器炸了....
		}
		s := string(recv)[:3]
		if s == "ACK" {
			// 与服务器同步完成
			break
		} else if s == "NAK" {
			// do noting
		} else {
			fmt.Println("Unknow Error")
		}
	}
}

func ServeClient(file *os.File, c net.Conn) {
	fmt.Println("receive ", c)
	defer c.Close()
	// 此处连接三个服务器
	c1, _ := net.Dial("tcp", ":11111")
	// c2, _ := net.Dial("tcp", ":11112")
	// c3, _ := net.Dial("tcp", ":11113")
	defer c1.Close()
	// defer c2.Close()
	// defer c3.Close()
	for {
		data := make([]byte, 1024)
		n, err := c.Read(data)
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		// manager收到消息，查看类型
		s := string(data)
		if s[:3] != "get" {
			// 如果是修改，写入日志，返回ACK or NAK，然后再向服务器中写数据。
			_, err := file.Write(append(data[:n], byte(10)))
			if err != nil {
				fmt.Println("file write false:", err)
				c.Write([]byte("NAK"))
			} else {
				// 向服务器写数据，直到收到ACK为止
				c.Write([]byte("ACK"))
				go SynWithServer(c1, data[:n])
				// go SynWithServer(c2, data)
				// go SynWithServer(c3, data)
			}
		} else {
			// 如果是查询，直接向服务器中查
			// 此处疑问，是否需要查3个服务器上的值。否则如果其中有一个不同步怎么办
			// 当然，也可以选择某一种方式使得服务器的数据始终同步
			// 暂时用最简单的查一个
			c1.Write(data[:n])
			recv := make([]byte, 1024)
			len, err := c1.Read(recv)
			if err != nil {
				// 服务器炸了，目前无实现
			}
			c.Write(recv[:len])
		}
	}
}
func main() {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("listen error:", err)
		return
	}
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_SYNC|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open file error:", err)
		return
	}
	defer file.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			return
		}
		go ServeClient(file, c)
	}
}
