package tk

type broker struct {
	topics map[string]chan interface{} // key： topic  value ： queue
}

var bro = &broker{
	topics: make(map[string]chan interface{}),
}

func (b *broker) publish(topic string, msg interface{}) {
	tpc, ok := b.topics[topic]
	if !ok {
		b.topics[topic] = make(chan interface{}, 10000)
		tpc, _ = b.topics[topic]
	}
	if msg != nil {
		tpc <- msg
	}
}

func (b *broker) subscribe(topic string) chan interface{} {
	if tpc, ok := b.topics[topic]; ok {
		return tpc
	} else {
		return nil
	}
}

func (b *broker) close(topic string) {
	if tpc, ok := b.topics[topic]; ok {
		close(tpc)
		delete(b.topics, topic)
	}
}

type MQClient struct {
	brk *broker
}

func NewMQClient() *MQClient {
	return &MQClient{
		brk: &broker{
			topics: make(map[string]chan interface{}),
		},
	}
}

func (c *MQClient) Publish(topic string, msg interface{}) {
	c.brk.publish(topic, msg)
}

func (c *MQClient) Close(topic string) {
	c.brk.close(topic)
}

func (c *MQClient) Subscribe(topic string) chan interface{} {
	return c.brk.subscribe(topic)
}

// publish topic with default mq
func Pub(topic string, msg interface{}) {
	bro.publish(topic, msg)
}

// subcribe topic from default mq
func Sub(topic string) chan interface{} {
	return bro.subscribe(topic)
}

// close topic of default mq
func Clo(topic string) {
	bro.close(topic)
}
