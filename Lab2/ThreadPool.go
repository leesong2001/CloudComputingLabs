package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
	"runtime"
)
const  debug_mode=false
const  rootPath="../Lab2"
const  resp404="HTTP/1.1 404 Not Found\r\n"
const  resp501="HTTP/1.1 501 Not Implemented\r\n"
const  resp200="HTTP/1.1 200 OK\r\n"
const  contentHtml="Content-type: text/html\r\n"
const  contentLen="Content-length: "
const  connectionAlive="Connection: keep-alive\r\n"
const  connectionClose="Connection: close\r\n"
const  TimeoutDuration=500 * time.Millisecond
var ip string
var port string
var ThreadNum int

func readFile(filePath string)([]byte,int){
	f, err := os.OpenFile(filePath, os.O_RDONLY,0600)
	defer f.Close()
	if err !=nil {
		fmt.Println(err.Error())
	} else {
		contentByte,_:=ioutil.ReadAll(f)
		return contentByte,len(contentByte)
	}
	return nil,0
}
func response(statusCode int,conn net.Conn,req string,filePath string,attachment string,longConn bool){
	/* statusCode状态码
	 * filePath 文件路径
	 * conn 客户端连接
	 * req 请求方式
	 * attachment post请求的name-id pair
	 * longConn 连接的长短类型
	 */

	var connectionHeader string
	if(longConn){connectionHeader=connectionAlive} else{connectionHeader=connectionClose}

	if(statusCode==501){
		conn.Write([]byte(resp501))
		conn.Write([]byte(contentLen+strconv.Itoa(0)+"\r\n") )//首部字段
		conn.Write([]byte(connectionHeader))
		conn.Write([]byte("\r\n"))//空行
	}else if(statusCode==404){
		conn.Write([]byte(resp404))
		conn.Write([]byte(contentLen+strconv.Itoa(0)+"\r\n") )//首部字段
		conn.Write([]byte(connectionHeader))
		conn.Write([]byte("\r\n"))//空行
	}else if(statusCode==200){
		if(req=="GET"){
			//conn.Write(200 OK 以及 html文件全部内容)
			data,datasize:=readFile(filePath)
			conn.Write([]byte(resp200))
			conn.Write([]byte(contentHtml))//首部字段
			conn.Write([]byte(contentLen+strconv.Itoa(datasize)+"\r\n") )//首部字段
			conn.Write([]byte(connectionHeader))
			conn.Write([]byte("\r\n"))//空行
			conn.Write(data)

		}else{
			//返回200 OK并回显 "Name"-"ID" pairs
			//conn.Write(200 OK + attachment )
			conn.Write([]byte(resp200))
			conn.Write([]byte(contentLen+strconv.Itoa(len([]byte(attachment)))+"\r\n") )//首部字段
			conn.Write([]byte(connectionHeader))
			conn.Write([]byte("\r\n"+attachment))
		}
	}
}
func fileIsExist(filePath string) (bool, error) {
	/*golang判断文件或文件夹是否存在的方法为使用os.Stat()函数返回的错误值进行判断:

	如果返回的错误为nil,说明文件或文件夹存在
	如果返回的错误类型使用os.IsNotExist()判断为true,说明文件或文件夹不存在
	如果返回的错误为其它类型,则不确定是否在存在*/
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func parseDataField(data string)(string,string,bool){
	var Name string
	var ID string
	//data:   'Name=XXXXX&ID=XXX'

	var splitIdx int
	for i := 0;i<len(data);i++{
		if data[i]=='&'{
			splitIdx=i
			break
		}
	}
	hasNameId:=false
	if data[0:5]=="Name=" {
		Name=data[5:splitIdx]
		if(data[splitIdx+1:splitIdx+4]=="ID="){
			ID=data[splitIdx+4:len(data)]
			hasNameId=true
		}
		if debug_mode{fmt.Println("hasNameId:",hasNameId)}
	}else if data[0:3]=="ID=" {
		ID=data[3:splitIdx]
		if(data[splitIdx+1:splitIdx+6]=="Name="){
			Name=data[splitIdx+6:len(data)]
			hasNameId=true
		}
		if debug_mode{fmt.Println("hasNameId:",hasNameId)}
	}else{
		Name=""
		ID=""
	}
	return Name,ID,hasNameId
}
func handle_request(conn net.Conn)  {
	timeoutDuration := TimeoutDuration
	for {
		var method_bd bytes.Buffer
		var url_bd bytes.Buffer
		//var data_bd strings.Builder

		var i int
		var filePath string
		longConn := false
		// 创建一个新切片， 用作保存数据的缓冲区
		buf := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		n, err := conn.Read(buf) // 从conn中读取客户端发送的数据内容

		if err != nil {
			fmt.Printf("客户端退出 err=%v\n", err)
			conn.Close()
			return
		}

		i = 0
		for (i < n && buf[i] != ' ') {
			//根据空格切分请求方法
			method_bd.WriteByte(buf[i])
			i++;
		}
		for (i < n && buf[i] == ' ') {
			i++
		} //游标移动到url field

		for (i < n && buf[i] != ' ') {
			//根据空格切分请求的url
			url_bd.WriteByte(buf[i])
			i++;
		}
		for (i < n && buf[i] == ' ') {
			i++
		} //游标移动到http version field

		var dataFieldStart int
		for ; i < n; i++ {
			if (buf[i] == '\r' && buf[i+1] == '\n') {
				//  协议版本\r\n首部字段\r\n...首部字段\r\n首部字段\r\n\r\n
				//connection首部字段解析
				if (i+11 < n) {
					if (strings.EqualFold(string(buf[i+2:i+12]), "connection")) {
						for (i < n && buf[i] != ':') {
							i++
						} //buf[i]==':'
						i++
						for (i < n && buf[i] == ' ') {
							i++
						}
						//游标移动到connection value
						valueStartIdx := i

						for ; i < n; i++ {
							if (buf[i] == ' ' || buf[i] == '\r') {
								break;
							}
						}
						if (strings.EqualFold(string(buf[valueStartIdx:i]), "keep-alive")) {
							longConn = true
						}
					}
				}
			}
			if (buf[i] == '\r' && buf[i+1] == '\n' && buf[i+2] == '\r' && buf[i+3] == '\n') {
				//连续的 \r\n\r\n 确定data field 起始
				dataFieldStart = i + 4 //指向data field起始字节
				break
			}
		}

		i = dataFieldStart
		data := string(buf[i:])

		method := method_bd.String()
		url := url_bd.String()

		var j = 0
		for ; j < len(url); j++ {
			if (url[j] == '/') {
				if (j+1 < len(url)) {
					if (url[j+1] != '/') {
						//请求路径 /path...
						break
					}
				} else { //  url 以'/'结尾
					break
				}
			}
		}

		if (j == len(url)) {
			filePath = string('/')
		} else {
			filePath = url[j:len(url)] //   '/path' or '/'
		}

		filePath = rootPath + filePath

		if debug_mode {
			fmt.Println("method", method)
			fmt.Println("url", url)
			fmt.Println("filePath", filePath)
			fmt.Println("ConnectionHeader:", longConn)
		}

		if (method == "GET") {
			/*需求4-2：如果请求的url对应于目录下已经存在的html文件，则返回200 OK以及文件的全部内容。特别的是，要求能够处理带有子目录的url:*/
			if debug_mode {
				fmt.Println("len filePath:", len(filePath))
				fmt.Println("filePath[-5:]:", filePath[len(filePath)-5:len(filePath)])
			}
			if (len(filePath) > 5 && filePath[len(filePath)-5:len(filePath)] == ".html") {
				fileExist, _ := fileIsExist(filePath)
				if (fileExist) {
					//返回200 OK 以及文件的全部内容
					//response(int statusCode,Conn conn,String req,String filePath,String attachment)
					response(200, conn, method, filePath, "", longConn)
				} else {
					//返回 404 Not Found response
					response(404, conn, "", "", "", longConn)
				}
			} else {
				//请求的是目录
				fileExist, _ := fileIsExist(filePath + "/index.html")
				if (fileExist) {
					//返回200 OK 以及index.html文件的全部内容
					response(200, conn, method, filePath+"/index.html", "", longConn)
				} else {
					//返回 404 Not Found response
					response(404, conn, "", "", "", longConn)
				}
			}
		} else if (method == "POST") {
			if (url == "/Post_show") {
				//解析data field
				if debug_mode {
					fmt.Println("data field: ", data)
				}
				Name, ID ,HasNameId:= parseDataField(data)
				attachment := "Your Name: " + Name + "\nYour ID: " + ID
				if HasNameId{
					response(200, conn, method, "", attachment, longConn)
				} else{
					//返回404 Not Found response message.
					response(404, conn, "", "", "", longConn)
				}

			} else {
				//返回404 Not Found response message.
				response(404, conn, "", "", "", longConn)
			}
		} else {
			//既不是GET也不是POST，返回501 Not Implemented error message
			response(501, conn, "", "", "", longConn)
		}
		if(!longConn){
			conn.Close()
			break
		}
	}
}
// 每一个协程的处理，现在只是简单打印conn
func SingleThreadWork(c1 chan net.Conn) {
	for {
		conn := <-c1
		handle_request(conn)
	}
}

func main() {
	ipInput := flag.String("ip", "127.0.0.1", "What is your ip_address?")
	portInput := flag.String("port", "8888", "What is the port?")
	numberThreadInput := flag.Int("number-thread", 1, "How much is the thread number?")

	flag.Parse() //解析输入的参数

	ip=*ipInput
	port=*portInput
	ThreadNum=*numberThreadInput
	runtime.GOMAXPROCS(ThreadNum)

	if debug_mode{
		fmt.Println("ip=", ip)
		fmt.Println("port=", port)
		fmt.Println("ThreadNum=", ThreadNum)
	}
	listen, err := net.Listen("tcp", ":"+port) // 创建用于监听的 socket
	if err != nil {
		if debug_mode{
			fmt.Println("listen err=", err)
		}
		return
	}
	if debug_mode {
		if debug_mode{
			fmt.Println("监听套接字，创建成功, 服务器开始监听。。。")
		}
	}
	defer listen.Close() // 服务器结束前关闭 listener

	connchan := make(chan net.Conn, ThreadNum*100)
	for i := 0; i < ThreadNum; i++ {
		go SingleThreadWork(connchan)
	}
	// 循环等待客户端链接
	for {
		if debug_mode{
			fmt.Println("阻塞等待客户端链接...")
		}
		conn, err := listen.Accept() // 创建用户数据通信的socket

		if err != nil {
			panic("Accept() err=  " + err.Error())
		}
		// 这里准备起一个协程，为客户端服务
		//go accept_request_thread(conn)

		//向任务队列中添加待处理连接conn
		connchan <- conn
	}
}
