package internal

import (
	"context"
	"log"
	"net"

	"calculator/calculatorpb"
	"google.golang.org/grpc"
)

type AgentServiceServerImpl struct {
	calculatorpb.UnimplementedAgentServiceServer
	TaskProvider func() (*calculatorpb.Task, bool)
	ResultHandler func(taskID int64, result float64) error
}

func (s *AgentServiceServerImpl) GetTask(ctx context.Context, req *calculatorpb.GetTaskRequest) (*calculatorpb.GetTaskResponse, error) {
	task, ok := s.TaskProvider()
	if !ok || task == nil {
		return &calculatorpb.GetTaskResponse{HasTask: false}, nil
	}
	return &calculatorpb.GetTaskResponse{HasTask: true, Task: task}, nil
}

func (s *AgentServiceServerImpl) SendResult(ctx context.Context, req *calculatorpb.SendResultRequest) (*calculatorpb.SendResultResponse, error) {
	err := s.ResultHandler(req.TaskId, req.Result)
	if err != nil {
		return &calculatorpb.SendResultResponse{Ok: false, Error: err.Error()}, nil
	}
	return &calculatorpb.SendResultResponse{Ok: true}, nil
}

func StartGRPCServer(taskProvider func() (*calculatorpb.Task, bool), resultHandler func(taskID int64, result float64) error, port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	srv := &AgentServiceServerImpl{TaskProvider: taskProvider, ResultHandler: resultHandler}
	calculatorpb.RegisterAgentServiceServer(grpcServer, srv)
	log.Printf("gRPC сервер запущен на порту %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
