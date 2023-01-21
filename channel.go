package EIM

import (
	"EIM/logger"
	"errors"
	"sync"
	"time"
)

// ChannelImpl websocket的channel实现
type ChannelImpl struct {
	sync.Mutex
	id string
	Conn
	writechan chan []byte
	once      sync.Once
	writewait time.Duration
	readwait  time.Duration
	closed    *Event
}

// NewChannel 创建一个新channel
func NewChannel(id string, conn Conn) Channel {
	log := logger.WithFields(logger.Fields{
		"module": "tcp_channel",
		"id":     id,
	})

	ch := &ChannelImpl{
		id:        id,
		Conn:      conn,
		writechan: make(chan []byte, 5),
		writewait: time.Second * 10,
		closed:    NewEvent(),
	}

	go func() {
		err := ch.writeloop()
		if err != nil {
			log.Info(err)
		}
	}()

	return ch
}

// writeloop 循环读取writechan发送过来的payload
func (ch *ChannelImpl) writeloop() error {
	for {
		select {
		case payload := <-ch.writechan:
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}
			// 批量写
			chanlen := len(ch.writechan)
			for i := 0; i < chanlen; i++ {
				payload = <-ch.writechan
				err = ch.WriteFrame(OpBinary, payload)
				if err != nil {
					return err
				}
			}
			err = ch.Conn.Flush()
			if err != nil {
				return err
			}
		case <-ch.closed.Done():
			return nil
		}
	}
}

// ID 实现Agent接口
func (ch *ChannelImpl) ID() string {
	return ch.id
}

// Push 异步写数据
func (ch *ChannelImpl) Push(payload []byte) error {
	if ch.closed.HasFired() {
		return errors.New("channel has closed")
	}
	// 异步写
	ch.writechan <- payload
	return nil
}

// Close 关闭连接
func (ch *ChannelImpl) Close() error {
	ch.once.Do(func() {
		close(ch.writechan)
		ch.closed.Fire()
	})
	return nil
}

// SetWriteWait 设置写超时
func (ch *ChannelImpl) SetWriteWait(writewait time.Duration) {
	if writewait == 0 {
		return
	}
	ch.SetWriteWait(writewait)
}

// SetReadWait 设置读超时
func (ch *ChannelImpl) SetReadWait(readwait time.Duration) {
	if readwait == 0 {
		return
	}
	ch.SetReadWait(readwait)
}

// WriteFrame 重写Conn的WriteFrame方法(增加了重置写超时的逻辑)
func (ch *ChannelImpl) WriteFrame(code OpCode, payload []byte) error {
	_ = ch.Conn.SetWriteDeadline(time.Now().Add(ch.writewait))
	return ch.Conn.WriteFrame(code, payload)
}

// Readloop 循环读取消息
func (ch *ChannelImpl) Readloop(lst MessageListener) error {
	ch.Lock()
	defer ch.Unlock()
	log := logger.WithFields(logger.Fields{
		"struct": "ChannelImpl",
		"func":   "Readloop",
		"id":     ch.id,
	})

	for {
		_ = ch.SetReadDeadline(time.Now().Add(ch.readwait))

		frame, err := ch.ReadFrame()
		if err != nil {
			return err
		}
		if frame.GetOpCode() == OpClose {
			return errors.New("remote side close the channel")
		}
		if frame.GetOpCode() == OpPing {
			log.Trace("receive a ping; respond with a pong")
			_ = ch.WriteFrame(OpPong, nil)
			continue
		}
		payload := frame.GetPayload()
		if len(payload) == 0 {
			continue
		}
		// TODO: Optimization point
		go lst.Receive(ch, payload)
	}
}
