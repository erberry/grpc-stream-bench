package main

import (
	"log"
	"net"

	"github.com/erberry/grpc-stream-bench/internal"
	"google.golang.org/grpc"
)

type server struct {
}

func (s *server) MsgHandler(stream internal.GateRoute_MsgHandlerServer) error {

	sendchan := make(chan *internal.GateMsg, 1000)

	go func() {
		for msg := range sendchan {
			if msg == nil {
				break
			}
			internal.PromeUsageServerSendChan.Dec()
			stream.Send(msg)
			internal.PromeHistogramServerSend.Inc()
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Println(err)
			close(sendchan)
			return err
		}
		internal.PromeHistogramServerRecv.Inc()
		internal.PromeUsageServerSendChan.Inc()
		sendchan <- msg
	}
}

func main() {
	internal.InitPrometheus()
	go internal.Serve()

	lis, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatalln(err)
	}

	s := grpc.NewServer(
		// grpc.WriteBufferSize(32*2*1024),
		// grpc.ReadBufferSize(32*2*1024),
		// grpc.InitialWindowSize(64*2*1024),
		// grpc.InitialConnWindowSize(64*2*1024),
		grpc.MaxConcurrentStreams(2),
	)
	internal.RegisterGateRouteServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}
