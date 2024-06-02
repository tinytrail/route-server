package main

import (
	"context"
	"flag"
	"fmt"

	pb "github.com/tinytrail/route-server/route_guide"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("addr", "localhost:1111", "The server address in the format of host:port")
)

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(*serverAddr, opts...)

	if err != nil {
		fmt.Println("Error: ", err)
	}
	defer conn.Close()
	client := pb.NewRouteGuideClient(conn)

	feature, err := client.GetFeature(context.Background(), &pb.Point{Latitude: 409146138, Longitude: -746188906})
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Feature: ", feature)

	feature, err = client.GetFeature(context.Background(), &pb.Point{Latitude: 0, Longitude: 0})
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Feature: ", feature)

}
