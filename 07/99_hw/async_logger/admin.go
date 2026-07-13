package main

import (
	"time"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type Admin struct {
	UnimplementedAdminServer
	queues *ClientQueues
}

func NewAdmin(queues *ClientQueues) Admin {
	return Admin{
		UnimplementedAdminServer: UnimplementedAdminServer{},
		queues:                   queues,
	}
}

func (a Admin) Logging(nothing *Nothing, inStream Admin_LoggingServer) error {
	id, ch := a.queues.Connect()
	defer a.queues.Disconnect(id)

	for {
		select {
		case e := <-ch:
			err := inStream.Send(e)
			if err != nil {
				return status.Error(codes.Internal, "log error")
			}
		case <-inStream.Context().Done():
			return nil
		}
	}
}

func (a Admin) Statistics(interval *StatInterval, inStream grpc.ServerStreamingServer[Stat]) error {
	id, ch := a.queues.Connect()
	defer a.queues.Disconnect(id)

	s := Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}

	ticker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)
	for {
		if s.ByMethod == nil {
			s.ByMethod = make(map[string]uint64)
		}
		if s.ByConsumer == nil {
			s.ByConsumer = make(map[string]uint64)
		}

		select {
		case e := <-ch:
			s.Timestamp = e.Timestamp
			s.ByMethod[e.Method]++
			s.ByConsumer[e.Consumer]++
		case <-ticker.C:
			err := inStream.Send(&s)
			if err != nil {
				return status.Error(codes.Internal, "statistics error")
			}
			s.Reset()
		case <-inStream.Context().Done():
			ticker.Stop()
			return nil
		}
	}
}
