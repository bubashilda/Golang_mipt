package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type MedalCount struct {
	Gold   int `json:"gold"`
	Silver int `json:"silver"`
	Bronze int `json:"bronze"`
	Total  int `json:"total"`
}

type AthleteInfo struct {
	Athlete      string                `json:"athlete"`
	Country      string                `json:"country"`
	Medals       MedalCount            `json:"medals"`
	MedalsByYear map[string]MedalCount `json:"medals_by_year"`
}

type OlympicEvent struct {
	Athlete string `json:"athlete"`
	Country string `json:"country"`
	Year    int    `json:"year"`
	Sport   string `json:"sport"`
	Gold    int    `json:"gold"`
	Silver  int    `json:"silver"`
	Bronze  int    `json:"bronze"`
}

type DataStore struct {
	events           []OlympicEvent
	athleteCountries map[string]string // первая страна спортсмена
}

func main() {
	serverPort := flag.Int("port", 8080, "server port")
	dataFilePath := flag.String("data", "", "path to olympic winners json")
	flag.Parse()

	if *dataFilePath == "" {
		log.Fatal("data path is required")
	}

	data, err := loadData(*dataFilePath)
	if err != nil {
		log.Fatalf("failed to read data: %v", err)
	}

	http.HandleFunc("/athlete-info", handleAthleteInfoRequest(data))
	http.HandleFunc("/top-athletes-in-sport", handleTopAthletesRequest(data))
	http.HandleFunc("/top-countries-in-year", handleTopCountriesRequest(data))

	log.Printf("Starting server on port %d", *serverPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *serverPort), nil))
}

func loadData(path string) (*DataStore, error) {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var eventsList []OlympicEvent
	if err := json.Unmarshal(fileData, &eventsList); err != nil {
		return nil, err
	}

	athleteCountryMap := make(map[string]string)
	for _, event := range eventsList {
		if _, exists := athleteCountryMap[event.Athlete]; !exists {
			athleteCountryMap[event.Athlete] = event.Country
		}
	}

	return &DataStore{
		events:           eventsList,
		athleteCountries: athleteCountryMap,
	}, nil
}

