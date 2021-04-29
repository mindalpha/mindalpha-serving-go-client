package pool

import (
	"bytes"
	"encoding/binary"
	"errors"
	consulApi "github.com/hashicorp/consul/api"
	net_error "github.com/mindalpha/mindalpha-serving-go-client/error"
	msg_hdr "github.com/mindalpha/mindalpha-serving-go-client/message_header"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"github.com/mindalpha/mindalpha-serving-go-client/util"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

// MindAlphaServingClientPoolConfig config of connection pool.
type MindAlphaServingClientPoolConfig struct {
	MaxConnNumPerAddr int    // max tcp connections per MindAlpha-Serving service address.
	MindAlphaServingService        string // consul service name of MindAlpha-Serving server.
	ConsulAddr        string // consul address. the address maybe the local consul agent address(such as 127.0.1:8500), or the consul cluster address.
}

// one connection with address.
type ConnWrap struct {
	net.Conn
	addr        string
	shouldClose bool
	reqChanMap  map[uint64]chan []byte
	mu          sync.RWMutex // protect reqChanMap.
}

func (cw *ConnWrap) CloseConnWrap() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	return cw.closeConnWrap()
}
func (cw *ConnWrap) closeConnWrap() error {
	client_logger.GetMindAlphaServingClientLogger().Errorf("closeConnWrap(), addr: %v, num of request on this closing connection: %v", cw.addr, len(cw.reqChanMap))

	for _, ch := range cw.reqChanMap {
		close(ch)
	}
	cw.shouldClose = true
	cw.reqChanMap = make(map[uint64]chan []byte)

	cw.Close()
	return nil
}

func (cw *ConnWrap) PutRequest(reqId uint64) <-chan []byte {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	ch := make(chan []byte, 1)
	cw.reqChanMap[reqId] = ch
	return ch
}

//this struct related to one connection with time.
type idleConn struct {
	connWrap *ConnWrap
	t        time.Time
}

type connReq struct {
	idleConn *idleConn
}

//one service addr related to one serviceConns,which has many connections.
type serviceConns struct {
	idleConns      chan *idleConn
	openingConnNum int
}

type addrWeight struct {
	addr   string
	weight int
}

type addrWeightSlice []*addrWeight

func (this addrWeightSlice) toString() string {
	buf := bytes.NewBufferString("[ ")
	for _, v := range this {
		buf.WriteString(v.addr)
		buf.WriteString(" ")
		buf.WriteString(strconv.FormatInt(int64(v.weight), 10))
		buf.WriteString(", ")
	}
	buf.WriteString("]")
	return buf.String()
}

func (sl addrWeightSlice) Len() int {
	return len(sl)
}
func (sl addrWeightSlice) Swap(i, j int) {
	sl[i], sl[j] = sl[j], sl[i]
}
func (sl addrWeightSlice) Less(i, j int) bool {
	return sl[j].weight < sl[i].weight
}

type addrWeightsInfo struct {
	addrWeights addrWeightSlice
	weights     []int
	minWeight   int
	maxWeight   int
	gcd         int // Greatest Common Divisor.
}

// channelPool store connection info.
type channelPool struct {
	mu           sync.RWMutex
	servConnsMap map[string]*serviceConns
	servAddrList addrWeightSlice
	curAddrIdx   int
	maxActive    int
	curWeight    int
	maxWeight    int
	minWeight    int
	gcd          int // Greatest Common Divisor.
}

var (
	_poolConfig  *MindAlphaServingClientPoolConfig = nil
	_channelPool *channelPool         = nil
)

