package generator

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"
	"github.com/brianvoe/gofakeit/v6"
)

// KTPData represents the personal information for a KTP
type KTPData struct {
	NIK           string
	Name          string
	PlaceOfBirth  string
	DateOfBirth   string
	Gender        string
	BloodType     string
	Address       string
	RT_RW         string
	Kel_Desa      string
	Kecamatan     string
	Religion      string
	MaritalStatus string
	Occupation    string
	Citizenship   string
	ValidUntil    string
}

var (
	maleFirstNames   = []string{"Ahmad", "Herman", "Budi", "Gunawan", "Hendra", "Joko", "Teguh", "Wahyu", "Yusuf"}
	femaleFirstNames = []string{"Dewi", "Dwi", "Sunarti", "Fitri", "Indah", "Kartika", "Lestari", "Murni", "Putri", "Zahra"}
	lastNames        = []string{"Sutrisno", "Setiawan", "Wahyudi", "Hidayat", "Wibowo", "Santoso", "Saputra", "Pratama", "Purnomo", "Wijaya"}
	indonesianStreets = []string{"Jalan Merdeka", "Jalan Sudirman", "Jalan Thamrin", "Jalan Diponegoro", "Jalan Gajah Mada"}
	indonesianProvinces = []string{"Jakarta", "Jawa Barat", "Jawa Tengah", "Jawa Timur", "Yogyakarta", "Bali", "Sumatera Utara", "Sumatera Barat", "Riau", "Kalimantan Timur"}
	indonesianCities    = map[string][]string{
		"Jakarta":         {"Jakarta Pusat", "Jakarta Barat", "Jakarta Timur", "Jakarta Utara", "Jakarta Selatan"},
		"Jawa Barat":      {"Bandung", "Bogor", "Bekasi", "Depok", "Cirebon"},
		"Jawa Tengah":     {"Semarang", "Surakarta", "Magelang", "Purwokerto", "Tegal"},
		"Jawa Timur":      {"Surabaya", "Malang", "Kediri", "Blitar", "Madiun"},
		"Yogyakarta":      {"Yogyakarta", "Sleman", "Bantul", "Gunungkidul", "Kulon Progo"},
		"Bali":            {"Denpasar", "Kuta", "Ubud", "Singaraja", "Gianyar"},
		"Sumatera Utara":  {"Medan", "Binjai", "Tebing Tinggi", "Pematangsiantar", "Tanjungbalai"},
		"Sumatera Barat":  {"Padang", "Bukittinggi", "Payakumbuh", "Solok", "Sawahlunto"},
		"Riau":            {"Pekanbaru", "Dumai", "Siak", "Bengkalis", "Kampar"},
		"Kalimantan Timur": {"Balikpapan", "Samarinda", "Bontang", "Tenggarong", "Kutai Barat"},
	}
	indonesianKecamatan = []string{"Sukajaya", "Pancoran Mas", "Cibadak", "Cipayung", "Cilandak", "Beji", "Tapos", "Sukmajaya", "Cengkareng", "Kemayoran"}
	indonesianDesa      = []string{"Mekarjaya", "Kebon Kacang", "Cilincing","Sukamaju", "Sindangsari", "Desakota", "Tanjungsari", "Purnawarman", "Karanganyar", "Cibubur", "Kebon Jeruk", "Tanah Abang"}
	bloodTypes          = []string{"A", "B", "AB", "O"}
	religions           = []string{"Islam", "Kristen", "Katolik", "Hindu", "Buddha", "Konghucu"}
	maritalStatuses     = []string{"Kawin", "Belum Kawin", "Cerai Hidup", "Cerai Hidup"}
	occupations         = []string{"Pegawai Negeri", "Pegawai Swasta", "Wiraswasta", "Pelajar/Mahasiswa", "Buruh", "Karyawan Honorer", "Wartawan", "Guru"}
	genders             = []string{"Laki-Laki", "Perempuan"}
)

