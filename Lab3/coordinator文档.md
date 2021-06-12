## 总体流程

入口为`start_coordinator`函数，首先根据配置文件中的信息，连接3个participant，连接完毕后首先检查3个par的一致性，若不一致则进行恢复。之后开启心跳协程，开始Accept用户的连接并处理。

## 用户操作的处理

函数名`clientHandle`，当前版本是单线程，每次只处理一个用户的请求。用户的请求信息符合RESP格式，通过工具函数解析后，塞到管道里，由心跳协程从管道中取数据进行具体的处理过程。处理结束后心跳协程把返回给用户的信息塞到另一个管道里，取出后返回给用户。

## 心跳协程

函数名```heartBeatsCheck```，使用了time.Ticker库，Ticker包含一个管道，每过一定的时间就会向管道中放一个资源，所以取Ticker管道的资源，如果时间没到会进行堵塞。之后会尝试从管道中取数据，即上文提到的用户操作command。如果没有数据，则简单发送心跳包。

> 此处补充：如何判断管道中有没有数据呢？可以加锁然后使用len来判断，但是有更加方便的方法：select。select和C++的switch相似，但是GO的select的各个case只能用于管道取资源的操作，如果所有case中存在可以取的管道，会通过GO的调度，较公平地选择分支。如果没有任何case可取，则会阻塞。可以添加default来让它不阻塞，所以代码可以写成：
>
> ```go
> select {
>     case cmd = <-c:
>     info = cmd2RESPArr(cmd)
>     fmt.Println("info=" + info)
>     default:
>     }
> ```

下面介绍一下通信协议：

1. 如果发送的是心跳包，大致过程为

    ```
    coo  heartbeat--->  par
    coo  <---------+OK  par
    ```

2. 如果是操作执行，大致过程为

    ```
    coo  command----->  par
    coo  <---------+OK  par		# 代表par准备就绪，但是不要操作
    |---当coo接收到所有存活的par发送的+OK后
    |	coo  commit----->  par
    |	coo  <-------info  par
    |
    |---否则
    	coo  rollback--->  par
    	coo  <--------+OK  par
    ```

    其中，info根据command的操作类型而定，具体内容可以参看participant函数。收到的info直接返回给用户即可。特别地，如果存活的par=0，由coordinator直接返回给用户，返回的info=`-ERROR`

当然，在心跳检测中可能会存在par挂掉的情况，此时连接断开，coordinator应该得知情况并且做出处理。

## 挂掉的par的处理

入口函数`eraseconn`，参数为挂掉的par的下标。每一次对par的连接read后，就要判断一次read的状态，如果失败就要调用eraseconn函数。大概的流程如下：

1. 当一个par挂掉时，eraseconn函数需要把对应的连接设置为nil表示连接断开，alive标识存活的par数量，isalive标识哪些par存活。

    ```GO
    connParticipant[p] = nil
    isalive[p] = false
    alive--
    ```

2. 之后这个par可能会复活，我们的处理方式是coordinator每过一段时间就尝试和挂掉的IP端口重新连接。具体地说，开启一个协程`try_dail`，传入挂掉的par下标p，每次休息1s，然后dail对应的par，不成功则继续循环。成功后，选择一个isalive=true的另一个存活的par，来给它恢复数据。

3. 恢复数据函数为`data_recover`，参数to表示刚刚恢复的par，p表示给to恢复数据的par。恢复协议为：

    ```
    coo  synData IP:PORT---->  to
    coo  <--------synData ACK  to
    coo  <--------synData FIN  to
    coo  synData FIN_ACK---->  to
    ```

    具体恢复的流程请参考participant实现。总之，coordinator给to的信息是p的IP和端口，p和to有另一个恢复协议。

4. 如果恢复失败，则继续调用`eraseconn`函数，把to挂了。否则，调用`recoverOK`，重置isalive的值。至此，恢复过程结束。

最后，实际上coordinator刚刚启动的时候就可能发生三个par不同步的问题，所以每次启动时就需要检查一致性。启动时的一致性检查函数为`init_participant`，具体协议是，请求3个participants的最后一次操作的id，认为拥有最大id的为最新版本，把不是最新版本的par和最新版本的par进行同步，还是使用`data_recover`函数。

