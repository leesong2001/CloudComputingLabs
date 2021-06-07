# 心跳检测

总的来说，需要有全局计数器变量序列`HeartbeatsCnt[]`与一个定时发送**心跳包**的函数。

1. 单独的功能模块：定义**心跳包**结构体，定期的向指定接收方发送心跳包，若接收方超时未响应，则在发送方对应的`HeartbeatsCnt`变量++；反之清零。
2. 二阶段提交过程中的判断：若在某一环节接收方超时未响应，则在发送方对应的`HeartbeatsCnt`变量++；反之清零。

# 解析与封装

详细的看指导书**3.2 KV store command formats**

1. bulk string 

   ```
   A $ byte followed by the number of bytes composing the string (a 
   prefixed length), terminated by CRLF.
   The actual string data.
   A final CRLF.
   
   e.g.
   CS06142=>$7\r\nCS06142\r\n
   ```

2. RESP Arrays

   ```
   A * character as the first byte, followed by the number of 
   elements in the array as a decimal number, followed by CRLF.
   Arbitrary number of bulk strings (up to 512 MB in length).
   ```

   `SET CS06142 "Cloud Computing"`
   =>
   `*4\r\n$3\r\nSET\r\n$7\r\nCS06142\r\n$5\r\nCloud\r\n$9\r\nCom
   puting\r\n`

   \*开头代表接下来的数据是RESP Arrays，有多个bulk string

   4表明了数组的元素个数

   然后就是具体的bulk string[i]的内容

# 两阶段提交



# 实现步骤

## makefile编写

编译后可执行文件名为`kvstore2pcsystem`

## 读取配置文件

### 配置文件路径指定

程序需要使用**长选项** `--config_path `获取配置文件路径的输入

命令行输入实例

1. 协同者进程：`./kvstore2pcsystem --config_path ./src/coordinator.conf`
2. 参与者进程：`./kvstore2pcsystem --config_path ./src/participant.conf`

### 配置文件内容格式

1. 注释行：` !`开头的单行文本，需要被程序忽略

2. 变量/参数行：

   1. mode：指定了进程的运行模式，是coordinator或者participant。mode行总是配置文件的参数行的第一行
   2. coordinator_info：协同者进程监听的`IP:Port`
   3. participant_info：参与者进程监听的`IP:Port`

   **备注：**所有进程的配置文件中都有coordinator_info，因为它们都需要知晓协同者的通讯地址。但participant_info则有所不同，在协同者进程的配置文件中包含所有参与者进程的通讯地址，而在参与者进程的配置文件中仅仅包含自身的participant_info。具体的例子可以看指导书。

## 建立连接

- 协同者结点-多个数据结点
- 客户端在需要操作数据库时，与协同者建立连接

## 为客户端提供服务

1. 客户端与协同者建立连接
2. 客户端向协同者传输指令
   1. 指令种类
      1. set(key, value): stores the value "value" with the key "key" (some 
         KV stores also call this command as "put").
      2. del(key): deletes any record associated with the key "key".
      3. value=get(key): retrieves and returns the value associated with 
         the key "key".
   2. 具体传输数据：RESP Arrays格式
3. 协同者监听并接收来自客户端的指令，解析出对应操作和数据
4. 协同者执行两阶段提交
   1. 准备：协同者-->数据结点
   2. 反馈：数据结点-->协同者
   3. 提交或者回滚：协同者-->数据结点
   4. 反馈：数据结点-->协同者
5. 协同者返回响应报文

