package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
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

func getParticipantByRoomId(room_id string) *livekit.ListParticipantsResponse {
	roomClient := InitRoomClient()
	res, err := roomClient.ListParticipants(context.Background(), &livekit.ListParticipantsRequest{
		Room: room_id,
	})
	if err != nil {
		log.Println(err)
	}
	return res

}

func GetParticipantHandler(w http.ResponseWriter, r *http.Request) {
	roomIdParam := r.URL.Query().Get("room_id")
	if roomIdParam == "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ERROR",
			"message": "room_id or identity_id cannot null",
		})
		return
	}
	paticipant := getParticipantByRoomId(roomIdParam)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   paticipant,
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
	muteParticipantInRoom(data)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
	})

}

func unmuteParticipantInRoom(data *ReqMuteUnmute) (*livekit.MuteRoomTrackResponse, error) {
	roomClient := InitRoomClient()
	res, err := roomClient.MutePublishedTrack(context.Background(), &livekit.MuteRoomTrackRequest{
		Room:     data.RoomID,
		Identity: data.UserID,
		TrackSid: data.TrackID,
		Muted:    false,
	})
	if err != nil {
		log.Println(err)
		return res, err
	}
	return res, nil
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
	res, err := unmuteParticipantInRoom(data)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "err",
			"message": err.Error(),
		})

	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"message": res,
		})
	}
}
