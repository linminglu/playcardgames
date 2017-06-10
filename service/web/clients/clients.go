package clients

import (
	"errors"
	"playcards/utils/log"
	"sync"
)

var waitGroup = new(sync.WaitGroup)

var clients = make(map[int32]map[*Client]bool)
var topics = make(map[string]map[*Client]bool)
var lock = new(sync.RWMutex)

func InitTopic(topic string) error {
	log.Info("service init topic: %v", topic)
	Lock()
	defer Unlock()

	_, ok := topics[topic]
	if ok {
		return errors.New("already init")
	}

	topics[topic] = make(map[*Client]bool)
	return nil
}

func Done() {
	RUlock()
}

func GetLockClients(topic string) map[*Client]bool {
	RLock()
	return topics[topic]
}

func GetClientByUserID(uid int32) []*Client {
	var out []*Client
	Lock()
	defer Unlock()

	val, ok := clients[uid]
	if !ok {
		return out
	}
	for v, _ := range val {
		out = append(out, v)
	}
	return out
}

func RLock() {
	lock.RLock()
}

func RUlock() {
	lock.RUnlock()
}

func Lock() {
	lock.Lock()
}

func Unlock() {
	lock.Unlock()
}

func Send(topic, typ string, msg interface{}) error {
	return SendWhere(topic, typ, msg, nil)
}

func SendTo(uid int32, topic, typ string, msg interface{}) error {
	return SendWhere(topic, typ, msg, func(c *Client) bool {
		return c.UserID() == uid
	})
}

func SendWhere(topic, typ string, msg interface{},
	f func(*Client) bool) error {
	cs := GetLockClients(topic)
	defer Done()

	for c, _ := range cs {
		if f != nil && !f(c) {
			continue
		}
		c.SendMessage(topic, typ, msg)
	}

	return nil
}

func DeleteClient(c *Client) {
	log.Debug("delete connection: %v", c)

	Lock()
	defer Unlock()

	uid := c.UserID()
	cs := clients[uid]

	delete(cs, c)
	if len(cs) == 0 {
		delete(clients, uid)
	}

	for t, _ := range c.topics {
		delete(topics[t], c)
	}
}

func CloseAll() {
	Lock()

	for _, cs := range clients {
		for c, _ := range cs {
			c.Close()
		}
	}

	Unlock()
	log.Info("all connections are closed: %v", waitGroup)
	waitGroup.Wait()
}
