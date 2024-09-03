package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

var db *supabase.Client

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Unable to load .env file")
	}

	API_URL := os.Getenv("SUPABASE_API_URL")
	API_KEY := os.Getenv("SUPABASE_API_KEY")

	newDb, err := supabase.NewClient(API_URL, API_KEY, nil)
	db = newDb

	if err != nil {
		log.Fatal("cannot initalize Supabase client", err)
	}

	fmt.Println("Starting AIS service.")

	for {
		if err := runWithRecovery(worker); err != nil {
			fmt.Printf("Worker stopped with error: %v", err)
			fmt.Println("Restarting worker in 1 minute...")
			time.Sleep(1 * time.Minute)
		}
	}
}

func runWithRecovery(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()
	return fn()
}

func worker() error {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		if err := callAPI(); err != nil {
			fmt.Printf("API call failed: %v", err)
		}
	}
}

type Cat struct {
	Height int    `json:"height"`
	Width  int    `json:"width"`
	Id     string `json:"id"`
	Url    string `json:"url"`
}
type Ais struct {
	Ais struct {
		Mmsi              int     `json:"MMSI"`
		Time              string  `json:"TIMESTAMP"`
		Latitude          float64 `json:"LATITUDE"`
		Longitude         float64 `json:"LONGITUDE"`
		Course            float64 `json:"COURSE"`
		Speed             float64 `json:"SPEED"`
		Heading           int     `json:"HEADING"`
		Navstat           int     `json:"NAVSTAT"`
		Imo               int     `json:"IMO"`
		Name              string  `json:"NAME"`
		Callsign          string  `json:"CALLSIGN"`
		Type              int     `json:"TYPE"`
		A                 int     `json:"A"`
		B                 int     `json:"B"`
		C                 int     `json:"C"`
		D                 int     `json:"D"`
		Draught           float64 `json:"DRAUGHT"`
		Destination       string  `json:"DESTINATION"`
		Locode            string  `json:"LOCODE"`
		EtaAis            string  `json:"ETA_AIS"`
		Eta               string  `json:"ETA"`
		Src               string  `json:"SRC"`
		Zone              string  `json:"ZONE"`
		Eca               bool    `json:"ECA"`
		DistanceRemaining any     `json:"DISTANCE_REMAINING"`
		EtaPredicted      any     `json:"ETA_PREDICTED"`
	} `json:"AIS"`
}

func callAPI() error {
	resp, err := http.Get("https://api.vesselfinder.com/vesselslist?userkey=WS-776925F8-8E1E9B")
	// resp, err := http.Get("https://api.thecatapi.com/v1/images/search")
	if err != nil {
		return fmt.Errorf("HTTP GET request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read Response Body: %d", err)
	}

	var target []Ais
	json.Unmarshal(body, &target)
	ais := target[0].Ais

	_, _, err = db.From("position_report").Insert(map[string]any{
		"location":      fmt.Sprintf("POINT(%f %f)", ais.Longitude, ais.Latitude),
		"time":          ais.Time,
		"course":        ais.Course,
		"speed":         ais.Speed,
		"heading":       ais.Heading,
		"draught":       ais.Draught,
		"destination":   ais.Destination,
		"eta_ais":       ais.EtaAis,
		"eta":           ais.Eta,
		"eta_predicted": ais.EtaPredicted,
	}, false, "", "", "").Execute()

	if err != nil {
		fmt.Println("Error writing to db:", err)
	}

	fmt.Printf("Logged position at location %f %f \n", ais.Longitude, ais.Latitude)

	return nil
}