func getConsulServiceAddrs() (*addrWeightsInfo, error) {
	const defaultBalanceFactor int = 200
	cfg := consulApi.DefaultConfig()
	cfg.Address = _poolConfig.ConsulAddr
	consulClient, err := consulApi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	//services, _, err := consulClient.Health().Service(_poolConfig.MindAlphaServingService, "", true, nil)
	services, _, err := consulClient.Health().Service(_poolConfig.MindAlphaServingService, "", true, &consulApi.QueryOptions{
		AllowStale: true})
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("api.Client.Health().Service() failed, err: %v", err)
		return nil, err
	}
	addrMap := make(map[string]*addrWeight)
	for _, service := range services {
		address := net.JoinHostPort(service.Service.Address, strconv.Itoa(service.Service.Port))
		if _, ok := addrMap[address]; ok {
			client_logger.GetMindAlphaServingClientLogger().Errorf("duplicate address: %v", address)
			continue
		}
		//client_logger.GetMindAlphaServingClientLogger().Debugf("getConsulServiceAddrs(): address: %v", address)

		var balanceFactor int = 0
		if balanceFactorStr, ok := service.Service.Meta["balanceFactor"]; ok {
			if i, err := strconv.Atoi(balanceFactorStr); err == nil {
				balanceFactor = i
				if balanceFactor <= 0 {
					client_logger.GetMindAlphaServingClientLogger().Errorf("address: %v, balanceFactor invalid: %v:%v, use default : %v", address, balanceFactorStr, i, defaultBalanceFactor)
					balanceFactor = defaultBalanceFactor
				} else {
					//client_logger.GetMindAlphaServingClientLogger().Infof("address: %v, balanceFactor: %v:%v", address, balanceFactorStr, i)
				}
			}
		} else {
			client_logger.GetMindAlphaServingClientLogger().Errorf("no balanceFactor for %v, use default: %v", address, defaultBalanceFactor)
			balanceFactor = defaultBalanceFactor
		}
		addrMap[address] = &addrWeight{addr: address, weight: balanceFactor}

	}
	if len(addrMap) == 0 {
		return nil, errors.New("getConsulServiceAddrs(): no service address")
	}

	addrList := make([]*addrWeight, 0, len(addrMap))
	for _, v := range addrMap {
		addrList = append(addrList, v)
	}
	sort.Sort(addrWeightSlice(addrList))

	client_logger.GetMindAlphaServingClientLogger().Infof("getConsulServiceAddrs(): weighted addrList: %v", addrWeightSlice(addrList).toString())
	//addrWeightSlice(addrList).debugPrint()

	var awInfo addrWeightsInfo
	awInfo.addrWeights = addrWeightSlice(addrList)
	awInfo.maxWeight = addrList[0].weight
	awInfo.minWeight = addrList[len(addrList)-1].weight
	for _, v := range addrList {
		lenOfWeights := len(awInfo.weights)
		if lenOfWeights == 0 {
			awInfo.weights = append(awInfo.weights, v.weight)
		} else {
			if v.weight == awInfo.weights[lenOfWeights-1] {
				continue
			} else {
				awInfo.weights = append(awInfo.weights, v.weight)
			}
		}
	}
	if 1 == awInfo.minWeight {
		awInfo.gcd = 1
	} else {
		awInfo.gcd = util.GCD(awInfo.weights)
	}

	return &awInfo, nil
}

func readResponseData(connWrap *ConnWrap, deadline time.Time) (uint64, []byte, error) {
	var messageHeader msg_hdr.MessageHeader

	connWrap.SetReadDeadline(deadline)

	rspHeadBuf, err := msg_hdr.ReadExactly(*connWrap, int(msg_hdr.KMessageHeaderLen+msg_hdr.KRequestMetaLen))
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("readResponseData(): read response header failed: %v", err)
		return 0, nil, err
	}

	hdrBytesBuf := bytes.NewBuffer(rspHeadBuf)
	err = binary.Read(hdrBytesBuf, binary.LittleEndian, &messageHeader)
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("readResponseData(): binary.Read(messageHeader) failed: %v", err)
		return 0, nil, err
	}

	if messageHeader.ProcessorClassId != msg_hdr.KResponseProcessor_ClassId {
		client_logger.GetMindAlphaServingClientLogger().Errorf("readResponseData(): error response processor class id: %v, expect: %v", messageHeader.ProcessorClassId, msg_hdr.KResponseProcessor_ClassId)
		return 0, nil, errors.New("readResponseData(): error messageHeader.ProcessorClassId")
	}
	rspDataLen := messageHeader.DataBufferSize
	rspData, err := msg_hdr.ReadExactly(connWrap, int(rspDataLen))
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("readResponseData(): read response body failed: %v", err)
		return 0, nil, err
	}
	return messageHeader.RequestId, rspData, nil
}

