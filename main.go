package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	aisstream "github.com/aisstream/ais-message-models/golang/aisStream"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func connect() (*websocket.Conn, error) {
	aisKey := os.Getenv("AIS_STREAM_KEY")
	MMSI := os.Getenv("MMSI")
	url := "wss://stream.aisstream.io/v0/stream"

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	subMsg := aisstream.SubscriptionMessage{
		APIKey:          aisKey,
		BoundingBoxes:   [][][]float64{{{-90.0, -180.0}, {90.0, 180.0}}}, // bounding box for the entire world
		FiltersShipMMSI: []string{MMSI},
	}

	subMsgBytes, _ := json.Marshal(subMsg)
	if err := ws.WriteMessage(websocket.TextMessage, subMsgBytes); err != nil {
		return nil, err
	}
	fmt.Println("Connection open.")

	return ws, nil
}

func reconnect() *websocket.Conn {
	for {
		c, err := connect()
		if err != nil {
			fmt.Println("Failed to connect. Retrying...")
			time.Sleep(60 * time.Second)
			continue
		}

		return c
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Unable to load .env file")
	}

	for {
		ws := reconnect()
		done := make(chan struct{})

		go func() {
			_, p, err := ws.ReadMessage()
			if err != nil {
				close(done)
			}

			var packet aisstream.AisStreamMessage

			err = json.Unmarshal(p, &packet)
			if err != nil {
				close(done)
			}

			switch packet.MessageType {
			// A vessels current position. The message we'll primarily use for saving the GPS coordinates.
			// https://aisstream.io/documentation#PositionReport
			case aisstream.POSITION_REPORT:
				var positionReport aisstream.PositionReport
				positionReport = *packet.Message.PositionReport
				fmt.Printf("MMSI: %d Latitude: %f Longitude: %f\n",
					positionReport.UserID, positionReport.Latitude, positionReport.Longitude)
			// Any other incoming message, just log the message type for now, to get a sense of how often they come.
			// https://aisstream.io/documentation#API-Message-Models
			default:
				fmt.Printf("%s\n",
					packet.MessageType)
			}
		}()

		<-done
		ws.Close()
		fmt.Println("Connection lost. Reconnecting...")
	}
}
