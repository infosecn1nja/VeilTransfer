package generator

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
	"github.com/brianvoe/gofakeit/v6"
)

// CreditCardData represents the personal and credit card information for the CSV
type CreditCardData struct {
	FirstName       string
	LastName        string
	ZipCode         string
	CreditCardNumber string
	ExpirationDate  string
}

func generateFakeCreditCard() CreditCardData {
	gofakeit.Seed(time.Now().UnixNano())

	// Generate random first name, last name, and zip code
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	zipCode := gofakeit.Zip()

	// Generate random credit card details
	creditCard := gofakeit.CreditCard()
	ccn := creditCard.Number

	// Generate random expiration date (MM/YY)
	expirationMonth := gofakeit.Number(1, 12)   // Month as an integer (1-12)
	expirationYear := gofakeit.Year() % 100 // Last two digits of the year
	expirationDate := fmt.Sprintf("%02d/%02d", expirationMonth, expirationYear)

	return CreditCardData{
		FirstName:       firstName,
		LastName:        lastName,
		ZipCode:         zipCode,
		CreditCardNumber: ccn,
		ExpirationDate:  expirationDate,
	}
}

// generateCreditCards generates and saves fake credit card data to a CSV file
func GenerateCreditCards(numCards int) {
	// Open a CSV file for writing
	file, err := os.Create("credit_cards.csv")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{"First Name", "Last Name", "Zip Code", "CCN", "Expiration Date"}
	writer.Write(header)

	// Generate and write fake credit card data to CSV
	for i := 0; i < numCards; i++ {
		cardData := generateFakeCreditCard()
		record := []string{
			cardData.FirstName,
			cardData.LastName,
			cardData.ZipCode,
			cardData.CreditCardNumber,
			cardData.ExpirationDate,
		}
		writer.Write(record)
		time.Sleep(10 * time.Millisecond)		
	}

	fmt.Printf("[*] Generated %d fake credit card entries and saved to credit_cards.csv\n", numCards)
}