func generateRandomKTPData() KTPData {
	gofakeit.Seed(time.Now().UnixNano())

	// Randomly select gender
	gender := genders[rand.Intn(len(genders))]

	// Generate a random name based on gender
	var firstName string
	if gender == "Laki-Laki" {
		firstName = maleFirstNames[rand.Intn(len(maleFirstNames))]
	} else {
		firstName = femaleFirstNames[rand.Intn(len(femaleFirstNames))]
	}
	lastName := lastNames[rand.Intn(len(lastNames))]
	fullName := fmt.Sprintf("%s %s", firstName, lastName)

	// Randomly select a province and corresponding city
	province := indonesianProvinces[rand.Intn(len(indonesianProvinces))]
	city := indonesianCities[province][rand.Intn(len(indonesianCities[province]))]

	// Generate a random date of birth ensuring age is at least 18
	minAge := 18
	maxAge := 80
	currentYear := time.Now().Year()
	birthYear := currentYear - gofakeit.Number(minAge, maxAge)
	birthMonth := time.Month(gofakeit.Number(1, 12)) // Convert to time.Month
	birthDay := gofakeit.Number(1, 28)               // Ensure valid day of month
	dateOfBirth := time.Date(birthYear, birthMonth, birthDay, 0, 0, 0, 0, time.UTC)
	day := dateOfBirth.Day()

	// Generate NIK (KTP Number) with real pattern
	provinceCode := fmt.Sprintf("%02d", rand.Intn(34)+11) // Random province code for Indonesia
	cityCode := fmt.Sprintf("%02d", rand.Intn(99)+1)      // Random city code
	districtCode := fmt.Sprintf("%02d", rand.Intn(99)+1)  // Random district code

	// Adjust day for females in NIK generation
	if gender == "Perempuan" {
		day += 40
	}

	dob := fmt.Sprintf("%02d%02d%02d", day, birthMonth, birthYear%100)
	uniqueNumber := fmt.Sprintf("%04d", gofakeit.Number(1, 9999))

	NIK := provinceCode + cityCode + districtCode + dob + uniqueNumber

	return KTPData{
		NIK:           NIK,
		Name:          fullName,
		PlaceOfBirth:  city,
		DateOfBirth:   dateOfBirth.Format("02-01-2006"),
		Gender:        gender,
		BloodType:     bloodTypes[rand.Intn(len(bloodTypes))],
		Address:       fmt.Sprintf("%s No.%d", indonesianStreets[rand.Intn(len(indonesianStreets))], gofakeit.Number(1, 200)),
		RT_RW:         fmt.Sprintf("%02d/%02d", gofakeit.Number(1, 20), gofakeit.Number(1, 24)),
		Kel_Desa:      indonesianDesa[rand.Intn(len(indonesianDesa))],
		Kecamatan:     indonesianKecamatan[rand.Intn(len(indonesianKecamatan))],
		Religion:      religions[rand.Intn(len(religions))],
		MaritalStatus: maritalStatuses[rand.Intn(len(maritalStatuses))],
		Occupation:    occupations[rand.Intn(len(occupations))],
		Citizenship:   "WNI",
		ValidUntil:    "Seumur Hidup",
	}
}

func GenerateKTPs(numKTPs int) {
	// Open a CSV file for writing
	file, err := os.Create("ktp.csv")
	if (err != nil) {
		fmt.Println("Unable to create file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header for KTP data
	header := []string{"NIK", "Nama", "Tempat Lahir", "Tanggal Lahir", "Jenis Kelamin", "Gol. Darah", "Alamat", "RT/RW", "Kel/Desa", "Kecamatan", "Agama", "Status Perkawinan", "Pekerjaan", "Kewarganegaraan", "Berlaku Hingga"}
	writer.Write(header)

	// Generate and write fake KTPs to CSV
	for i := 0; i < numKTPs; i++ {
		ktpData := generateRandomKTPData()
		record := []string{
			ktpData.NIK,
			ktpData.Name,
			ktpData.PlaceOfBirth,
			ktpData.DateOfBirth,
			ktpData.Gender,
			ktpData.BloodType,
			ktpData.Address,
			ktpData.RT_RW,
			ktpData.Kel_Desa,
			ktpData.Kecamatan,
			ktpData.Religion,
			ktpData.MaritalStatus,
			ktpData.Occupation,
			ktpData.Citizenship,
			ktpData.ValidUntil,
		}
		writer.Write(record)

		// Add 10ms delay to enhance randomness
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Printf("[*] Generated %d fake KTP entries and saved to ktp.csv\n", numKTPs)
}
