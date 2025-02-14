package generator

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
	"github.com/brianvoe/gofakeit/v6"
)

// SSNData represents the personal information for the SSN CSV
type SSNData struct {
	Name     string
	SSN      string
	Gender   string
	Birthday string
	Age      int
	Address  string
}

func generateFakeSSN() string {
	// Area numbers: 001-899, excluding 666
	areaNumber := gofakeit.Number(1, 899)
	for areaNumber == 666 {
		areaNumber = gofakeit.Number(1, 899)
	}

	// Group numbers: 01-99
	groupNumber := gofakeit.Number(1, 99)

	// Serial numbers: 0001-9999
	serialNumber := gofakeit.Number(1, 9999)

	return fmt.Sprintf("%03d-%02d-%04d", areaNumber, groupNumber, serialNumber)
}

// generateSSNData generates random SSN data including name, gender, birthday, age, and address
func generateSSNData() SSNData {
	gofakeit.Seed(time.Now().UnixNano())

	// Generate random gender
	gender := gofakeit.Gender()

	// Generate a random name based on gender
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	fullName := fmt.Sprintf("%s %s", firstName, lastName)

	// Generate a random date of birth ensuring age is at least 18
	minAge := 18
	maxAge := 80
	currentYear := time.Now().Year()
	birthYear := currentYear - gofakeit.Number(minAge, maxAge)
	birthMonth := time.Month(gofakeit.Number(1, 12))
	birthDay := gofakeit.Number(1, 28)
	dateOfBirth := time.Date(birthYear, birthMonth, birthDay, 0, 0, 0, 0, time.UTC)
	birthday := dateOfBirth.Format("02-01-2006")
	age := currentYear - birthYear

	// Generate a random US address
	address := gofakeit.Address().Address

	// Generate a fake SSN
	ssn := generateFakeSSN()

	return SSNData{
		Name:     fullName,
		SSN:      ssn,
		Gender:   gender,
		Birthday: birthday,
		Age:      age,
		Address:  address,
	}
}

func GenerateSSNs(numSSNs int) {
	// Open a CSV file for writing
	file, err := os.Create("ssns.csv")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{"Name", "SSN", "Gender", "Birthday", "Age", "Address"}
	writer.Write(header)

	// Generate and write fake SSNs to CSV
	for i := 0; i < numSSNs; i++ {
		ssnData := generateSSNData()
		record := []string{
			ssnData.Name,
			ssnData.SSN,
			ssnData.Gender,
			ssnData.Birthday,
			strconv.Itoa(ssnData.Age),
			ssnData.Address,
		}
		writer.Write(record)
		time.Sleep(10 * time.Millisecond)		
	}

	fmt.Printf("[*] Generated %d fake SSNs and saved to ssns.csv\n", numSSNs)
}