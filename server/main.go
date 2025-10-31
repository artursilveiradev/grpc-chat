package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	pb "github.com/artursilveiradev/grpc-chat/server/pb"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Connection represents a single connected client.
// We use a channel to send messages to this client.
type Connection struct {
	stream pb.ChatService_ConnectServer
	user   string
	error  chan error
}

// ChatServer stores all active connections.
// We use a Mutex to protect concurrent access to the connections map.
type ChatServer struct {
	pb.UnimplementedChatServiceServer                        // Required for gRPC implementation
	connections                       map[string]*Connection // Map of active connections (User -> Connection)
	mutex                             sync.RWMutex           // Mutex to protect the map
}

// Connect is the main method called when a client connects.
func (s *ChatServer) Connect(stream pb.ChatService_ConnectServer) error {
	log.Println("New client attempting to connect...")

	// 1. Receive the first message to identify the user
	initialMsg, err := stream.Recv()
	if err != nil {
		log.Printf("Error receiving initial message: %v", err)
		return err
	}
	user := initialMsg.User
	log.Printf("Client '%s' connected.", user)

	// 2. Create the Connection struct for this client
	connection := &Connection{
		stream: stream,
		user:   user,
		error:  make(chan error),
	}

	// 3. Add the connection to the map (protected by Mutex)
	s.addConnection(user, connection)

	// 4. Announce to everyone that this user has joined
	joinMsg := &pb.ChatMessage{
		User:      "Server",
		Text:      fmt.Sprintf("%s joined the room.", user),
		Timestamp: timestamppb.Now(),
	}
	s.broadcast(joinMsg)

	// 5. Start a goroutine to receive messages from this client
	go s.receiveMessages(connection)

	// 6. Return the error channel to know when the client disconnects
	return <-connection.error
}

// addConnection adds a client to the connections map
func (s *ChatServer) addConnection(user string, connection *Connection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.connections[user] = connection
}

// removeConnection removes a client and announces their departure
func (s *ChatServer) removeConnection(connection *Connection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if the connection still exists (might have been removed by another goroutine)
	if _, ok := s.connections[connection.user]; !ok {
		return
	}

	delete(s.connections, connection.user)
	log.Printf("Client '%s' disconnected.", connection.user)

	// Announce to everyone that the user has left
	leaveMsg := &pb.ChatMessage{
		User:      "Server",
		Text:      fmt.Sprintf("%s left the room.", connection.user),
		Timestamp: timestamppb.Now(),
	}
	s.broadcast(leaveMsg)
}

// receiveMessages runs in a separate goroutine for each client.
// It listens for messages from this client and broadcasts them to everyone.
func (s *ChatServer) receiveMessages(connection *Connection) {
	for {
		msg, err := connection.stream.Recv()

		// If the client disconnects (io.EOF) or there's another error
		if err == io.EOF {
			s.removeConnection(connection)
			connection.error <- nil // Inform the main goroutine that this client left
			return
		}
		if err != nil {
			log.Printf("Error receiving from client %s: %v", connection.user, err)
			s.removeConnection(connection)
			connection.error <- err // Report the error
			return
		}

		// Add a server timestamp
		msg.Timestamp = timestamppb.Now()

		// Broadcast the message to everyone else
		log.Printf("Received from %s: %s", msg.User, msg.Text)
		s.broadcast(msg)
	}
}

// broadcast sends a message to ALL connected clients
func (s *ChatServer) broadcast(msg *pb.ChatMessage) {
	s.mutex.RLock() // RLock allows for multiple concurrent reads
	defer s.mutex.RUnlock()

	for user, connection := range s.connections {
		// Send the message to the client's stream
		if err := connection.stream.Send(msg); err != nil {
			log.Printf("Error sending to %s: %v. Removing connection.", user, err)
			// If sending fails, remove the connection
			// We use a goroutine to avoid deadlock (removeConnection uses Lock)
			go s.removeConnection(connection)
		}
	}
}

func main() {
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}
	log.Printf("gRPC Server listening on %s", port)

	// Create the gRPC server
	grpcServer := grpc.NewServer()

	// Instantiate our chat server
	chatServer := &ChatServer{
		connections: make(map[string]*Connection),
	}

	// Register the service with the gRPC server
	pb.RegisterChatServiceServer(grpcServer, chatServer)

	// Start the server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
