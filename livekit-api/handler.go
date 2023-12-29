package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"jumat/protocol/auth"
	"jumat/protocol/livekit"
	lksdk "jumat/server-sdk-go"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/net/websocket"
)

func getJoinToken(room, identity string) string {
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
	at := auth.NewAccessToken(apiKey, apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	at.AddGrant(grant).
		SetIdentity(identity).
		SetValidFor(24 * time.Hour)
	token, _ := at.ToJWT()
	return token
}

func raiseHand() error {
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
	info := lksdk.ConnectInfo{
		APIKey:              apiKey,
		APISecret:           apiSecret,
		RoomName:            "1",
		ParticipantIdentity: "1",
	}
	room, err := lksdk.ConnectToRoom(
		host,
		info,
		&lksdk.RoomCallback{
			OnParticipantConnected: func(*lksdk.RemoteParticipant) {
			},
		},
	)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	data := map[string]string{
		"message": "Raise Hand",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// publish lossy data to the entire room
	room.LocalParticipant.PublishData(jsonData, 0, nil)

	// publish reliable data to a set of participants
	room.LocalParticipant.PublishData(jsonData, 1, nil)
	return nil
}

func RaiseHand(w http.ResponseWriter, r *http.Request) {
	err := raiseHand()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "Fail",
			"err":    err,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "OK",
		})
	}
}

func GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	roomIdParam := r.URL.Query().Get("room_id")
	identityParam := r.URL.Query().Get("identity_id")
	if roomIdParam == "" || identityParam == "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ERROR",
			"message": "room_id or identity_id cannot null",
		})
		return
	}
	token := getJoinToken(roomIdParam, identityParam)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"token":  token,
	})
}

func getListRoom() *livekit.ListRoomsResponse {
	roomClient := InitRoomClient()
	rooms, _ := roomClient.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
	return rooms
}

func GetRoomHandler(w http.ResponseWriter, r *http.Request) {
	rooms := getListRoom()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   rooms,
	})
}

type ReqRoom struct {
	RoomID          string `json:"room_id"`
	TimeOut         string `json:"time_out"`
	MaxParticipants string `json:"max_participants"`
}

func createRoom(data *ReqRoom) *livekit.Room {
	roomClient := InitRoomClient()
	timeOut, _ := strconv.ParseInt(data.TimeOut, 10, 64)
	maxParticipants, _ := strconv.ParseInt(data.MaxParticipants, 10, 64)

	room, _ := roomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
		Name:            data.RoomID,
		EmptyTimeout:    uint32(timeOut) * 60, // TimeOut minutes
		MaxParticipants: uint32(maxParticipants),
	})
	return room
}

func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	data := &ReqRoom{
		RoomID:          r.FormValue("room_id"),
		TimeOut:         r.FormValue("time_out"),
		MaxParticipants: r.FormValue("max_participants"),
	}
	room := createRoom(data)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"room":   room,
	})
}

func getParticipantByRoomId(room_id string) ([]livekit.ParticipantData, int) {
	roomClient := InitRoomClient()
	res, err := roomClient.ListParticipants(context.Background(), &livekit.ListParticipantsRequest{
		Room: room_id,
	})
	if err != nil {
		log.Println(err)
	}
	data, err := res.GetDataParticipants()
	if err != nil {
		log.Println(err)
	}
	total := res.CountParticipants()
	if data == nil {
		// Set data to an empty slice
		data = []livekit.ParticipantData{}
	}
	return data, total
}

type Client struct {
	conn    *websocket.Conn
	roomID  string
	closeCh chan struct{}
}

func WebSocketHandler(ws *websocket.Conn) {

	roomIdParam := ws.Request().URL.Query().Get("room_id")
	if roomIdParam == "" {
		log.Println("room_id cannot be null")
		ws.Close()
		return
	}
	client := &Client{
		conn:    ws,
		roomID:  roomIdParam,
		closeCh: make(chan struct{}),
	}

	go client.handleWebSocket()

	for {
		// Get participants and total count
		participants, total := getParticipantByRoomId(roomIdParam)

		// Encode data as JSON
		data := map[string]interface{}{
			"data":  participants,
			"total": total,
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Println("Error encoding JSON:", err)
			return
		}

		// Send JSON data over WebSocket
		if _, err := ws.Write(jsonData); err != nil {
			log.Println("Error writing to WebSocket:", err)
			return
		}
		time.Sleep(3 * time.Second)
	}
}

