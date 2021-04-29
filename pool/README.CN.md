# pool
[English Document](README.md) <br>

连接池

## 功能：

- 利用consul实现服务的自动发现，周期性检查服务列表是否发生变化
- 利用加权轮询机制实现负载均衡
- 根据连接上发生的错误决定是否关掉一个连接
- 没有可用连接时，创建新的连接

## 基本用法

```go

//创建一个连接池配置
poolConfig := &pool.MindAlphaServingClientPoolConfig {
	MaxConnNumPerAddr:     30, //对于每一个MindAlpha-Serving服务端地址，最多创建多少个tcp连接.
	MindAlphaServingService: "server_on_consul", //MindAlpha-Serving服务端在consul上注册的service.
	ConsulAddr: "127.0.0.1:8500", //consul 连接地址.
}

//初始化、创建连接池
pool.InitMindAlphaServingClientPool(poolConfig)

//获取连接池
p, err := pool.GetConnPoolInstance()

//从连接池中取得一个连接
v, err := p.Get()

conn := v.(*pool.ConnWrap)
// send data on conn

//记录下发送的消息的requestId, 返回一个channel rspChan.
rspChan := conn.PutRequest(reqeustId) 

//将连接放回连接池中. err 指示是否需要关掉该连接. 如果是超时错误或者网络错误，则连接池内部会关闭该连接.
p.Put(v, err)

//从 rspChan 中读取MindAlpha-Serving服务端的返回数据
rspData, ok := <-rspChan

//处理服务端返回的消息.


//程序结束时释放连接池中的所有连接
p.Release()

```

### 基于consul的服务发现
 MindAlpha-Serving服务端启动后会在consul上注册service，service信息中包含服务端的ip:port;<br>
 该客户端会周期性(默认15秒)从consul上读取service 信息获得所有服务端的ip:port 列表； 如果有服务端down/up, 该客户端也能在15秒内感知到服务端的变化.<br>
 所以, 该客户端需要依赖consul, 并需要知道如下信息: <br>
 1. consul的连接地址(比如consul agent的连接地址127.0.0.1:8500)
 2. 服务端在consul上注册的service name
### 负载均衡
 使用加权轮询算法实现负载均衡。加权轮询算法参考lvs实现 http://www.linuxvirtualserver.org/zh/lvs4.html <br>
 不同的机器配置其性能是有差异的。我们为不同的服务实例根据机器类型的不同配置了不同的权重。服务实例在启动时会将自己的权重写入到service信息中, 客户端可以从consul获取到该服务实例的ip:port以及权重, 并根据该权重作为加权轮询的权重. 
 
#### 注意事项
 1. pool.InitMindAlphaServingClientPool() 创建的连接池是全局的，该函数只需要调用一次.
 2. 业务协程从pool中 Get()一个连接后, 该协程是独占该连接的, 所以Get()连接后应该尽快做完自己的事情然后将该连接Put()回连接池。如果协程数较多、Get()连接后占用连接时间较长，则会导致连接池中没有可用的连接。用户需要合理设置 MaxConnNumPerAddr 的值，该值如果设置的太大，则可能会导致连接数过多超过系统限制;如果该值设置的太小而用户处理协程较多且独占连接的时间较长，则可能会导致连接池中的连接不够用.
 3. 业务协程从pool中 Get()的连接，只能向该连接写数据，写完数据后需调用PutRequest(reqId),该调用会返回一个channel， 业务协程需要从channel中读取服务端返回的数据.
