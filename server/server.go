package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	pb "github.com/tinytrail/route-server/route_guide"
)

var (
	port = flag.Int("port", 1111, "The server port")
	jsonFile = flag.String("json_file", "./server/route_guide_db.json", "A json file containing a list of features")
)

type routeGuideServer struct {
	pb.UnimplementedRouteGuideServer
	savedFeatures []*pb.Feature
	routeNotes map[string][]*pb.RouteNote
}

func (s *routeGuideServer) loadFeatures(filepath string) {
	// Load features from file
	fmt.Print("Loading features from file: ", filepath)
	var data []byte
	var err error
	data, err = os.ReadFile(filepath)
	if (err != nil) {
		fmt.Println("Error: ", err)
		data = []byte("{}")
	}
	fmt.Println("Data: ", string(data))
	if err := json.Unmarshal(data, &s.savedFeatures); err != nil {
		fmt.Println("Error: ", err)
	}
}

func (s *routeGuideServer) GetFeature(context context.Context, point *pb.Point) (*pb.Feature, error) {
	for _, feature := range s.savedFeatures {
		if feature.Location.Latitude == point.Latitude && feature.Location.Longitude == point.Longitude {
			return feature, nil
		}
	}
	// No feature was found, return an unnamed feature
	return &pb.Feature{Location: point, Name: "Unnamed"}, nil
}

func newServer() *routeGuideServer {
	s := routeGuideServer{routeNotes: make(map[string][]*pb.RouteNote)}
	s.loadFeatures(*jsonFile)
	return &s
}

func main() {
	flag.Parse()
	fmt.Print("Starting server on port:", *port)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if (err != nil) {
		fmt.Println("Error: ", err)
		return
	}

	var opts []grpc.ServerOption
	// Add more options from the cmdline flags to opts

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRouteGuideServer(grpcServer, newServer())
	grpcServer.Serve(lis);

}