package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"sort"
)

func init() {
	// Charger les variables d'environnement depuis .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Avertissement : Impossible de charger le fichier .env, utilisant les variables d'environnement système.")
	}
}

type MostPlayedResponse struct {
	Response struct {
		Ranks []struct {
			AppID int `json:"appid"`
		} `json:"ranks"`
	} `json:"response"`
}

type GameDetails struct {
	Name        string `json:"name"`
	SteamAppID  int    `json:"steam_appid"`
	HeaderImage string `json:"header_image"`
}

type GameInfo struct {
	Name           string `json:"name"`
	SteamAppID     int    `json:"steam_appid"`
	HeaderImage    string `json:"header_image"`
	CurrentPlayers int    `json:"current_players"`
}

// Obtenir les joueurs en temps réel pour un jeu spécifique
func getCurrentPlayers(appID int, apiKey string) int {
	url := fmt.Sprintf("https://api.steampowered.com/ISteamUserStats/GetNumberOfCurrentPlayers/v1/?appid=%d&key=%s", appID, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Erreur lors de la récupération des joueurs pour AppID %d : %v", appID, err)
		return 0
	}
	defer resp.Body.Close()

	var result struct {
		Response struct {
			PlayerCount int `json:"player_count"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Erreur lors du décodage des joueurs pour AppID %d : %v", appID, err)
		return 0
	}

	return result.Response.PlayerCount
}

func getTopGamesRealTime(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("STEAM_API_KEY")

	// Récupérer les jeux les plus joués
	url := fmt.Sprintf("https://api.steampowered.com/ISteamChartsService/GetMostPlayedGames/v1/?key=%s", apiKey)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des données Steam", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var mostPlayed MostPlayedResponse
	if err := json.NewDecoder(resp.Body).Decode(&mostPlayed); err != nil {
		http.Error(w, "Erreur lors du décodage des données Steam", http.StatusInternalServerError)
		return
	}

	topGames := mostPlayed.Response.Ranks[:10] // Limiter à 10 jeux pour des performances optimales
	var enrichedGames []GameInfo

	// Récupérer les détails et les joueurs en temps réel pour chaque jeu
	for _, game := range topGames {
		detailsURL := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d", game.AppID)
		detailsResp, err := http.Get(detailsURL)
		if err != nil {
			log.Printf("Erreur lors de la récupération des détails pour AppID %d : %v", game.AppID, err)
			continue
		}
		defer detailsResp.Body.Close()

		var gameDetail map[string]struct {
			Data GameDetails `json:"data"`
		}
		if err := json.NewDecoder(detailsResp.Body).Decode(&gameDetail); err != nil {
			log.Printf("Erreur lors du décodage des détails pour AppID %d : %v", game.AppID, err)
			continue
		}

		if data, ok := gameDetail[fmt.Sprintf("%d", game.AppID)]; ok && data.Data.Name != "" {
			currentPlayers := getCurrentPlayers(game.AppID, apiKey)
			enrichedGames = append(enrichedGames, GameInfo{
				Name:           data.Data.Name,
				SteamAppID:     data.Data.SteamAppID,
				HeaderImage:    data.Data.HeaderImage,
				CurrentPlayers: currentPlayers,
			})
		}
	}

	// Trier les jeux par nombre de joueurs en temps réel
	sort.Slice(enrichedGames, func(i, j int) bool {
		return enrichedGames[i].CurrentPlayers > enrichedGames[j].CurrentPlayers
	})

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(enrichedGames); err != nil {
		log.Printf("Erreur lors de l'encodage de la réponse : %v", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("Missing port")
		port = "8080"
	}
	http.HandleFunc("/top5", getTopGamesRealTime)
	fmt.Println("Serveur démarré sur le port ", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		return
	}
}
