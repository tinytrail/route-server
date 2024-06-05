package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	pb "github.com/tinytrail/route-server/route_guide"
)

var (
	port     = flag.Int("port", 1111, "The server port")
	jsonFile = flag.String("json_file", "./server/route_guide_db.json", "A json file containing a list of features")
)

type routeGuideServer struct {
	pb.UnimplementedRouteGuideServer
	savedFeatures []*pb.Feature
	routeNotes    map[string][]*pb.RouteNote
}

func (s *routeGuideServer) loadFeatures(filepath string) {
	// Load features from file
	fmt.Print("Loading features from file: ", filepath)
	var data []byte
	var err error
	data, err = os.ReadFile(filepath)
	if err != nil {
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

func (s *routeGuideServer) ListFeatures(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	for _, feature := range s.savedFeatures {
		if inRange(feature.Location, rect) {
			stream.Send(feature)
		}
	}
	//stream.Context().Done()
	return nil
}

func (s *routeGuideServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	var pointCount, featureCount, distance int32
	var lastPoint *pb.Point
	startTime := time.Now()
	for {
		point, err := stream.Recv()

		// If the err is EOF, we have reached the end of the stream
		if err == io.EOF {
			// End the stream and construct the statistics
			currentTime := time.Now()
			return stream.SendAndClose(&pb.RouteSummary{
				FeatureCount: featureCount,
				PointCount:   pointCount,
				Distance:     distance,
				ElapsedTime:  int32(currentTime.Sub(startTime).Seconds())})
		}

		// For unknown errors, break the loop and quit
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}

		pointCount++
		// Compare the point to an existing feature and add to the count if they match
		for _, feature := range s.savedFeatures {
			if proto.Equal(feature.Location, point) {
				featureCount++
			}
		}
		if lastPoint != nil {
			distance += int32(math.Sqrt(math.Pow(float64(point.Latitude-lastPoint.Latitude), 2) + math.Pow(float64(point.Longitude-lastPoint.Longitude), 2)))
		}

		lastPoint = point
	}
	return nil
}

func (s *routeGuideServer) RouteChat(stream pb.RouteGuide_RouteChatServer) error {
	for {
		note, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			fmt.Println("Error: ", err)
			return err
		}

		// Construct a key for the incoming note
		key := fmt.Sprintf("%d %d", note.Location.Latitude, note.Location.Longitude)

		// Append the note to the list of notes for the key
		s.routeNotes[key] = append(s.routeNotes[key], note)

		// Send the note back to the client
		for _, n := range s.routeNotes[key] {
			err := stream.Send(n)
			if err != nil {
				return err
			}
		}

	}
}

func inRange(point *pb.Point, rect *pb.Rectangle) bool {
	left := math.Min(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	right := math.Max(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	top := math.Max(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))
	bottom := math.Min(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))

	return left <= float64(point.Longitude) && float64(point.Longitude) <= right &&
		bottom <= float64(point.Latitude) && float64(point.Latitude) <= top
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
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	var opts []grpc.ServerOption
	// Add more options from the cmdline flags to opts

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRouteGuideServer(grpcServer, newServer())
	grpcServer.Serve(lis)

}
