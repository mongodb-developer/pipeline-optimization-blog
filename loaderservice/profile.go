package loaderservice

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type DeviceData struct {
	DeviceSN   string `bson:"deviceSN"`
	DeviceName string `bson:"deviceName"`
}

type ContactData struct {
	Address     map[string]string `bson:"address"`
	PhoneNumber string            `bson:"phoneNumber"`
}

type Profile struct {
	LastName     string       `bson:"lastName"`
	FirstName    string       `bson:"firstName"`
	DOB          time.Time    `bson:"DOB"`
	SSN          string       `bson:"SSN"`
	AccountNum   string       `bson:"accountNum"`
	ProfileID    string       `bson:"profileID"`
	DeviceSNs    []string     `bson:"deviceSNs"`
	Devices      []DeviceData `bson:"devices"`
	Contact      ContactData  `bson:"contact"`
	CustomerType string       `bson:"customerType"`
}

func GenerateProfile(lastName, familyID, personType, accountNum string, address map[string]string, primaryAge int, deviceSNs, deviceNames []string) (Profile, int) {

	profileID := accountNum + "-" + familyID

	firstName := randomFirstName()
	ssn := randomSSN()
	phoneNumber := randomPhoneNumber()

	var dob DateOfBirth
	if personType == "P" {
		dob, _ = generatePrimaryDOB()
	} else if personType == "S" {
		dob, _ = generateSpouseDOB(primaryAge)
	} else {
		dob, _ = generateChildDOB(primaryAge)
	}

	var devices []DeviceData
	for index, sn := range deviceSNs {
		var device DeviceData
		device.DeviceSN = sn
		device.DeviceName = deviceNames[index]
		devices = append(devices, device)
	}

	var contact ContactData
	contact.Address = address
	contact.PhoneNumber = phoneNumber

	// Create the profile data
	var profile Profile
	profile.ProfileID = profileID
	profile.AccountNum = accountNum
	profile.FirstName = firstName
	profile.LastName = lastName
	profile.DOB = dob.DOB
	profile.SSN = ssn
	profile.Contact = contact
	profile.DeviceSNs = deviceSNs
	profile.CustomerType = personType
	profile.Devices = devices

	return profile, dob.Age
}