func (c *Client) handleWebSocket() {
	defer close(c.closeCh)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.closeCh:
			// The client disconnected, exit the loop
			return
		case <-ticker.C:
			// Your WebSocket update logic here
			participants, total := getParticipantByRoomId(c.roomID)

			// Encode data as JSON
			data := map[string]interface{}{
				"data":  participants,
				"total": total,
			}
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error encoding JSON:", err)
				return
			}

			// Send JSON data over WebSocket
			if err := websocket.Message.Send(c.conn, string(jsonData)); err != nil {
				log.Println("Error writing to WebSocket:", err)
				return
			}
		}
	}
}

func writeEvent(w io.Writer, e []byte) error {
	if _, err := fmt.Fprintf(
		w, "event: LIST_PARTICIPANT\ndata: %s", e,
	); err != nil {
		return err
	}
	_, err := w.Write([]byte("\n\n"))
	return err
}

func GetParticipantHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	roomIdParam := r.URL.Query().Get("room_id")
	if roomIdParam == "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ERROR",
			"message": "room_id or identity_id cannot null",
		})
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			participant, total := getParticipantByRoomId(roomIdParam)
			data := map[string]interface{}{
				"data":  participant,
				"total": total,
			}
			jsonData, err := json.Marshal(data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := writeEvent(w, jsonData); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func GetParticipantHandler2(w http.ResponseWriter, r *http.Request) {
	roomIdParam := r.URL.Query().Get("room_id")
	if roomIdParam == "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ERROR",
			"message": "room_id or identity_id cannot null",
		})
		return
	}
	paticipant, total := getParticipantByRoomId(roomIdParam)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   paticipant,
		"total":  total,
	})
}

type ReqMuteUnmute struct {
	RoomID  string `json:"room_id"`
	UserID  string `json:"user_id"`
	TrackID string `json:"track_id"`
}

func muteParticipantInRoom(data *ReqMuteUnmute) error {
	roomClient := InitRoomClient()
	_, err := roomClient.MutePublishedTrack(context.Background(), &livekit.MuteRoomTrackRequest{
		Room:     data.RoomID,
		Identity: data.UserID,
		TrackSid: data.TrackID,
		Muted:    true,
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func MuteHandler(w http.ResponseWriter, r *http.Request) {
	roomIdParam := r.URL.Query().Get("room_id")
	userIdParam := r.URL.Query().Get("user_id")
	trackIdParam := r.URL.Query().Get("track_id")
	if roomIdParam == "" || userIdParam == "" || trackIdParam == "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ERROR",
			"message": "room_id, user_id, track_id cannot null",
		})
		return
	}
	data := &ReqMuteUnmute{
		RoomID:  roomIdParam,
		UserID:  userIdParam,
		TrackID: trackIdParam,
	}
	err := muteParticipantInRoom(data)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "err",
			"message": err.Error(),
		})

	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}
}

func unmuteParticipantInRoom(data *ReqMuteUnmute) error {
	roomClient := InitRoomClient()
	_, err := roomClient.MutePublishedTrack(context.Background(), &livekit.MuteRoomTrackRequest{
		Room:     data.RoomID,
		Identity: data.UserID,
		TrackSid: data.TrackID,
		Muted:    false,
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func UnmuteHandler(w http.ResponseWriter, r *http.Request) {
	roomIdParam := r.URL.Query().Get("room_id")
	userIdParam := r.URL.Query().Get("user_id")
	trackIdParam := r.URL.Query().Get("track_id")
	if roomIdParam == "" || userIdParam == "" || trackIdParam == "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ERROR",
			"message": "room_id, user_id, track_id cannot null",
		})
		return
	}
	data := &ReqMuteUnmute{
		RoomID:  roomIdParam,
		UserID:  userIdParam,
		TrackID: trackIdParam,
	}
	err := unmuteParticipantInRoom(data)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "err",
			"message": err.Error(),
		})

	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}
}
