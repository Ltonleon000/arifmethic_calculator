package internal

import (
	"context"
	"log"
	"time"

	"calculator/calculatorpb"
	"google.golang.org/grpc"
)

type AgentGRPCClient struct {
	client calculatorpb.AgentServiceClient
	conn   *grpc.ClientConn
}

func NewAgentGRPCClient(addr string) (*AgentGRPCClient, error) {
	// Вместо устаревшего WithTimeout используем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Используем контекст для подключения
	conn, err := grpc.DialContext(
		ctx,
		addr, 
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	client := calculatorpb.NewAgentServiceClient(conn)
	return &AgentGRPCClient{client: client, conn: conn}, nil
}

func (c *AgentGRPCClient) Close() error {
	return c.conn.Close()
}

func (c *AgentGRPCClient) GetTask() (*calculatorpb.Task, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := c.client.GetTask(ctx, &calculatorpb.GetTaskRequest{})
	if err != nil {
		log.Printf("Ошибка gRPC GetTask: %v", err)
		return nil, false
	}
	if !resp.HasTask || resp.Task == nil {
		return nil, false
	}
	return resp.Task, true
}

func (c *AgentGRPCClient) SendResult(taskID int64, result float64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.client.SendResult(ctx, &calculatorpb.SendResultRequest{TaskId: taskID, Result: result})
	if err != nil {
		log.Printf("Ошибка gRPC SendResult: %v", err)
		return false
	}
	return true
}
