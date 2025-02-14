package generator

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
	"github.com/brianvoe/gofakeit/v6"
)

var medicalConditions = []string{
	"Hypertension", "Diabetes Mellitus", "Asthma", "Chronic Obstructive Pulmonary Disease (COPD)",
	"Coronary Artery Disease", "Heart Failure", "Stroke", "Arthritis", "Osteoporosis",
	"Anemia", "Hyperlipidemia", "Gastroesophageal Reflux Disease (GERD)", "Peptic Ulcer Disease",
	"Chronic Kidney Disease (CKD)", "Hepatitis", "HIV/AIDS", "Cancer (Lung Cancer, Breast Cancer, Prostate Cancer)",
	"Depression", "Anxiety Disorder", "Bipolar Disorder", "Schizophrenia", "Alzheimer's Disease",
	"Parkinson's Disease", "Multiple Sclerosis (MS)", "Epilepsy", "Migraine", "Psoriasis",
	"Eczema", "Allergic Rhinitis", "Sleep Apnea", "Obesity", "Chronic Pain Syndrome",
	"Irritable Bowel Syndrome (IBS)", "Crohn's Disease", "Ulcerative Colitis", "Celiac Disease",
	"Hypothyroidism", "Hyperthyroidism", "Rheumatoid Arthritis", "Systemic Lupus Erythematosus (SLE)",
	"Fibromyalgia", "Polycystic Ovary Syndrome (PCOS)", "Endometriosis", "Menopause",
}

var kondisiKesehatan = []string{
	"Hipertensi", "Diabetes Melitus", "Asma", "Penyakit Paru Obstruktif Kronis (PPOK)",
	"Penyakit Jantung Koroner", "Gagal Jantung", "Stroke", "Artritis", "Osteoporosis",
	"Anemia", "Hiperlipidemia", "Penyakit Refluks Gastroesofagus (GERD)", "Penyakit Tukak Lambung",
	"Penyakit Ginjal Kronis (PGK)", "Hepatitis", "HIV/AIDS", "Kanker Paru",
	"Depresi", "Gangguan Kecemasan", "Gangguan Bipolar", "Skizofrenia", "Penyakit Alzheimer",
	"Penyakit Parkinson", "Sklerosis Multipel (MS)", "Epilepsi", "Migrain", "Psoriasis",
	"Eksim", "Rinitis Alergi", "Apnea Tidur", "Obesitas", "Sindrom Nyeri Kronis",
	"Irritable Bowel Syndrome (IBS)", "Penyakit Crohn", "Kolitis Ulseratif", "Penyakit Celiac",
	"Hipotiroidisme", "Hipertiroidisme", "Artritis Reumatoid", "Lupus Eritematosus Sistemik (LES)",
	"Fibromyalgia", "Sindrom Ovarium Polikistik (PCOS)", "Endometriosis", "Menopause",
}

var indonesianMaleFirstNames = []string{"Ahmad", "Budi", "Fajar", "Hadi", "Joko"}
var indonesianFemaleFirstNames = []string{"Citra", "Dewi", "Eka", "Gita", "Indah"}
var englishMaleFirstNames = []string{"James", "John", "Robert", "Michael", "William"}
var englishFemaleFirstNames = []string{"Mary", "Patricia", "Jennifer", "Linda", "Elizabeth"}
var indonesianLastNames = []string{"Santoso", "Wijaya", "Pratama", "Susilo", "Setiawan", "Putri", "Lestari", "Saputra", "Nugroho", "Sari"}
var englishLastNames = []string{"Smith", "Johnson", "Williams", "Brown", "Jones"}
var medications = []string{"Paracetamol", "Ibuprofen", "Amoxicillin", "Lisinopril", "Metformin", "Amlodipine", "Omeprazole", "Atorvastatin"}

func generateName(gender, language string) string {
	if language == "id" {
		if gender == "Laki-laki" {
			return fmt.Sprintf("%s %s", indonesianMaleFirstNames[rand.Intn(len(indonesianMaleFirstNames))], indonesianLastNames[rand.Intn(len(indonesianLastNames))])
		}
		return fmt.Sprintf("%s %s", indonesianFemaleFirstNames[rand.Intn(len(indonesianFemaleFirstNames))], indonesianLastNames[rand.Intn(len(indonesianLastNames))])
	}
	if gender == "Male" {
		return fmt.Sprintf("%s %s", englishMaleFirstNames[rand.Intn(len(englishMaleFirstNames))], englishLastNames[rand.Intn(len(englishLastNames))])
	}
	return fmt.Sprintf("%s %s", englishFemaleFirstNames[rand.Intn(len(englishFemaleFirstNames))], englishLastNames[rand.Intn(len(englishLastNames))])
}

func generateData(id int, language string) []string {
	gofakeit.Seed(time.Now().UnixNano())

	gender := "Male"
	if rand.Intn(2) == 1 {
		gender = "Female"
	}

	if language == "id" {
		gender = []string{"Laki-laki", "Perempuan"}[rand.Intn(2)]
	}

	name := generateName(gender, language)
	birthDate := gofakeit.DateRange(time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2010, 12, 31, 0, 0, 0, 0, time.UTC)).Format("2006-01-02")
	visitDate := gofakeit.DateRange(time.Now().AddDate(-7, 0, 0), time.Now()).Format("2006-01-02")

	var diagnosis, prescription, doctorNote string

	if language == "id" {
		diagnosis = kondisiKesehatan[rand.Intn(len(kondisiKesehatan))]
		prescription = fmt.Sprintf("%s %dmg, diminum %d kali sehari", medications[rand.Intn(len(medications))], gofakeit.Number(50, 500), gofakeit.Number(1, 3))
		doctorNote = fmt.Sprintf("Pantau kondisi %s secara rutin dan perhatikan gejala baru.", diagnosis)
	} else {
		diagnosis = medicalConditions[rand.Intn(len(medicalConditions))]
		prescription = fmt.Sprintf("%s %dmg, take %d times a day", medications[rand.Intn(len(medications))], gofakeit.Number(50, 500), gofakeit.Number(1, 3))
		doctorNote = fmt.Sprintf("Monitor %s condition regularly and watch for new symptoms.", diagnosis)
	}

	return []string{strconv.Itoa(id), name, birthDate, gender, diagnosis, visitDate, prescription, doctorNote}
}

func GenerateMedicalRecords(numRecords int, language string) {
	fileName := "medical_records.csv"
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if language == "id" {
		writer.Write([]string{"ID Pasien", "Nama Pasien", "Tanggal Lahir", "Jenis Kelamin", "Diagnosa", "Tanggal Kunjungan", "Resep Obat", "Catatan Dokter"})
	} else {
		writer.Write([]string{"Patient ID", "Patient Name", "Date of Birth", "Gender", "Diagnosis", "Visit Date", "Prescription", "Doctor's Note"})
	}

	for i := 1; i <= numRecords; i++ {
		record := generateData(i, language)
		writer.Write(record)
		time.Sleep(10 * time.Millisecond)		
	}

	fmt.Printf("Successfully generated %d medical records and saved in %s\n", numRecords, fileName)
}