package server

import (
	"context"
	"fmt"
	_ "fmt"
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

// Unary (buat test di Postman)
func (s *ChatServer) SendMessage(_ context.Context, msg *pb.ChatMessage) (*pb.SendMessageResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg.Timestamp = time.Now().Unix()
	log.Printf("[%s] %s: %s", msg.RoomId, msg.Sender, msg.Content)
	return &pb.SendMessageResponse{
		Success: true,
		Message: "Message received successfully!",
	}, nil
}

// Streaming (bidirectional)
func (s *ChatServer) StreamChat(stream pb.ChatService_StreamChatServer) error {
	clientID := fmt.Sprintf("%p", stream)

	// simpan stream client
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

		// broadcast ke semua stream lain
		s.mu.Lock()
		for id, sstream := range s.clients {
			if id != clientID {
				_ = sstream.Send(msg)
			}
		}
		s.mu.Unlock()
	}
}

func (s *ChatServer) broadcast(msg *pb.ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, client := range s.clients {
		if id == msg.Sender {
			continue // jangan kirim balik ke pengirim
		}

		if err := client.Send(msg); err != nil {
			log.Printf("⚠️ Error kirim ke %s: %v", id, err)
		}
	}
}
