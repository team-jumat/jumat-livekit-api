package main

import (
	"log"
	"net/http"
	"os"

	lksdk "jumat/server-sdk-go"

	"github.com/joho/godotenv"
)

func main() {
	InitRoomClient()
	http.HandleFunc("/token", GetTokenHandler)
	http.HandleFunc("/rooms", GetRoomHandler)
	http.HandleFunc("/room", CreateRoomHandler)
	http.HandleFunc("/participants", GetParticipantHandler)
	http.HandleFunc("/mute", MuteHandler)
	http.HandleFunc("/unmute", UnmuteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func InitRoomClient() *lksdk.RoomServiceClient {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	apiKey := os.Getenv("LK_API_KEY")
	if apiKey == "" {
		log.Println("LK_API_KEY MISSING")
	}

	apiSecret := os.Getenv("LK_API_SECRET")
	if apiSecret == "" {
		log.Println("LK_API_SECRET MISSING")
	}
	wsIP := os.Getenv("WEBSOCKET_IP")
	if wsIP == "" {
		log.Println("WEBSOCKET_IP MISSING")
	}
	wsPort := os.Getenv("WEBSOCKET_PORT")
	if wsPort == "" {
		log.Println("WEBSOCKET_PORT MISSING")
	}
	host := "ws://" + wsIP + ":" + wsPort
	roomClient := lksdk.NewRoomServiceClient(host, apiKey, apiSecret)
	return roomClient
}
