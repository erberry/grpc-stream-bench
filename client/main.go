package main

import (
	"context"
	"encoding/binary"
	"flag"
	"log"
	"sync"
	"time"

	"github.com/erberry/grpc-stream-bench/internal"
	"google.golang.org/grpc"
)

var (
	senddata   []byte
	ServerAddr = flag.String("server", "10.60.80.2:9999", "服务器地址：ip:port")
	MsgSize    = flag.Int("msgsize", 120, "发包数据大小")
	SendTick   = flag.Int("sendtick", 10, "每隔X毫秒进行一次发送")
	SendCnt    = flag.Int("sendcnt", 10, "每次的发包数量")
	buf        *sync.Pool
)

func init() {
	senddata = make([]byte, *MsgSize)
	for i := 0; i < len(senddata); i++ {
		senddata[i] = 6
	}

	buf = &sync.Pool{
		New: func() any {
			return make([]byte, *MsgSize+8)
		},
	}
}

func multStream() {

	streamcnt := 2 //runtime.NumCPU() * 2
	log.Println("num of streams:", streamcnt)
	log.Printf("interval: %d milliseconds, send cnt: %d, pkg size: %d \n", *SendTick, *SendCnt, *MsgSize+8)

	conn, err := grpc.Dial(
		*ServerAddr,
		grpc.WithInsecure(),
		// grpc.WithWriteBufferSize(32*2*1024),
		// grpc.WithReadBufferSize(32*2*1024),
		// grpc.WithInitialWindowSize(64*2*1024),
		// grpc.WithInitialConnWindowSize(64*2*1024),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := internal.NewGateRouteClient(conn)
	streams := make([]internal.GateRoute_MsgHandlerClient, streamcnt)
	for i := range streams {
		streams[i], err = client.MsgHandler(context.Background())
		if err != nil {
			log.Fatalf("failed to call MsgHandler: %v", err)
		}
	}

	totalcnt := streamcnt * (1000 / *SendTick) * (*SendCnt)
	log.Printf("total pkg cnt:%d, size:%d(byte) size:%d(bit)", totalcnt, totalcnt*(*MsgSize+8), totalcnt*(*MsgSize+8)*8)
	for i := 0; i < streamcnt; i++ {
		// Send
		go func(goindex int) {
			ticker := time.NewTicker(time.Duration(*SendTick) * time.Millisecond)
			for t := range ticker.C {
				for i := 0; i < *SendCnt; i++ {
					send := t.UnixMilli()
					data := buf.Get().([]byte)
					binary.BigEndian.PutUint64(data, uint64(send))
					copy(data[8:], senddata)
					if err := streams[goindex].Send(&internal.GateMsg{
						Cmdid: 1,
						Udid:  "my-device-id",
						Ucid:  "my-user-id",
						Data:  data,
					}); err != nil {
						log.Fatalf("failed to send message: %v", err)
					}
					buf.Put(data)
					success := time.Now().UnixMilli()
					internal.PromeHistogramClientSend.WithLabelValues().Observe(float64(success-send) / 1000)
				}
			}
		}(i)
	}

	for i := 0; i < streamcnt; i++ {
		go func(goindex int) {
			// Receive messages in a loop
			for {
				msg, err := streams[goindex].Recv()
				if err != nil {
					log.Fatalf("failed to receive message: %v", err)
				}

				send := binary.BigEndian.Uint64(msg.Data)
				recv := time.Now().UnixMilli()

				internal.PromeHistogramClientRecv.WithLabelValues().Observe(float64(uint64(recv)-send) / 1000)
			}
		}(i)
	}

	select {}
}

func main() {
	flag.Parse()

	internal.InitPrometheus()
	go internal.Serve()

	multStream()
}

// func multClient() {
// 	for i := 0; i < 3; i++ {
// 		conn, err := grpc.Dial(
// 			*ServerAddr,
// 			grpc.WithInsecure(),
// 			grpc.WithTimeout(3*time.Second),
// 			// grpc.WithWriteBufferSize(32*2*1024),
// 			// grpc.WithReadBufferSize(32*2*1024),
// 			// grpc.WithInitialWindowSize(64*2*1024),
// 			// grpc.WithInitialConnWindowSize(64*2*1024),
// 		)

// 		if err != nil {
// 			log.Fatalf("failed to connect: %v", err)
// 		}

// 		client := internal.NewGateRouteClient(conn)
// 		streams := make([]internal.GateRoute_MsgHandlerClient, 3)
// 		for i := range streams {
// 			streams[i], err = client.MsgHandler(context.Background())
// 			if err != nil {
// 				log.Fatalf("failed to call MsgHandler: %v", err)
// 			}
// 		}
// 	}

// 	log.Println("stream over")
// 	select {}
// }

// func multiGo() {

// 	goroutines := runtime.NumCPU() * 2
// 	log.Println("num of goroutines:", goroutines)
// 	log.Printf("interval: %d milliseconds, send cnt: %d, pkg size: %d \n", *SendTick, *SendCnt, *MsgSize+8)

// 	conn, err := grpc.Dial(
// 		*ServerAddr,
// 		grpc.WithInsecure(),
// 		// grpc.WithWriteBufferSize(32*2*1024),
// 		// grpc.WithReadBufferSize(32*2*1024),
// 		// grpc.WithInitialWindowSize(64*2*1024),
// 		// grpc.WithInitialConnWindowSize(64*2*1024),
// 	)
// 	if err != nil {
// 		log.Fatalf("failed to connect: %v", err)
// 	}
// 	defer conn.Close()

// 	client := internal.NewGateRouteClient(conn)
// 	stream, err := client.MsgHandler(context.Background())
// 	if err != nil {
// 		log.Fatalf("failed to call MsgHandler: %v", err)
// 	}

// 	totalcnt := goroutines * (1000 / *SendTick) * (*SendCnt)
// 	log.Printf("total pkg cnt:%d, size:%d(byte) size:%d(bit)", totalcnt, totalcnt*(*MsgSize+8), totalcnt*(*MsgSize+8)*8)
// 	for i := 0; i < goroutines; i++ {
// 		// Send
// 		go func(goindex int) {
// 			ticker := time.NewTicker(time.Duration(*SendTick) * time.Millisecond)
// 			for t := range ticker.C {
// 				for i := 0; i < *SendCnt; i++ {
// 					send := t.UnixMilli()
// 					data := buf.Get().([]byte)
// 					binary.BigEndian.PutUint64(data, uint64(send))
// 					copy(data[8:], senddata)
// 					if err := stream.Send(&internal.GateMsg{
// 						Cmdid: 1,
// 						Udid:  "my-device-id",
// 						Ucid:  "my-user-id",
// 						Data:  data,
// 					}); err != nil {
// 						log.Fatalf("failed to send message: %v", err)
// 					}
// 					buf.Put(data)
// 					success := time.Now().UnixMilli()
// 					internal.PromeHistogramClientSend.WithLabelValues().Observe(float64(success-send) / 1000)
// 				}
// 			}
// 		}(i)
// 	}

// 	//for i := 0; i < goroutines; i++ {
// 	go func(goindex int) {
// 		// Receive messages in a loop
// 		for {
// 			msg, err := stream.Recv()
// 			if err != nil {
// 				log.Fatalf("failed to receive message: %v", err)
// 			}

// 			send := binary.BigEndian.Uint64(msg.Data)
// 			recv := time.Now().UnixMilli()

// 			internal.PromeHistogramClientRecv.WithLabelValues().Observe(float64(uint64(recv)-send) / 1000)
// 		}
// 	}(0)
// 	//}

// 	select {}
// }