func handleAthleteInfoRequest(store *DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		athleteName := r.URL.Query().Get("name")
		if athleteName == "" {
			http.Error(w, "name parameter is required", http.StatusBadRequest)
			return
		}

		athleteData := AthleteInfo{
			Athlete:      athleteName,
			Country:      store.athleteCountries[athleteName],
			Medals:       MedalCount{},
			MedalsByYear: make(map[string]MedalCount),
		}

		if athleteData.Country == "" {
			http.Error(w, fmt.Sprintf("athlete %s not found", athleteName), http.StatusNotFound)
			return
		}

		var found bool
		for _, event := range store.events {
			if event.Athlete == athleteName {
				found = true
				athleteData.Medals.Gold += event.Gold
				athleteData.Medals.Silver += event.Silver
				athleteData.Medals.Bronze += event.Bronze
				athleteData.Medals.Total += event.Gold + event.Silver + event.Bronze

				year := strconv.Itoa(event.Year)
				yearMedals := athleteData.MedalsByYear[year]
				yearMedals.Gold += event.Gold
				yearMedals.Silver += event.Silver
				yearMedals.Bronze += event.Bronze
				yearMedals.Total += event.Gold + event.Silver + event.Bronze
				athleteData.MedalsByYear[year] = yearMedals
			}
		}

		if !found {
			http.Error(w, fmt.Sprintf("athlete %s not found", athleteName), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(athleteData)
	}
}

func handleTopAthletesRequest(store *DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sportType := r.URL.Query().Get("sport")
		if sportType == "" {
			http.Error(w, "sport parameter is required", http.StatusBadRequest)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 3
		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				http.Error(w, "invalid limit parameter", http.StatusBadRequest)
				return
			}
		}

		var sportFound bool
		for _, event := range store.events {
			if strings.EqualFold(event.Sport, sportType) {
				sportFound = true
				break
			}
		}

		if !sportFound {
			http.Error(w, fmt.Sprintf("sport '%s' not found", sportType), http.StatusNotFound)
			return
		}

		athletesStats := make(map[string]AthleteInfo)
		for _, event := range store.events {
			if !strings.EqualFold(event.Sport, sportType) {
				continue
			}

			info, exists := athletesStats[event.Athlete]
			if !exists {
				info = AthleteInfo{
					Athlete:      event.Athlete,
					Country:      store.athleteCountries[event.Athlete],
					Medals:       MedalCount{},
					MedalsByYear: make(map[string]MedalCount),
				}
			}

			info.Medals.Gold += event.Gold
			info.Medals.Silver += event.Silver
			info.Medals.Bronze += event.Bronze
			info.Medals.Total += event.Gold + event.Silver + event.Bronze

			year := strconv.Itoa(event.Year)
			yearMedals := info.MedalsByYear[year]
			yearMedals.Gold += event.Gold
			yearMedals.Silver += event.Silver
			yearMedals.Bronze += event.Bronze
			yearMedals.Total += event.Gold + event.Silver + event.Bronze
			info.MedalsByYear[year] = yearMedals

			athletesStats[event.Athlete] = info
		}

		var sortedAthletes []AthleteInfo
		for _, info := range athletesStats {
			sortedAthletes = append(sortedAthletes, info)
		}

		sort.Slice(sortedAthletes, func(i, j int) bool {
			if sortedAthletes[i].Medals.Gold != sortedAthletes[j].Medals.Gold {
				return sortedAthletes[i].Medals.Gold > sortedAthletes[j].Medals.Gold
			}
			if sortedAthletes[i].Medals.Silver != sortedAthletes[j].Medals.Silver {
				return sortedAthletes[i].Medals.Silver > sortedAthletes[j].Medals.Silver
			}
			if sortedAthletes[i].Medals.Bronze != sortedAthletes[j].Medals.Bronze {
				return sortedAthletes[i].Medals.Bronze > sortedAthletes[j].Medals.Bronze
			}
			return sortedAthletes[i].Athlete < sortedAthletes[j].Athlete
		})

		if limit > len(sortedAthletes) {
			limit = len(sortedAthletes)
		}
		topAthletes := sortedAthletes[:limit]

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topAthletes)
	}
}

func handleTopCountriesRequest(store *DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		yearStr := r.URL.Query().Get("year")
		if yearStr == "" {
			http.Error(w, "year parameter is required", http.StatusBadRequest)
			return
		}

		year, err := strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "invalid year parameter", http.StatusBadRequest)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 3
		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				http.Error(w, "invalid limit parameter", http.StatusBadRequest)
				return
			}
		}

		var yearFound bool
		for _, event := range store.events {
			if event.Year == year {
				yearFound = true
				break
			}
		}

		if !yearFound {
			http.Error(w, fmt.Sprintf("year %d not found", year), http.StatusNotFound)
			return
		}

		countryMedalStats := make(map[string]MedalCount)
		for _, event := range store.events {
			if event.Year != year {
				continue
			}

			stats := countryMedalStats[event.Country]
			stats.Gold += event.Gold
			stats.Silver += event.Silver
			stats.Bronze += event.Bronze
			stats.Total += event.Gold + event.Silver + event.Bronze
			countryMedalStats[event.Country] = stats
		}

		type CountryMedals struct {
			Country string `json:"country"`
			MedalCount
		}

		var sortedCountries []CountryMedals
		for country, medals := range countryMedalStats {
			sortedCountries = append(sortedCountries, CountryMedals{
				Country:    country,
				MedalCount: medals,
			})
		}

		sort.Slice(sortedCountries, func(i, j int) bool {
			if sortedCountries[i].Gold != sortedCountries[j].Gold {
				return sortedCountries[i].Gold > sortedCountries[j].Gold
			}
			if sortedCountries[i].Silver != sortedCountries[j].Silver {
				return sortedCountries[i].Silver > sortedCountries[j].Silver
			}
			if sortedCountries[i].Bronze != sortedCountries[j].Bronze {
				return sortedCountries[i].Bronze > sortedCountries[j].Bronze
			}
			return sortedCountries[i].Country < sortedCountries[j].Country
		})

		if limit > len(sortedCountries) {
			limit = len(sortedCountries)
		}
		topCountries := sortedCountries[:limit]

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topCountries)
	}
}
