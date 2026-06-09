package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/laura838/shld-i-buy-this/agent"
	"github.com/laura838/shld-i-buy-this/db"
	"github.com/laura838/shld-i-buy-this/models"
)

const dbPath = "shld-i-buy-this.db"

const onboardingPrompt = `Welcome to shld-i-buy-this. Let's set up your profile.

For each line, describe your top spending categories, what you own, and your go-to brands.
Example: "electronics: AirPods, MacBook, Sony headphones, brands: Apple, Sony. clothing: casual basics, brands: Uniqlo, Zara."

Your profile (press Enter twice when finished):`

func main() {
	godotenv.Load()

	conn, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer conn.Close()

	profile, err := db.ReadProfile(conn)
	if err != nil {
		log.Fatalf("reading profile: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)

	if profile == nil {
		p := onboard(reader)
		if err := db.WriteProfile(conn, p); err != nil {
			log.Fatalf("saving profile: %v", err)
		}
		profile = &p
	}

	fmt.Println("Enter a product URL or description:")
	product, err := readLine(reader)
	if err != nil {
		log.Fatalf("reading product input: %v", err)
	}

	decision, err := agent.Evaluate(product, *profile)
	if err != nil {
		log.Fatalf("evaluating product: %v", err)
	}

	fmt.Println("Verdict:", decision.Label)
	fmt.Println("Reasoning:", decision.Reasoning)

	session := models.Session{
		Product:   product,
		Verdict:   decision.Label,
		Reasoning: decision.Reasoning,
		CreatedAt: time.Now(),
	}
	if err := db.WriteSession(conn, session); err != nil {
		log.Fatalf("saving session: %v", err)
	}
}

func onboard(reader *bufio.Reader) models.UserProfile {
	fmt.Println(onboardingPrompt)

	var lines []string
	for {
		line, err := readLine(reader)
		if err != nil || line == "" {
			break
		}
		lines = append(lines, line)
	}

	return models.UserProfile{PurchasePatterns: strings.Join(lines, "\n")}
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
