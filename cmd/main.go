package main

import (
	"fmt"
	"net"
	"os"

	pb "github.com/ranggadablues/relik-proto-library/user-proto/pb"

	"github.com/joho/godotenv"
	logger "github.com/ranggadablues/gosok/logger"
	"google.golang.org/grpc"

	"relik-user-service/internal/auth"
	"relik-user-service/internal/businesslogic"
	"relik-user-service/internal/endpoint"
)

func main() {
	log := logger.NewLogger()

	// Load .env into environment (optional; containers inject env vars directly)
	if err := godotenv.Load(".env"); err != nil {
		log.Error("msg", "No .env file found "+err.Error())
	}

	if err := auth.Init(); err != nil {
		log.Error("msg", "failed to init auth: "+err.Error())
		return
	}

	port := os.Getenv("SERVICE_PORT")
	userServicePort := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", userServicePort)
	if err != nil {
		log.Error("msg", "failed to listen: "+err.Error())
	}

	s := grpc.NewServer()

	// Create repository and inject into business logic
	// repo := repository.NewUserRepo()
	// bl := businesslogic.NewUserBusinessLogic(repo)
	bl := businesslogic.NewUserBusinessLogic()
	// Create gRPC server directly with business logic
	grpcServer := endpoint.NewUserServer(bl)
	// Create gRPC server and register service
	pb.RegisterUserServiceServer(s, grpcServer)

	blGoogle := businesslogic.NewGoogleBusinessLogic()
	grpcGoogleServer := endpoint.NewGoogleServer(blGoogle)
	pb.RegisterGoogleAuthServiceServer(s, grpcGoogleServer)

	blFacebook := businesslogic.NewFacebookBusinessLogic()
	grpcFacebookServer := endpoint.NewFacebookServer(blFacebook)
	pb.RegisterFacebookAuthServiceServer(s, grpcFacebookServer)

	log.Info("User service running", "addr", userServicePort)
	if err := s.Serve(lis); err != nil {
		log.Error("msg", "failed to serve: "+err.Error())
	}
}