func createConnection(service string) (*idleConn, error) {
	conn, err := net.DialTimeout("tcp", service, time.Millisecond*5)
	if err != nil {
		//client_logger.GetMindAlphaServingClientLogger().Errorf("createConnection(): failed: %v", err)
		return nil, err
	}
	cw := &ConnWrap{Conn: conn, addr: service, shouldClose: false, reqChanMap: make(map[uint64]chan []byte)}

	// create a goroutine for a connection. This goroutine read data on this connection returned from server side,
	// decode MessageHeader, get requestId from MessageHeader, then get the chann related to the requestId, send
	// the response data to chann.
	go func(cw *ConnWrap) {
		for {
			reqNum := len(cw.reqChanMap)
			deadline := time.Time{}
			if reqNum > 0 {
				deadline = time.Now().Add(time.Second * 1)
			}
			reqId, rspData, err := readResponseData(cw, deadline)
			if err != nil {
				client_logger.GetMindAlphaServingClientLogger().Errorf("connection routine: %v request on connection %v, shouldClose: %v, error:  %v", reqNum, cw.addr, cw.shouldClose, err)
				cw.CloseConnWrap()
				break
			} else {
				cw.mu.Lock()
				if ch, ok := cw.reqChanMap[reqId]; ok {
					//client_logger.GetMindAlphaServingClientLogger().Errorf("connection routine: rspData len: %v, reqId: %v", len(rspData), reqId)
					ch <- rspData
					close(ch)
					delete(cw.reqChanMap, reqId)
				} else {
					client_logger.GetMindAlphaServingClientLogger().Errorf("connection routine: not find reqId: %v", reqId)
				}
				cw.mu.Unlock()
			}
		}
		client_logger.GetMindAlphaServingClientLogger().Errorf("connection routine exit")
	}(cw)
	return &idleConn{connWrap: cw, t: time.Now()}, nil
}

func createConnsForService(addr string, connNum int) ([]*idleConn, error) {
	var conns []*idleConn
	for i := 0; i < connNum; i++ {
		conn, err := createConnection(addr)
		if err != nil {
			client_logger.GetMindAlphaServingClientLogger().Errorf("createConnsForService(): create %v'th connection to %v failed: %v", i, addr, err)
		} else {
			conns = append(conns, conn)
		}
	}
	if len(conns) == 0 {
		return conns, errors.New("can not create connection for " + addr)
	}
	return conns, nil
}

func GetChannelPoolInstance() (Pool, error) {
	return _channelPool, nil
}