// Generate a random first name from a predefined list of 500 names
func randomFirstName() string {
	firstNames := []string{
		"Aaron", "Abby", "Abigail", "Adam", "Adrian", "Aiden", "Alex", "Alexa", "Alexander", "Alexis",
		"Alice", "Alicia", "Alison", "Amanda", "Amber", "Amelia", "Amy", "Andrea", "Andrew", "Angela",
		"Anna", "Anthony", "April", "Ariana", "Ashley", "Austin", "Ava", "Barbara", "Ben", "Benjamin",
		"Beth", "Blake", "Brandon", "Brenda", "Brian", "Brianna", "Brittany", "Brooke", "Bryan", "Caleb",
		"Camila", "Cameron", "Carl", "Carla", "Carlos", "Carmen", "Carol", "Caroline", "Carter", "Casey",
		"Catherine", "Cecilia", "Chad", "Charles", "Charlotte", "Cheryl", "Chris", "Christian", "Christina",
		"Christopher", "Cindy", "Claire", "Clara", "Clayton", "Clifford", "Cody", "Colin", "Connor", "Courtney",
		"Cynthia", "Daisy", "Dakota", "Dale", "Daniel", "Danielle", "David", "Dean", "Deborah", "Debra",
		"Derek", "Diana", "Diane", "Diego", "Dominic", "Donald", "Donna", "Doris", "Dorothy", "Douglas",
		"Dylan", "Eddie", "Edgar", "Edward", "Edwin", "Elaine", "Elena", "Eli", "Elijah", "Elizabeth",
		"Ella", "Ellen", "Ellie", "Emily", "Emma", "Erica", "Eric", "Erin", "Ethan", "Eugene", "Eva",
		"Evan", "Evelyn", "Faith", "Fiona", "Frank", "Gabriel", "Gabriella", "Gabrielle", "Garrett", "Gavin",
		"George", "Georgia", "Gianna", "Gina", "Grace", "Gregory", "Hailey", "Hannah", "Harper", "Harrison",
		"Hazel", "Heather", "Helen", "Henry", "Holly", "Hope", "Hudson", "Hunter", "Ian", "Isaac", "Isabel",
		"Isabella", "Isaiah", "Isla", "Ivan", "Jack", "Jackson", "Jacob", "Jade", "James", "Jasmine",
		"Jason", "Jasper", "Jayden", "Jean", "Jeffrey", "Jennifer", "Jeremy", "Jessica", "Joan", "Joanna",
		"John", "Jonathan", "Jordan", "Joseph", "Joshua", "Joy", "Juan", "Judith", "Julia", "Julian",
		"Justin", "Karen", "Katherine", "Kathleen", "Kathryn", "Katie", "Kayla", "Keith", "Kelly", "Kelsey",
		"Kenneth", "Kevin", "Kimberly", "Kyle", "Laura", "Lauren", "Leah", "Leo", "Leon", "Leonard",
		"Liam", "Lillian", "Lily", "Linda", "Lisa", "Logan", "Louis", "Lucas", "Lucy", "Luis", "Luke",
		"Lydia", "Madeline", "Madison", "Maggie", "Margaret", "Maria", "Mariah", "Marie", "Marilyn",
		"Marion", "Martha", "Martin", "Mary", "Mason", "Matthew", "Megan", "Melanie", "Melissa", "Michael",
		"Michelle", "Miguel", "Mila", "Miles", "Molly", "Monica", "Morgan", "Nancy", "Natalie", "Nathan",
		"Nathaniel", "Neil", "Nicholas", "Nicole", "Noah", "Nolan", "Nora", "Norman", "Oliver", "Olivia",
		"Oscar", "Owen", "Pamela", "Patricia", "Patrick", "Paul", "Paula", "Peter", "Philip", "Phoebe",
		"Piper", "Quinn", "Rachel", "Ralph", "Raymond", "Rebecca", "Regina", "Richard", "Riley", "Robert",
		"Robin", "Roger", "Ronald", "Rose", "Ruby", "Russell", "Ryan", "Samantha", "Samuel", "Sandra",
		"Sarah", "Savannah", "Scott", "Sean", "Sebastian", "Serena", "Shane", "Shannon", "Sharon",
		"Sheila", "Sierra", "Simon", "Sophia", "Sophie", "Spencer", "Stacy", "Stanley", "Stephanie",
		"Stephen", "Steven", "Stuart", "Susan", "Sydney", "Sylvia", "Tara", "Taylor", "Teresa", "Terrence",
		"Thomas", "Tiffany", "Timothy", "Tina", "Todd", "Travis", "Tristan", "Tyler", "Vanessa", "Victoria",
		"Vincent", "Violet", "Virginia", "Walter", "Wendy", "Whitney", "William", "Willow", "Wyatt",
		"Xavier", "Yvonne", "Zachary", "Zoe",
		// Extend this array to 500 names if required
	}

	return firstNames[rand.Intn(len(firstNames))]
}

func randomAccountIDBase() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 10)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// Generate a random last name from a predefined list of 500 names
func RandomLastName() string {
	lastNames := []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
		"Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin",
		"Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson",
		"Walker", "Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
		"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell", "Carter", "Roberts",
		"Gomez", "Phillips", "Evans", "Turner", "Diaz", "Parker", "Cruz", "Edwards", "Collins", "Reyes",
		"Stewart", "Morris", "Morales", "Murphy", "Cook", "Rogers", "Gutierrez", "Ortiz", "Morgan", "Cooper",
		"Peterson", "Bailey", "Reed", "Kelly", "Howard", "Ramos", "Kim", "Cox", "Ward", "Richardson",
		"Watson", "Brooks", "Chavez", "Wood", "James", "Bennett", "Gray", "Mendoza", "Ruiz", "Hughes",
		"Price", "Alvarez", "Castillo", "Sanders", "Patel", "Myers", "Long", "Ross", "Foster", "Jimenez",
		"Powell", "Jenkins", "Perry", "Russell", "Sullivan", "Bell", "Coleman", "Butler", "Henderson", "Barnes",
		"Gonzales", "Fisher", "Vasquez", "Simmons", "Romero", "Jordan", "Patterson", "Alexander", "Hamilton", "Graham",
		"Reynolds", "Griffin", "Wallace", "Moreno", "West", "Cole", "Hayes", "Bryant", "Herrera", "Gibson",
		"Elliott", "Hunter", "Pearson", "Harper", "Fowler", "Holland", "Mendoza", "Lowe", "Duncan", "Wagner",
		"Zimmerman", "Castro", "Flores", "Shaw", "Bush", "Ford", "Chen", "Ferguson", "Cross", "Black",
		"Ray", "Webb", "Garner", "Wells", "Rice", "Dean", "Moreno", "Salazar", "Jordan", "Freeman",
		"Barker", "Greene", "Steele", "Norris", "Richards", "Fletcher", "Wheeler", "Schmidt", "Austin", "Carlson",
		"Carr", "Hansen", "Love", "Snyder", "Carpenter", "Manning", "Wade", "Shields", "Townsend", "Gates",
		// Extend this array to 500 names as needed
	}

	return lastNames[rand.Intn(len(lastNames))]
}

