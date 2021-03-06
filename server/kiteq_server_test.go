package server

import (
	"github.com/golang/protobuf/proto"
	"kiteq/binding"
	"kiteq/client"
	"kiteq/protocol"
	"kiteq/store"
	"log"
	"testing"
	"time"
)

func buildStringMessage(id string) *protocol.StringMessage {
	//创建消息
	entity := &protocol.StringMessage{}
	entity.Header = &protocol.Header{
		MessageId:    proto.String(store.MessageId() + id),
		Topic:        proto.String("trade"),
		MessageType:  proto.String("pay-succ"),
		ExpiredTime:  proto.Int64(time.Now().Add(10 * time.Minute).Unix()),
		DeliverLimit: proto.Int32(-1),
		GroupId:      proto.String("go-kite-test"),
		Commit:       proto.Bool(true),
		Fly:          proto.Bool(false)}
	entity.Body = proto.String("hello go-kite")

	return entity
}

//初始化存储
var kitestore = &store.MockKiteStore{}
var ch = make(chan bool, 1)

var kiteClient *client.KiteQClient
var kiteQServer *KiteQServer
var c int32 = 0
var lc int32 = 0

type defualtListener struct {
}

func (self *defualtListener) OnMessage(msg *protocol.QMessage) bool {
	log.Printf("defualtListener|OnMessage|%s\n", msg.GetHeader().GetMessageId())
	return true
}

func (self *defualtListener) OnMessageCheck(tx *protocol.TxResponse) error {
	log.Printf("defualtListener|OnMessageCheck", tx.MessageId)
	tx.Commit()
	return nil
}

func init() {

	rc := protocol.NewRemotingConfig(
		"KiteQ-localhost:13800",
		1000, 16*1024,
		16*1024, 10000, 10000,
		10*time.Second, 160000)

	kc := NewKiteQConfig("localhost:13800", "localhost:2181", 1*time.Second, 10, 1*time.Minute, []string{"trade"}, "mmap://file=.", rc)

	kiteQServer = NewKiteQServer(kc)
	kiteQServer.Start()
	log.Println("KiteQServer START....")

	kiteClient = client.NewKiteQClient("localhost:2181", "s-trade-a", "123456", &defualtListener{})
	kiteClient.SetTopics([]string{"trade"})
	kiteClient.SetBindings([]*binding.Binding{
		binding.Bind_Direct("s-trade-a", "trade", "pay-succ", 1000, true),
	})
	kiteClient.Start()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			log.Printf("%d\n", (c - lc))
			lc = c
		}
	}()
}

func BenchmarkRemotingServer(t *testing.B) {
	t.SetParallelism(4)
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := kiteClient.SendStringMessage(buildStringMessage("1"))
			if nil != err {
				t.Logf("SEND MESSAGE |FAIL|%s\n", err)
			}
		}
	})
}

func TestRemotingServer(t *testing.T) {

	err := kiteClient.SendStringMessage(buildStringMessage("1"))
	if nil != err {
		t.Logf("SEND MESSAGE |FAIL|%s\n", err)
		t.Fail()

	}

	err = kiteClient.SendStringMessage(buildStringMessage("2"))
	if nil != err {
		t.Logf("SEND MESSAGE |FAIL|%s\n", err)
		t.Fail()

	}

	err = kiteClient.SendStringMessage(buildStringMessage("3"))
	if nil != err {
		t.Logf("SEND MESSAGE |FAIL|%s\n", err)
		t.Fail()

	}

	msg := buildStringMessage("4")
	msg.GetHeader().Commit = proto.Bool(false)

	err = kiteClient.SendStringMessage(msg)
	if nil != err {
		t.Logf("SEND MESSAGE |FAIL|%s\n", err)
		t.Fail()

	}

	time.Sleep(10 * time.Second)
}
