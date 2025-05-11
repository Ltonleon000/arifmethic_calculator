package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/Knetic/govaluate"
	"google.golang.org/grpc"
	"calculator/internal/calculatorpb"
)

// Сервер gRPC для вычислений
type calculatorServer struct {
	calculatorpb.UnimplementedCalculatorServiceServer
}

func (s *calculatorServer) Calculate(ctx context.Context, req *calculatorpb.CalculateRequest) (*calculatorpb.CalculateResponse, error) {
	log.Printf("Запрос: %s", req.Expression)
	
	result, err := evaluateExpression(req.Expression)
	if err != nil {
		log.Printf("Ошибка: %v", err)
		return nil, fmt.Errorf("ошибка вычисления: %v", err)
	}
	
	log.Printf("Результат: %f", result)
	return &calculatorpb.CalculateResponse{
		Result: fmt.Sprintf("%f", result),
	}, nil
}

// Вычисление выражения
func evaluateExpression(expr string) (float64, error) {
	expression, err := govaluate.NewEvaluableExpression(expr)
	if err != nil {
		return 0, fmt.Errorf("ошибка парсинга: %v", err)
	}
	
	result, err := expression.Evaluate(nil)
	if err != nil {
		return 0, fmt.Errorf("ошибка вычисления: %v", err)
	}
	
	switch v := result.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("неподдерживаемый тип: %T", result)
	}
}

func main() {
	log.Println("Запуск gRPC агента на порту 8082")
	
	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("Ошибка: %v", err)
	}
	
	grpcServer := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(grpcServer, &calculatorServer{})
	
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
