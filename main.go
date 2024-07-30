package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	aisstream "github.com/aisstream/ais-message-models/golang/aisStream"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

func connect() (*websocket.Conn, error) {
	aisKey := os.Getenv("AIS_STREAM_KEY")
	url := "wss://stream.aisstream.io/v0/stream"

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	subMsg := aisstream.SubscriptionMessage{
		APIKey:          aisKey,
		BoundingBoxes:   [][][]float64{{{-90.0, -180.0}, {90.0, 180.0}}}, // bounding box for the entire world
		FiltersShipMMSI: []string{"266064000"},
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

	API_URL := os.Getenv("SUPABASE_API_URL")
	API_KEY := os.Getenv("SUPABASE_API_KEY")

	db, err := supabase.NewClient(API_URL, API_KEY, nil)
	if err != nil {
		log.Fatal("cannot initalize Supabase client", err)
	}

	for {
		ws := reconnect()
		done := make(chan struct{})

		go func() {
			for {
				_, p, err := ws.ReadMessage()
				if err != nil {
					close(done)
					return
				}

				var packet aisstream.AisStreamMessage
				json.Unmarshal(p, &packet)

				switch packet.MessageType {
				case aisstream.POSITION_REPORT:
					var positionReport aisstream.PositionReport
					positionReport = *packet.Message.PositionReport
					fmt.Printf("Latitude: %f Longitude: %f\n", positionReport.Latitude, positionReport.Longitude)

					_, _, err := db.From("position_report").Insert(map[string]interface{}{
						"cog":                       positionReport.Cog,
						"communicationstate":        positionReport.CommunicationState,
						"latitude":                  positionReport.Latitude,
						"longitude":                 positionReport.Longitude,
						"messageid":                 positionReport.MessageID,
						"navigationalstatus":        positionReport.NavigationalStatus,
						"positionaccuracy":          positionReport.PositionAccuracy,
						"raim":                      positionReport.Raim,
						"rateofturn":                positionReport.RateOfTurn,
						"repeatindicator":           positionReport.RepeatIndicator,
						"sog":                       positionReport.Sog,
						"spare":                     positionReport.Spare,
						"specialmanoeuvreindicator": positionReport.SpecialManoeuvreIndicator,
						"timestamp":                 positionReport.Timestamp,
						"trueheading":               positionReport.TrueHeading,
						"userid":                    positionReport.UserID,
						"valid":                     positionReport.Valid,
					}, false, "", "", "").Execute()

					if err != nil {
						fmt.Println("Error writing to db:", err)
					}

				case aisstream.SHIP_STATIC_DATA:
					var shipStaticData aisstream.ShipStaticData
					shipStaticData = *packet.Message.ShipStaticData
					fmt.Printf("Destination: %s\n", shipStaticData.Destination)

					_, _, err := db.From("position_report").Insert(map[string]interface{}{
						"callsign":    shipStaticData.CallSign,
						"destination": shipStaticData.Destination,
						"etamonth":    shipStaticData.Eta.Month,
						"etaday":      shipStaticData.Eta.Day,
						"etahour":     shipStaticData.Eta.Hour,
						"etaminute":   shipStaticData.Eta.Minute,
						"name":        shipStaticData.Name,
					}, false, "", "", "").Execute()

					if err != nil {
						fmt.Println("Error writing to db:", err)
					}

				default:
					fmt.Printf("%s\n", packet.MessageType)
				}
			}
		}()

		<-done
		ws.Close()
		fmt.Println("Connection lost. Reconnecting...")
	}
}
