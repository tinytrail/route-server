package main

import (
	"context"
	"flag"
	"fmt"
	"time"
	"io"

	pb "github.com/tinytrail/route-server/route_guide"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("addr", "localhost:1111", "The server address in the format of host:port")
)

func getAndPrintFeature(client *pb.RouteGuideClient, point *pb.Point) {
	feature, err := (*client).GetFeature(context.Background(), point)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Feature: ", feature)

}

func listAndPrintFeatures(client *pb.RouteGuideClient, rectangle *pb.Rectangle) {
	// Provide a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := (*client).ListFeatures(ctx, rectangle)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	for {
		feature, err := stream.Recv()
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error: ", err)
			}
			break
		}
		fmt.Printf("Feature: %v with Location: %v, %v\n", feature.GetName(), feature.GetLocation().Latitude, feature.GetLocation().Longitude)
	}
}

func recordRoute(client *pb.RouteGuideClient) {

	points := []*pb.Point{
		{Latitude: 409146138, Longitude: -746188906},
		{Latitude: 0, Longitude: 0},
		{Latitude: 0, Longitude: 1},
		{Latitude: 1, Longitude: 0},
		{Latitude: 1, Longitude: 1},
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := (*client).RecordRoute(ctx)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	for _, point := range points {
		if err := stream.Send(point); err != nil {
			fmt.Println("Error: ", err)
		}
	}

	summary, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Route summary: ", summary)

}

func routeChat(client *pb.RouteGuideClient) {

	points := []*pb.RouteNote{
		{Location: &pb.Point{Latitude: 409146138, Longitude: -746188906}, Message: "First message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 0}, Message: "Second message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 1}, Message: "Third message"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := (*client).RouteChat(ctx)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	waitc := make(chan struct{})
	go func() {
		for {
			note, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				fmt.Println("Error: ", err)
				return
			}
			fmt.Printf("Got message: %v\n", note)
		}
	
		}()

		for _, point := range points {
			if err := stream.Send(point); err != nil {
				fmt.Println("Error: ", err)
			}
		}
		stream.CloseSend()
		<-waitc
}
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
	point := pb.Point{Latitude: 409146138, Longitude: -746188906}
	getAndPrintFeature(&client, &point)

	point = pb.Point{Latitude: 0, Longitude: 0}
	getAndPrintFeature(&client, &point)


	rect := pb.Rectangle{Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000}, Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000}}
	listAndPrintFeatures(&client, &rect)
	
	recordRoute(&client)

	routeChat(&client)
}