// Generate Random Street Address
func randomStreetName() string {
	// List of last names of all U.S. Presidents, including duplicates
	lastNames := []string{
		"Washington", "Adams", "Jefferson", "Madison", "Monroe",
		"Adams", "Jackson", "Van Buren", "Harrison", "Tyler",
		"Polk", "Taylor", "Fillmore", "Pierce", "Buchanan",
		"Lincoln", "Johnson", "Grant", "Hayes", "Garfield",
		"Arthur", "Cleveland", "Harrison", "McKinley", "Roosevelt",
		"Taft", "Wilson", "Harding", "Coolidge", "Hoover",
		"Truman", "Eisenhower", "Kennedy", "Johnson",
		"Nixon", "Ford", "Carter", "Reagan", "Bush",
		"Clinton", "Obama",
	}

	streetTypes := []string{
		"Street",     // St
		"Road",       // Rd
		"Avenue",     // Ave
		"Boulevard",  // Blvd
		"Lane",       // Ln
		"Drive",      // Dr
		"Court",      // Ct
		"Circle",     // Cir
		"Terrace",    // Ter
		"Place",      // Pl
		"Way",        // Way
		"Alley",      // Aly
		"Parkway",    // Pkwy
		"Crescent",   // Cres
		"Square",     // Sq
		"Highway",    // Hwy
		"Expressway", // Expy
		"Loop",       // Loop
		"Row",        // Row
		"Esplanade",  // Esp
		"Passage",    // Psge
		"Vista",      // Vis
		"Garden",     // Gdn
		"Pike",       // Pike
	}

	streetName := lastNames[rand.Intn(41)]
	streetNumber := strconv.Itoa(rand.Intn(9999) + 1)
	streetType := streetTypes[rand.Intn(24)]

	return streetNumber + " " + streetName + " " + streetType

}

// Generate a random SSN in the format XXX-XX-XXXX
func randomSSN() string {
	return fmt.Sprintf("%03d-%02d-%04d", rand.Intn(900)+100, rand.Intn(100), rand.Intn(10000))
}

// Generate a random 10-digit North American phone number in the format (XXX) XXX-XXXX
func randomPhoneNumber() string {
	areaCode := rand.Intn(800) + 200 // Avoids invalid area codes like 0xx or 1xx
	exchangeCode := rand.Intn(800) + 200
	lineNumber := rand.Intn(10000)
	return fmt.Sprintf("(%03d) %03d-%04d", areaCode, exchangeCode, lineNumber)
}