// NewChannelPool initialize/config the connection pool.
func NewChannelPool(poolConfig *MindAlphaServingClientPoolConfig) (Pool, error) {
	_poolConfig = poolConfig
	client_logger.GetMindAlphaServingClientLogger().Infof("NewChannelPool(): config: addr: %v, service: %v", poolConfig.ConsulAddr, poolConfig.MindAlphaServingService)
	if 0 == poolConfig.MaxConnNumPerAddr {
		return nil, errors.New("invalid capacity settings")
	}

	c := &channelPool{
		servConnsMap: make(map[string]*serviceConns),
		curAddrIdx:   -1,
		maxActive:    poolConfig.MaxConnNumPerAddr,
		minWeight:    1,
		gcd:          1,
	}
	_channelPool = c

	var total_conn_num = 0
	awInfo, err := getConsulServiceAddrs()
	if err != nil || nil == awInfo || len(awInfo.addrWeights) == 0 {
		client_logger.GetMindAlphaServingClientLogger().Errorf("can not get any service address from consul: %v", err)
		return nil, errors.New("can not get any service address from consul: " + err.Error())
	}
	addrWeights := awInfo.addrWeights

	c.servAddrList = addrWeights
	c.maxWeight = awInfo.maxWeight
	c.minWeight = awInfo.minWeight
	c.gcd = awInfo.gcd

	for _, addrWeight := range addrWeights {
		addr := addrWeight.addr
		c.servConnsMap[addr] = &serviceConns{idleConns: make(chan *idleConn, poolConfig.MaxConnNumPerAddr), openingConnNum: 0}

		ics, err := createConnsForService(addr, poolConfig.MaxConnNumPerAddr)
		if err != nil {
			client_logger.GetMindAlphaServingClientLogger().Errorf("can not create connnection for %v. error: %v", addr, err)
		}
		for _, ic := range ics {
			c.servConnsMap[addr].idleConns <- ic
			c.servConnsMap[addr].openingConnNum += 1
			total_conn_num += 1
		}
		client_logger.GetMindAlphaServingClientLogger().Infof("set addr %v openingConnNum = %v", addr, c.servConnsMap[addr].openingConnNum)
	}
	if 0 == total_conn_num {
		client_logger.GetMindAlphaServingClientLogger().Errorf("can not create any connection for any service address")
		return nil, errors.New("can not create any connection for any service address")
	}
	// watch consul, check address change
	go func(*channelPool) {
		for {
			time.Sleep(time.Second * 15)

			awInfo, err := getConsulServiceAddrs()
			if err != nil || nil == awInfo || len(awInfo.addrWeights) == 0 {
				client_logger.GetMindAlphaServingClientLogger().Errorf("Watch(): can not get any service address from consul: %v", err)
				continue
			}
			newAddrWeights := awInfo.addrWeights
			var addedAddrs []string
			var delAddrs []string
			newAddrsMap := make(map[string]int, len(newAddrWeights))
			for _, newAddrWeight := range newAddrWeights {
				newAddr := newAddrWeight.addr
				newAddrsMap[newAddr] = 1
				_, ok := c.servConnsMap[newAddr]
				if !ok {
					addedAddrs = append(addedAddrs, newAddr)
					client_logger.GetMindAlphaServingClientLogger().Errorf("Watch(): get new address %v from consul", newAddr)
				}
			}
			// check deleted address.
			for k, _ := range c.servConnsMap {
				if _, ok := newAddrsMap[k]; !ok {
					delAddrs = append(delAddrs, k)
				}
			}
			if len(delAddrs) > 0 {
				client_logger.GetMindAlphaServingClientLogger().Errorf("Watch(): deleted address: %v", delAddrs)
			}
			if len(addedAddrs) > 0 {
				client_logger.GetMindAlphaServingClientLogger().Errorf("Watch(): added  address: %v", addedAddrs)
			}

			addrIcsMap := make(map[string][]*idleConn)
			for _, addAddr := range addedAddrs {

				ics, err := createConnsForService(addAddr, poolConfig.MaxConnNumPerAddr)
				if err != nil {
					client_logger.GetMindAlphaServingClientLogger().Errorf("can not create connnection for %v. error: %v", addAddr, err)
				}
				total_conn_num := 0
				total_conn_num += len(ics)
				addrIcsMap[addAddr] = ics

				if 0 == total_conn_num {
					client_logger.GetMindAlphaServingClientLogger().Errorf("Watch(): can not create any connection for new added service %v", addAddr)
				}
			}
			c.mu.Lock()
			for addAddr, ics := range addrIcsMap {
				c.servConnsMap[addAddr] = &serviceConns{idleConns: make(chan *idleConn, poolConfig.MaxConnNumPerAddr), openingConnNum: 0}
				c.servConnsMap[addAddr].openingConnNum += len(ics)
				for _, ic := range ics {
					c.servConnsMap[addAddr].idleConns <- ic
				}
			}

			//delete deleted address related connections
			for _, delAddr := range delAddrs {
				close(c.servConnsMap[delAddr].idleConns)
				for ic := range c.servConnsMap[delAddr].idleConns {
					ic.connWrap.CloseConnWrap()
					c.servConnsMap[delAddr].openingConnNum--
				}
				delete(c.servConnsMap, delAddr)
			}
			c.servAddrList = newAddrWeights
			c.maxWeight = awInfo.maxWeight
			c.minWeight = awInfo.minWeight
			c.gcd = awInfo.gcd
			c.mu.Unlock()
		}
	}(c) // end of goroutine.

	return c, nil
}

