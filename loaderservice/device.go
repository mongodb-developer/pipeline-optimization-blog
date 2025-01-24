package loaderservice

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func GenerateDevice(shared bool) (string, string, string) {

	deviceSN := uuid.New()
	deviceName := RandomDeviceName(shared)
	deviceLastIP4 := generateRandomIPv4()
	deviceLastIP6 := generateRandomIPv6()
	deviceLastSeenDate := generateRandomDate(-45)
	deviceAuthorizationExpiry := generateRandomDate(30)
	deviceParentalControls := false
	if rand.Intn(10) <= 3 {
		deviceParentalControls = true
	}

	deviceData := map[string]interface{}{
		"deviceSN":                deviceSN.String(),
		"deviceName":              deviceName,
		"lastIP4":                 deviceLastIP4,
		"lastIP6":                 deviceLastIP6,
		"lastSeenDate":            deviceLastSeenDate,
		"authorizationExpiryDate": deviceAuthorizationExpiry,
		"parentalControls":        deviceParentalControls,
	}

	deviceJSON, err := json.Marshal(deviceData)
	if err != nil {
		panic(err)
	}
	return string(deviceJSON), deviceSN.String(), deviceName
}

func RandomDeviceName(shared bool) string {
	sharedDeviceNames := []string{
		"LG TV", "Samsung TV", "Sony TV", "Panasonic TV", "Vizio TV", "TCL TV", "Amazon Fire TV", "Apple TV", "Roku", "Hisense TV",
		"Philips TV", "Sharp TV", "Insignia TV", "Toshiba TV", "Xiami TV", "OnePlus TV", "Skyworth TV", "JVC TV", "Element TV", "Sceptre TV",
		"Haier TV", "Grundig TV",
	}

	personalDeviceNames := []string{
		"PlayStation 3", "PlayStation 4", "PlayStation 5", "Xbox 360", "Xbox One", "Xbox S", "Xbox X", "Nintendo Switch", "iPhone 12",
		"iPhone 13", "iPhone 14", "iPhone 15", "iPhone 16", "iPad", "iPad Mini", "iPad Air", "iPad Pro", "Amazon Fire Tablet", "Windows 10",
		"Windows 11", "Mac OSX", "Chromebook", "Meta Quest", "Android Phone", "Linux PC", "Hisense TV",
	}

	if shared {
		return sharedDeviceNames[rand.Intn(len(sharedDeviceNames))]
	} else {
		return personalDeviceNames[rand.Intn(len(personalDeviceNames))]
	}
}

// generateRandomDate generates a random date within the last "daysRange" days
func generateRandomDate(dayRange int) time.Time {

	// Get the current time
	now := time.Now()

	// Generate a random number of days (0 to 44)
	daysAgo := rand.Intn(int(math.Abs(float64(dayRange))))

	//Add (or subtract if the range was negative) the random number of days from the current time
	if dayRange < 0 {
		return now.AddDate(0, 0, -daysAgo)
	} else {
		return now.AddDate(0, 0, daysAgo)
	}
}

// generateRandomIPv4 generates a random IPv4 address as a string
func generateRandomIPv4() string {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate 4 random numbers between 0 and 255
	octet1 := rand.Intn(256)
	octet2 := rand.Intn(256)
	octet3 := rand.Intn(256)
	octet4 := rand.Intn(256)

	// Format the numbers as an IPv4 address
	return fmt.Sprintf("%d.%d.%d.%d", octet1, octet2, octet3, octet4)
}

// generateRandomIPv6 generates a random IPv6 address as a string
func generateRandomIPv6() string {

	// Generate 8 groups of 4 hex digits (16 bits each)
	segments := make([]string, 8)
	for i := 0; i < 8; i++ {
		// Each segment is a random 16-bit value represented in hexadecimal
		segments[i] = fmt.Sprintf("%x", rand.Intn(0x10000)) // 0x10000 = 65536
	}

	// Join the segments with colons to form the IPv6 address
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s",
		segments[0], segments[1], segments[2], segments[3],
		segments[4], segments[5], segments[6], segments[7])
}