// Generate a random US address with city, stateCode, and zipCode
func RandomCityState() map[string]string {
	cities := []struct {
		City      string
		StateCode string
	}{
		{"New York", "NY"}, {"Los Angeles", "CA"}, {"Chicago", "IL"}, {"Houston", "TX"}, {"Phoenix", "AZ"},
		{"Philadelphia", "PA"}, {"San Antonio", "TX"}, {"San Diego", "CA"}, {"Dallas", "TX"}, {"San Jose", "CA"},
		{"Austin", "TX"}, {"Jacksonville", "FL"}, {"Fort Worth", "TX"}, {"Columbus", "OH"}, {"Charlotte", "NC"},
		{"Indianapolis", "IN"}, {"San Francisco", "CA"}, {"Seattle", "WA"}, {"Denver", "CO"},
		{"Washington", "DC"}, {"Boston", "MA"}, {"El Paso", "TX"}, {"Nashville", "TN"}, {"Detroit", "MI"},
		{"Oklahoma City", "OK"}, {"Portland", "OR"}, {"Las Vegas", "NV"}, {"Memphis", "TN"}, {"Louisville", "KY"},
		{"Baltimore", "MD"}, {"Milwaukee", "WI"}, {"Albuquerque", "NM"}, {"Tucson", "AZ"}, {"Fresno", "CA"},
		{"Mesa", "AZ"}, {"Sacramento", "CA"}, {"Atlanta", "GA"}, {"Kansas City", "MO"}, {"Colorado Springs", "CO"},
		{"Miami", "FL"}, {"Raleigh", "NC"}, {"Omaha", "NE"}, {"Long Beach", "CA"}, {"Virginia Beach", "VA"},
		{"Oakland", "CA"}, {"Minneapolis", "MN"}, {"Tulsa", "OK"}, {"Tampa", "FL"}, {"Arlington", "TX"},
		{"New Orleans", "LA"}, {"Wichita", "KS"}, {"Cleveland", "OH"}, {"Bakersfield", "CA"}, {"Aurora", "CO"},
		{"Anaheim", "CA"}, {"Honolulu", "HI"}, {"Santa Ana", "CA"}, {"Riverside", "CA"}, {"Corpus Christi", "TX"},
		{"Lexington", "KY"}, {"Stockton", "CA"}, {"Henderson", "NV"}, {"Saint Paul", "MN"}, {"St. Louis", "MO"},
		{"Cincinnati", "OH"}, {"Pittsburgh", "PA"}, {"Greensboro", "NC"}, {"Anchorage", "AK"}, {"Plano", "TX"},
		{"Lincoln", "NE"}, {"Orlando", "FL"}, {"Irvine", "CA"}, {"Newark", "NJ"}, {"Toledo", "OH"}, {"Durham", "NC"},
		{"Chula Vista", "CA"}, {"Fort Wayne", "IN"}, {"Jersey City", "NJ"}, {"St. Petersburg", "FL"},
		{"Laredo", "TX"}, {"Madison", "WI"}, {"Chandler", "AZ"}, {"Buffalo", "NY"}, {"Lubbock", "TX"},
		{"Scottsdale", "AZ"}, {"Reno", "NV"}, {"Glendale", "AZ"}, {"Gilbert", "AZ"}, {"Winston–Salem", "NC"},
		{"North Las Vegas", "NV"}, {"Norfolk", "VA"}, {"Chesapeake", "VA"}, {"Garland", "TX"}, {"Irving", "TX"},
		{"Hialeah", "FL"}, {"Fremont", "CA"}, {"Boise", "ID"}, {"Richmond", "VA"}, {"Baton Rouge", "LA"},
		{"Spokane", "WA"}, {"Des Moines", "IA"}, {"Tacoma", "WA"}, {"San Bernardino", "CA"}, {"Modesto", "CA"},
		{"Fontana", "CA"}, {"Santa Clarita", "CA"}, {"Birmingham", "AL"}, {"Oxnard", "CA"}, {"Fayetteville", "NC"},
		{"Rochester", "NY"}, {"Moreno Valley", "CA"}, {"Glendale", "CA"}, {"Yonkers", "NY"}, {"Huntington Beach", "CA"},
		{"Aurora", "IL"}, {"Salt Lake City", "UT"}, {"Amari", "TX"}, {"Montgomery", "AL"}, {"Little Rock", "AR"},
		{"Akron", "OH"}, {"Columbus", "GA"}, {"Augusta", "GA"}, {"Grand Rapids", "MI"}, {"Shreveport", "LA"},
		{"Overland Park", "KS"}, {"Tallahassee", "FL"}, {"Mobile", "AL"}, {"Knoxville", "TN"}, {"Worcester", "MA"},
		{"Tempe", "AZ"}, {"Cape Coral", "FL"}, {"Providence", "RI"}, {"Fort Lauderdale", "FL"}, {"Chattanooga", "TN"},
		{"Sioux Falls", "SD"}, {"Brownsville", "TX"}, {"Vancouver", "WA"}, {"Peoria", "AZ"}, {"New Haven", "CT"},
		{"Pasadena", "TX"}, {"McKinney", "TX"}, {"Mesquite", "TX"}, {"Savannah", "GA"}, {"Syracuse", "NY"}, {"Frisco", "TX"},
		{"Torrance", "CA"}, {"Bridgeport", "CT"}, {"McAllen", "TX"}, {"Midland", "TX"}, {"Bellevue", "WA"},
		{"Clearwater", "FL"}, {"Manchester", "NH"}, {"Topeka", "KS"}, {"Elgin", "IL"}, {"West Valley City", "UT"},
		{"Evansville", "IN"}, {"Abilene", "TX"}, {"Norman", "OK"},
	}

	// Select a random city-state pair
	cityState := cities[rand.Intn(len(cities))]

	return map[string]string{
		"city":      cityState.City,
		"stateCode": cityState.StateCode,
	}
}

