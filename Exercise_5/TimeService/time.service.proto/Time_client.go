package time_service_proto

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"

	pb "time.service.proto/time_service.proto" // Import your generated protobuf package
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTimeServiceClient(conn)

	ctx := context.Background()
	request := &pb.GetCurrentTimeRequest{}

	response, err := client.GetCurrentTime(ctx, request)
	if err != nil {
		log.Fatalf("Error calling GetCurrentTime: %v", err)
	}

	fmt.Printf("Current Time: %s\n", response.GetCurrentTime())
}
