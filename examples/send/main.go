package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	ternwire "github.com/ternwire/ternwire-go"
)

func main() {
	client, err := ternwire.NewClient(ternwire.Config{
		APIKey: os.Getenv("TERNWIRE_API_KEY"), BaseURL: os.Getenv("TERNWIRE_API_ORIGIN"),
	})
	if err != nil {
		log.Fatal(err)
	}
	accepted, err := client.SubmitEmail(context.Background(), ternwire.Object{
		"from": os.Getenv("TERNWIRE_FROM"), "to": []string{os.Getenv("TERNWIRE_TO")},
		"subject": "Your receipt", "text": "Thanks for your order.",
	}, "order-123-receipt")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(accepted); err != nil {
		log.Fatal(err)
	}
}
