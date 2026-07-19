package notifier

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type DiscordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
	Timestamp   string `json:"timestamp"`
}

type DiscordPayload struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

// SendDiscordNotification dispatches a formatted rich embed message to a Discord Webhook URL.
func SendDiscordNotification(webhookURL string, title string, description string, color int) {
	if webhookURL == "" {
		return
	}

	payload := DiscordPayload{
		Embeds: []DiscordEmbed{
			{
				Title:       title,
				Description: description,
				Color:       color, // e.g. 65280 (Green), 16711680 (Red)
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WARNING] Failed to marshal Discord notification payload: %v\n", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Printf("[WARNING] Failed to create Discord notification HTTP request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[WARNING] Discord notification dispatch failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
}
