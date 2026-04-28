package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"jumat/protocol/livekit"
	lksdk "jumat/server-sdk-go"

	"github.com/joho/godotenv"
	"golang.org/x/net/websocket"
)

var globalRoomClient *lksdk.RoomServiceClient

func main() {
	globalRoomClient = InitRoomClient()
	http.HandleFunc("/token", GetTokenHandler)
	http.HandleFunc("/rooms", GetRoomHandler)
	http.HandleFunc("/room", CreateRoomHandler)
	http.HandleFunc("/sse/participants", GetParticipantHandler)
	http.HandleFunc("/sse/room-status", GetRoomStatus)
	http.HandleFunc("/participants", GetParticipantHandler2)
	http.HandleFunc("/mute", MuteHandler)
	http.HandleFunc("/unmute", UnmuteHandler)
	http.HandleFunc("/raise-hand", RaiseHand)
	http.HandleFunc("/call/direct", DirectCallHandler)
	http.Handle("/ws/participant", websocket.Handler(WebSocketHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func InitRoomClient() *lksdk.RoomServiceClient {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("[WARN] Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("LK_API_KEY")
	if apiKey == "" {
		log.Println("[WARN] LK_API_KEY MISSING")
	}

	apiSecret := os.Getenv("LK_API_SECRET")
	if apiSecret == "" {
		log.Println("[WARN] LK_API_SECRET MISSING")
	}
	wsIP := os.Getenv("WEBSOCKET_IP")
	if wsIP == "" {
		log.Println("[WARN] WEBSOCKET_IP MISSING")
	}
	wsPort := os.Getenv("WEBSOCKET_PORT")
	if wsPort == "" {
		log.Println("[WARN] WEBSOCKET_PORT MISSING")
	}
	host := "ws://" + wsIP + ":" + wsPort

	log.Printf("[INFO] Connecting to LiveKit server at %s", host)
	roomClient := lksdk.NewRoomServiceClient(host, apiKey, apiSecret)

	// Test the connection with a simple ListRooms call
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = roomClient.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		log.Printf("[WARN] LiveKit connection test failed: %v", err)
	} else {
		log.Printf("[INFO] LiveKit client connected successfully")
	}

	return roomClient
}