// Generate a random US address with city, stateCode, and zipCode
func randomAddress() map[string]string {

	cityState := RandomCityState()

	// Generate a random ZIP code in US format (5 digits)
	zipCode := fmt.Sprintf("%05d", rand.Intn(100000))
	cityState["zipCode"] = zipCode

	cityState["street"] = randomStreetName()

	return cityState
}

// DateOfBirth holds a date and the corresponding age.
type DateOfBirth struct {
	DOB time.Time
	Age int
}

// Generate a primary date of birth for an adult aged 21-90
func generatePrimaryDOB() (DateOfBirth, error) {
	age := rand.Intn(70) + 21 // Random age between 21 and 90
	year := time.Now().Year() - age
	month := rand.Intn(12) + 1
	day := rand.Intn(28) + 1 // Simplify by assuming 28 days in every month

	dob := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return DateOfBirth{DOB: dob, Age: age}, nil
}

// Generate a spouse's date of birth based on the primary's age
func generateSpouseDOB(primaryAge int) (DateOfBirth, error) {
	if primaryAge < 21 || primaryAge > 90 {
		return DateOfBirth{}, fmt.Errorf("primary age out of range: %d", primaryAge)
	}

	ageOffset := rand.Intn(21) - 10 // Random offset between -10 and +10
	spouseAge := primaryAge + ageOffset

	// Ensure the spouse's age is between 20 and 95
	if spouseAge < 20 {
		spouseAge = 20
	} else if spouseAge > 95 {
		spouseAge = 95
	}

	year := time.Now().Year() - spouseAge
	month := rand.Intn(12) + 1
	day := rand.Intn(28) + 1

	dob := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return DateOfBirth{DOB: dob, Age: spouseAge}, nil
}

// Generate a child's date of birth based on the primary's age
func generateChildDOB(primaryAge int) (DateOfBirth, error) {
	if primaryAge < 21 || primaryAge > 90 {
		return DateOfBirth{}, fmt.Errorf("primary age out of range: %d", primaryAge)
	}

	// Calculate the range of birth years for the child
	minChildAge := 0
	maxChildAge := 24
	primaryYoungestParentAge := 21
	primaryOldestParentAge := 40

	if primaryAge > primaryOldestParentAge {

	}

	if primaryAge < primaryYoungestParentAge+minChildAge || primaryAge > primaryOldestParentAge+maxChildAge {
		return DateOfBirth{}, fmt.Errorf("primary age does not support having a child under these constraints")
	}

	childAge := rand.Intn(maxChildAge + 1) // Random age between 0 and 24
	childBirthYear := time.Now().Year() - childAge

	// Ensure the primary was between 21 and 40 when the child was born
	for childBirthYear < (time.Now().Year()-primaryAge+primaryYoungestParentAge) || childBirthYear > (time.Now().Year()-primaryAge+primaryOldestParentAge) {
		childAge = rand.Intn(maxChildAge + 1) // Random age between 0 and 24
		childBirthYear = time.Now().Year() - childAge
	}

	month := rand.Intn(12) + 1
	day := rand.Intn(28) + 1

	dob := time.Date(childBirthYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return DateOfBirth{DOB: dob, Age: childAge}, nil
}