// Get get a connection from connection pool.
//func (c *channelPool) Get() (*ConnWrap, error) {
func (c *channelPool) Get() (interface{}, error) {
	var outerLoopTimes uint64 = 0
	const maxOuterLoopTimes uint64 = 10
	for {
		if outerLoopTimes >= maxOuterLoopTimes {
			return nil, errors.New("Get(): too long time to get connection from pool. Looped times: " + strconv.FormatUint(outerLoopTimes, 10))
		}
		outerLoopTimes++
		c.mu.Lock()
		addrNum := len(c.servAddrList)

		for i := 0; i < addrNum; i++ {
			c.curAddrIdx = (c.curAddrIdx + 1) % addrNum
			if 0 == c.curAddrIdx {
				c.curWeight = c.curWeight - c.gcd
				if c.curWeight <= 0 {
					c.curWeight = c.maxWeight
				}
			}

			addrWeight := c.servAddrList[c.curAddrIdx%addrNum]
			addr := addrWeight.addr

			if addrWeight.weight < c.curWeight {
				// addrWeightSlice has already sorted, so we should not iterate anymore
				c.curAddrIdx = -1
				continue
			}
			servConns, ok := c.servConnsMap[addr]
			if !ok {
				//should not comes here
				client_logger.GetMindAlphaServingClientLogger().Errorf("error: servAddrList do not math servConnsMap. servAddrList: %v, servConnsMap: %v", c.servAddrList, len(c.servConnsMap))
				break
			}

			idleConnsChan := servConns.idleConns
			select {
			case ic := <-idleConnsChan:
				if ic.connWrap.shouldClose {
					ic.connWrap.CloseConnWrap()
					servConns.openingConnNum--
					client_logger.GetMindAlphaServingClientLogger().Errorf("Get(): connection on %v should close, set openingConnNum: %v", ic.connWrap.addr, servConns.openingConnNum)
				} else {
					c.mu.Unlock()
					return ic.connWrap, nil
				}
			default:
				if servConns.openingConnNum < c.maxActive { // create a new connection on addr.
					client_logger.GetMindAlphaServingClientLogger().Errorf("Get(): no available connection on address %v, openingConnNum: %v, less than %v, try create a new one", addr, servConns.openingConnNum, c.maxActive)
					ics, err := createConnsForService(addr, 1)
					if err != nil {
						//client_logger.GetMindAlphaServingClientLogger().Errorf("Get(): can not create connnection for %v. error: %v", addr, err)
					} else {
						servConns.openingConnNum += 1
						client_logger.GetMindAlphaServingClientLogger().Errorf("Get(): return new created connection on address %v", addr)
						c.mu.Unlock()
						return ics[0].connWrap, nil
					}
				} else {
					//client_logger.GetMindAlphaServingClientLogger().Errorf("Get(): no available connection on address %v, and openingConnNum exceed ", addr, c.maxActive)
				}

			} // end of select
		} // end of for().

		//unlock, no use to lock, or Put() / watch()  will block.
		c.mu.Unlock()

		if outerLoopTimes%5 == 0 {
			//client_logger.GetMindAlphaServingClientLogger().Errorf("Get(): all services has no available connection. We have looped %v times", outerLoopTimes)
		}
		time.Sleep(time.Millisecond * 1)
	} //end of for()
}

// Put put a connection to connection pool.
// connErr: indicates if we should close the connection. If timeout error or NetError happend, we should close the connection.
//func (c *channelPool) Put(connWrap *ConnWrap, connErr error) error {
func (c *channelPool) Put(connWrap interface{}, connErr error) error {

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.put(connWrap, connErr)
}

func (c *channelPool) put(connWrap interface{}, connErr error) error {
	if connWrap == nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("Put(): connWrap is nil. rejecting")
		return errors.New("connection is nil. rejecting")
	}

	if c.servConnsMap == nil {
		return connWrap.(*ConnWrap).CloseConnWrap()
	}

	// first check if the address has removed.
	scs, ok := c.servConnsMap[connWrap.(*ConnWrap).addr]
	if !ok {
		client_logger.GetMindAlphaServingClientLogger().Errorf("Put(): channlPool.serverConnsMap has no addr %v, maybe the address be deregistered on consul. close it", connWrap.(*ConnWrap).addr)
		connWrap.(*ConnWrap).CloseConnWrap()
		return nil
	}

	netErrorClose := false
	if connErr != nil {
		if _, ok := connErr.(*net_error.NetError); ok {
			netErrorClose = true
		}
	}

	if netErrorClose {
		connWrap.(*ConnWrap).CloseConnWrap()
		scs.openingConnNum--
		client_logger.GetMindAlphaServingClientLogger().Errorf("put(): net error on %v, set openingConnNum to %v, error: %v", connWrap.(*ConnWrap).addr, scs.openingConnNum, connErr.(*net_error.NetError))
		return nil
	}

	select {
	case scs.idleConns <- &idleConn{connWrap: connWrap.(*ConnWrap), t: time.Now()}:
		//client_logger.GetMindAlphaServingClientLogger().Errorf("Put(): put conn %v to poll", connWrap.(ConnWrap).addr)
		return nil
	default:
		client_logger.GetMindAlphaServingClientLogger().Errorf("Put(%v), connection pool already full. Error may happened, openingConnNum: %v", connWrap.(*ConnWrap).addr, scs.openingConnNum)

		connWrap.(*ConnWrap).CloseConnWrap()
		if scs.openingConnNum > c.maxActive {
			scs.openingConnNum = c.maxActive
		}
		return nil
	}
}

// Release release all connections
func (c *channelPool) Release() {
	c.mu.Lock()
	for _, servConn := range c.servConnsMap {
		for ic := range servConn.idleConns {
			ic.connWrap.CloseConnWrap()
		}
		close(servConn.idleConns)
		servConn.openingConnNum = 0
	}

	c.servConnsMap = nil
	c.servAddrList = nil

	c.mu.Unlock()
}
