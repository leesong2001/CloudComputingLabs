# Lab 2: Your Own HTTP Server

## 1. 概述

基于HTTP/1.1，使用网络编程实现一个HTTP服务器。尝试使用**高并发**编程思想来保证服务器的性能。

### 目标

* 练习基本的网络编程技能，例如使用socket API、抓包分析；
* 熟悉稳定且高性能的并发编程技能。

## 2.服务器运行代码

1. Linux环境下运行，安装go编译器，如：`apt-get install golang`；
2. `go run ThreadPoolLongOption.go`

[Tips]：服务器程序支持参数选项，可以通过指定参数来运行程序。

1. `--ip`							指定服务器IP地址；
2. `--port`                        指定服务器监听端口；
3. `--number-thread`     指定服务器运行线程数。

如：`go run ThreadPoolLongOption.go --ip 127.0.0.1 --port 8888 --
number-thread 8`表示服务器IP地址为“127.0.0.1”，监听“8888”端口，使用线程数为8。

## 3.输入提供格式

1. GET 方法

```
curl -i -X GET http://127.0.0.1:8888/index.html
curl -i -X GET 127.0.0.1:8888/index.html
curl -i -X GET 127.0.0.1:8888/src/index.html
curl -i -X GET 127.0.0.1:8888/src/
curl -i -X GET 127.0.0.1:8888
curl -i -X GET 127.0.0.1:8888/
```

[Tips]：`-i`表示显示HTTP响应报文头部信息；`-X `表示支持不同的方法如GET、POST……


2. POST 方法

```
curl -i -X POST --data "Name=HNU&ID=CS06142" http://127.0.0.1:8888/Post_show
curl -i -X POST --data "Name=HNU&ID=CS06142" http://127.0.0.1:8888/none_Post_show
```

[Tips]：`-data`用于POST方法中，将数据与URL分隔开。

3. DELETE 方法

```
curl -i -X DELETE http://127.0.0.1:8888/none_Post_show
```


## 4.输出解释

1. GET 方法

   对于成功的请求，其返回报文格式如下：

   ```
   HTTP/1.1 200 OK	
   Server: 
   Content-length:
   Content-type:
   
   //.html
   ```

   而对于失败的请求，则返回404状态码：

   ```
   HTTP/1.1 404 Not Found
   Server: 
   Content-length:
   Content-type:
   
   //.html
   ```


2. POST 方法

   对于成功的POST，其返回报文格式如下：

   ```
   HTTP/1.1 200 OK	
   Server: 
   Content-length:
   Content-type:
   
   //.html
   ```

   而对于失败的请求，则返回404状态码：

   ```
   HTTP/1.1 404 Not Found
   Server: 
   Content-length:
   Content-type:
   
   //.html
   ```

3. 其他方法

   对于GET与POST以外的方法，HTTP服务器不做处理，返回501状态码表示服务器不支持当前请求所需要的某个功能：

   ```
   HTTP/1.1 501 Not Implemented
   Server: 
   Content-length:
   Content-type:
   
   //.html
   ```

   
