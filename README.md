# grpc-stream-bench
grpc stream 模式吞吐量测试


## grpc-go stream mode best practice

1. 流模式下，客户端和服务器为每个流分配两个goroutine，一个用来发送数据，一个用来接收数据，共同组成一个可靠的数据传输通道。
2. stream.Recv不是线程安全的，多个goroutine调用会出现panic或者数据错乱
3. stream.Send可以多个goroutine并发调用，但并不会加快发送速度，单个goroutine发送已经可以将带宽打满
4. 如果接收端goroutine需要处理较重的逻辑，影响了接收速度，可以让客户端建立多个流，多通道并发进行数据传输
		
### client option

1. WithWriteBufferSize 发送缓冲区，默认32K，设置为0时每次Send都直接调用连接的Write方法
2. WithReadBufferSize 接收缓冲区，默认32K，设置为0时每次Recv都直接调用连接的Read方法
3. WithInitialWindowSize stream流控窗口大小，默认初始值64K，小于此值将被忽略，大于此值将关闭动态流控
4. WithInitialConnWindowSize 连接流控窗口大小，默认初始值64K，小于此值将被忽略，大于此值将关闭动态流控

### server option
1. WriteBufferSize 同client option
2. ReadBufferSize 同client option
3. InitialWindowSize 同client option
4. InitialConnWindowSize 同client option
5. MaxConcurrentStreams 每条连接可以支持的stream数量，达到上限后，客户端新建stream将被阻塞
6. NumStreamWorkers 实验性方法，设置一个work池来处理所有stream，为了解决为每个stream分配goroutine时morestack的性能影响；默认为0，关闭
			
### 动态流控 BDP

Bandwidth Delay Product (BDP), 即带宽延迟积

用于统计流量大小，并动态调整接收者的窗口大小

统计BDP ping frame和BDP ping frame ack之间的流量大小，作为当前“流量大小”，如解决当前窗口大小，则将窗口调整为BDP（采样流量）的两倍
			
