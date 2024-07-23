package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	aisstream "github.com/aisstream/ais-message-models/golang/aisStream"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func main2() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	url := "wss://stream.aisstream.io/v0/stream"
	aisKey := os.Getenv("AIS_STREAM_KEY")
	MMSI := os.Getenv("MMSI")

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer ws.Close()

	subMsg := aisstream.SubscriptionMessage{
		APIKey:          aisKey,
		BoundingBoxes:   [][][]float64{{{-90.0, -180.0}, {90.0, 180.0}}}, // bounding box for the entire world
		FiltersShipMMSI: []string{MMSI},
	}

	subMsgBytes, _ := json.Marshal(subMsg)
	if err := ws.WriteMessage(websocket.TextMessage, subMsgBytes); err != nil {
		log.Fatalln(err)
	}

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Fatalln(err)
		}
		var packet aisstream.AisStreamMessage

		err = json.Unmarshal(p, &packet)
		if err != nil {
			log.Fatalln(err)
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
			fmt.Println(packet.MessageType)
		}
	}
}
