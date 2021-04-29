package pool

import (
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"sync"
)

var (
	_once sync.Once
)

func InitMindAlphaServingClientPool(config *MindAlphaServingClientPoolConfig) {
	_once.Do(func() {
		_, err := NewChannelPool(config)
		if err != nil {
			client_logger.GetMindAlphaServingClientLogger().Errorf("InitMindAlphaServingClientPool().NewChannelPool() failed, error: %v", err)
			panic("call NewChannelPool() failed :" + err.Error())
		}
	})
}

func GetConnPoolInstance() (Pool, error) {
	return GetChannelPoolInstance()
}
