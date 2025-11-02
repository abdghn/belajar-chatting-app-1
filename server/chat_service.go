package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	pb "belajar-chatting-app-1/proto"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	mu      sync.Mutex
	clients map[string]pb.ChatService_StreamChatServer // key = client id
}

// Constructor
func NewChatServer() *ChatServer {
	return &ChatServer{
		clients: make(map[string]pb.ChatService_StreamChatServer),
	}
}

// Unary RPC (Postman / NestJS)
func (s *ChatServer) SendMessage(_ context.Context, msg *pb.ChatMessage) (*pb.SendMessageResponse, error) {
	msg.Timestamp = time.Now().Unix()

	log.Printf("[%s] %s: %s", msg.RoomId, msg.Sender, msg.Content)

	// ✅ Broadcast ke semua stream aktif
	s.broadcast(msg)

	return &pb.SendMessageResponse{
		Success: true,
		Message: "Message broadcasted successfully!",
	}, nil
}

// Streaming RPC (Bidirectional)
func (s *ChatServer) StreamChat(stream pb.ChatService_StreamChatServer) error {
	clientID := fmt.Sprintf("%p", stream)

	s.mu.Lock()
	s.clients[clientID] = stream
	s.mu.Unlock()

	fmt.Println("🟢 New stream connected:", clientID)

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientID)
		s.mu.Unlock()
		fmt.Println("🔴 Stream disconnected:", clientID)
	}()

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			fmt.Println("❌ Error receiving:", err)
			return err
		}

		msg.Timestamp = time.Now().Unix()

		log.Printf("[%s] %s: %s", msg.RoomId, msg.Sender, msg.Content)

		// ✅ Broadcast ke semua stream aktif (termasuk dari stream client lain)
		s.broadcast(msg)
	}
}

// Utility untuk broadcast ke semua client stream aktif
func (s *ChatServer) broadcast(msg *pb.ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, client := range s.clients {
		// kirim ke semua stream
		if err := client.Send(msg); err != nil {
			log.Printf("⚠️ Error kirim ke %s: %v", id, err)
		}
	}
}
