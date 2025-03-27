package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"jumat/protocol/auth"
	"jumat/protocol/livekit"
	lksdk "jumat/server-sdk-go"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/net/websocket"
)

func logRequest(r *http.Request, responseData []byte) {

	logDir := filepath.Join("..", "log")

	// Ensure log directory exists
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			log.Println("Error creating log directory:", err)
			return
		}
	}

	logFilePath := filepath.Join(logDir, "livekit.log")
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	endpointDescriptions := map[string]string{
		"/token":            "Token",
		"/rooms":            "Room List",
		"/room":             "Create Room",
		"/sse/participants": "SSE Participant In Room",
		"/sse/room-status":  "Room Status",
		"participants":      "SSE Participant In Room [2]",
		"/mute":             "Mute",
		"/unmute":           "Unmute",
		"/raise-hand":       "Raise Hand",
		"/ws/participant":   "Participant Web Socket",
	}

	description, exists := endpointDescriptions[r.URL.Path]
	if !exists {
		description = "Unknown Endpoint"
	}

	logger := log.New(file, "", log.LstdFlags)

	logEntry := fmt.Sprintf("%s %s \"%s\"", r.Method, r.URL.Path, description)

	params := ""
	for key, values := range r.URL.Query() {
		for _, value := range values {
			if params != "" {
				params += ", " // Add a separator if there are multiple params
			}
			params += fmt.Sprintf("%s:%s", key, value)
		}
	}

	if params != "" {
		logEntry += fmt.Sprintf(" Params: %s", params)
	}

	// If it's a GET request, include the response
	if r.Method == http.MethodGet {
		logEntry += fmt.Sprintf(" Response: %s", string(responseData))
	}

	logger.Println(logEntry)
}

func LogWebSocketRequest(r *http.Request) {
	file, err := os.OpenFile("livekit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)

	roomID := r.URL.Query().Get("room_id")

	// Log request details
	logger.Printf("WebSocket connection: %s %s?room_id=%s", r.Method, r.URL.Path, roomID)
}

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

func GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	var responseBuffer bytes.Buffer
	responseWriter := io.MultiWriter(w, &responseBuffer)

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
	response := map[string]interface{}{
		"status": "OK",
		"token":  token,
	}

	json.NewEncoder(responseWriter).Encode(response)

	// Log request with response
	logRequest(r, responseBuffer.Bytes())
}

type RequestRaiseHand struct {
	RoomID   string   `json:"room_id"`
	Sender   string   `json:"sender_id"`
	Receiver []string `json:"receiver_id"`
	Msg      string   `json:"msg"`
}

func raiseHand(req RequestRaiseHand) error {
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
		RoomName:            req.RoomID,
		ParticipantIdentity: req.Sender,
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
		"message":   req.Msg,
		"sender_id": req.Sender,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// publish lossy data to the entire room
	// room.LocalParticipant.PublishData(jsonData, 0, nil)

	// publish reliable data to a set of participants
	room.LocalParticipant.PublishData(jsonData, 0, req.Receiver)
	return nil
}

func RaiseHand(w http.ResponseWriter, r *http.Request) {

	logRequest(r, nil)

	var requestData RequestRaiseHand
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ERROR",
			"message": "Invalid request body",
		})
		return
	}
	err = raiseHand(requestData)
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

func getListRoom() *livekit.ListRoomsResponse {
	roomClient := InitRoomClient()
	rooms, _ := roomClient.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
	return rooms
}

func GetRoomHandler(w http.ResponseWriter, r *http.Request) {
	recorder := httptest.NewRecorder()
	rooms := getListRoom()
	responseData, _ := json.Marshal(map[string]interface{}{
		"status": "OK",
		"data":   rooms,
	})
	recorder.Write(responseData)

	logRequest(r, responseData)
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
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

	logRequest(r, nil)

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

func getRoomStatus(room_id string) bool {
	roomClient := InitRoomClient()
	res, err := roomClient.ListParticipants(context.Background(), &livekit.ListParticipantsRequest{
		Room: room_id,
	})
	if err != nil {
		log.Println(err)
	}
	total := res.CountParticipants()
	if total != 0 {
		return true
	} else {
		return false
	}
}

type Client struct {
	conn    *websocket.Conn
	roomID  string
	closeCh chan struct{}
}

func WebSocketHandler(ws *websocket.Conn) {

	LogWebSocketRequest(ws.Request())

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

			logRequest(r, jsonData)
		}
	}
}

func GetRoomStatus(w http.ResponseWriter, r *http.Request) {

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
			"message": "room_id cannot null",
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
			status := getRoomStatus(roomIdParam)
			data := map[string]interface{}{
				"active": status,
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

			logRequest(r, jsonData)
		}
	}
}

func GetParticipantHandler2(w http.ResponseWriter, r *http.Request) {

	logRequest(r, nil)

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

	logRequest(r, nil)

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

	logRequest(r, nil)

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

func RemoveParticipantInRoom(data *ReqMuteUnmute) error {
	roomClient := InitRoomClient()
	_, err := roomClient.RemoveParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     "1",
		Identity: "1",
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
