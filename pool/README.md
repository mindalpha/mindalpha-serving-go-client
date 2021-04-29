# pool
[中文文档](README.CN.md) <br>

Connection pool

## Features
- consul based service discovery
- weighted Round-robin load balance
- automatically create connection when no connection availble

## Usage

```go

// consul and connection pool configuration
poolConfig := &pool.MindAlphaServingClientPoolConfig {
	MaxConnNumPerAddr:     30, // how many connections per service address
	MindAlphaServingService: "server_on_consul", // consul service the server registered on consul
	ConsulAddr: "127.0.0.1:8500", //consul address
}

// init, create connection pool
pool.InitMindAlphaServingClientPool(poolConfig)

// get connection pool
p, err := pool.GetConnPoolInstance()

// get a connection from connection pool
v, err := p.Get()
conn := v.(*pool.ConnWrap)

// send request on conn, record you requestId

// use the requestId as the param to call conn.PutRequest(), this will return a channel rspChan
rspChan := conn.PutRequest(reqeustId) 

// call p.Put() to put connection to connection pool
// param v is the return value of p.Get()
// param err indicates if the connection pool should close the connection v.
p.Put(v, err)

// get response data from rspChan
rspData, ok := <-rspChan

// process the response data

// when your process comes end, Release connection pool.
p.Release()

```

### Consul based service discovery
 MindAlpha-Serving server side will register itself to consul service after it started. In consul service, the ip:port of the MindAlpha-Serving service is contained. <br>
 This client read consul service information periodically (15 second). If there has any service instance down/up , this client will get the change in 15 second. <br>
 So, this client depends on consul, and needs to know the follow information: <br>
 1. consul address (for example, the local consul agent's address 127.0.0.1:8500)
 2. consul service name the MindAlpha-Serving service registered.
### Load balance
 Weighted Round-robin load balance. The algorithm refers [lvs](http://www.linuxvirtualserver.org/zh/lvs4.html) <br>
 The performance of different machine type is different. We configured different weight for MindAlpha-Serving service instance depends on its machine type it runs on. MindAlpha-Serving service instance will write its weight information to consul service when it starts. This client will get every service instance's ip:port and weight information, and use it as the weight of weighted Round-robin load balance.
 
#### Notice
 1. pool.InitMindAlphaServingClientPool() will create a global connection pool, this function should be called just once.
 2. Business go-routine calls pool.Get() to get a connection from pool, and this connection is owned by this go-routine exclusively. So businness code should call poolPut() to put this connection to pool as quickly as possible. If there's too many business go-routines and the business go-routine hold the connection too long time, it will lead to no useable connection in pool. Users should reasonably set MaxConnNumPerAddr.
 3. Business go-routine can only write data to connection gotten from pool, and must call PutRequest(reqestId) to record the requestId, and tihs function will return a channel, business go-routine should read response data from the channel.
