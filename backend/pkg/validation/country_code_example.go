package validation

import (
	"fmt"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// CountryCodeExample demonstrates country code validation
func CountryCodeExample() {
	// Example valid country codes
	validCountryCodes := []string{"US", "GB", "IN", "CA", "AU", "DE", "FR", "JP", "CN", "BR"}

	// Example invalid country codes
	invalidCountryCodes := []string{"XX", "ZZ", "INVALID", "123", "AB"}

	fmt.Println("=== Country Code Validation Examples ===")

	fmt.Println("\nValid Country Codes:")
	for _, code := range validCountryCodes {
		// Test with phonenumbers library
		dummyNumber := "1234567890"
		_, err := phonenumbers.Parse(dummyNumber, code)
		if err == nil {
			fmt.Printf("  ✓ %s\n", code)
		} else {
			fmt.Printf("  ✗ %s: %v\n", code, err)
		}
	}

	fmt.Println("\nInvalid Country Codes:")
	for _, code := range invalidCountryCodes {
		// Test with phonenumbers library
		dummyNumber := "1234567890"
		_, err := phonenumbers.Parse(dummyNumber, code)
		if err != nil {
			fmt.Printf("  ✓ %s (correctly rejected)\n", code)
		} else {
			fmt.Printf("  ✗ %s (should be rejected)\n", code)
		}
	}
}

// GetCommonCountryCodes returns a list of common country codes
func GetCommonCountryCodes() []string {
	return []string{
		"US", // United States
		"GB", // United Kingdom
		"IN", // India
		"CA", // Canada
		"AU", // Australia
		"DE", // Germany
		"FR", // France
		"JP", // Japan
		"CN", // China
		"BR", // Brazil
		"IT", // Italy
		"ES", // Spain
		"NL", // Netherlands
		"SE", // Sweden
		"NO", // Norway
		"DK", // Denmark
		"FI", // Finland
		"CH", // Switzerland
		"AT", // Austria
		"BE", // Belgium
		"PT", // Portugal
		"GR", // Greece
		"PL", // Poland
		"CZ", // Czech Republic
		"HU", // Hungary
		"RO", // Romania
		"BG", // Bulgaria
		"HR", // Croatia
		"SI", // Slovenia
		"SK", // Slovakia
		"LT", // Lithuania
		"LV", // Latvia
		"EE", // Estonia
		"IE", // Ireland
		"LU", // Luxembourg
		"MT", // Malta
		"CY", // Cyprus
		"MX", // Mexico
		"AR", // Argentina
		"CL", // Chile
		"PE", // Peru
		"CO", // Colombia
		"VE", // Venezuela
		"EC", // Ecuador
		"BO", // Bolivia
		"PY", // Paraguay
		"UY", // Uruguay
		"GY", // Guyana
		"SR", // Suriname
		"FK", // Falkland Islands
		"GF", // French Guiana
		"ZA", // South Africa
		"EG", // Egypt
		"NG", // Nigeria
		"KE", // Kenya
		"GH", // Ghana
		"UG", // Uganda
		"TZ", // Tanzania
		"ET", // Ethiopia
		"SD", // Sudan
		"SS", // South Sudan
		"CD", // Democratic Republic of the Congo
		"CG", // Republic of the Congo
		"CM", // Cameroon
		"CI", // Ivory Coast
		"BF", // Burkina Faso
		"ML", // Mali
		"NE", // Niger
		"TD", // Chad
		"CF", // Central African Republic
		"GA", // Gabon
		"GQ", // Equatorial Guinea
		"ST", // Sao Tome and Principe
		"GW", // Guinea-Bissau
		"GN", // Guinea
		"SL", // Sierra Leone
		"LR", // Liberia
		"TG", // Togo
		"BJ", // Benin
		"SN", // Senegal
		"GM", // Gambia
		"CV", // Cape Verde
		"MR", // Mauritania
		"MA", // Morocco
		"DZ", // Algeria
		"TN", // Tunisia
		"LY", // Libya
		"SO", // Somalia
		"DJ", // Djibouti
		"ER", // Eritrea
		"RW", // Rwanda
		"BI", // Burundi
		"MW", // Malawi
		"ZM", // Zambia
		"ZW", // Zimbabwe
		"BW", // Botswana
		"NA", // Namibia
		"LS", // Lesotho
		"SZ", // Eswatini
		"MZ", // Mozambique
		"MG", // Madagascar
		"MU", // Mauritius
		"SC", // Seychelles
		"KM", // Comoros
		"YT", // Mayotte
		"RE", // Reunion
		"TF", // French Southern Territories
		"IQ", // Iraq
		"IR", // Iran
		"SA", // Saudi Arabia
		"AE", // United Arab Emirates
		"QA", // Qatar
		"KW", // Kuwait
		"BH", // Bahrain
		"OM", // Oman
		"YE", // Yemen
		"JO", // Jordan
		"LB", // Lebanon
		"SY", // Syria
		"IL", // Israel
		"PS", // Palestine
		"TR", // Turkey
		"CY", // Cyprus
		"GR", // Greece
		"BG", // Bulgaria
		"RO", // Romania
		"MD", // Moldova
		"UA", // Ukraine
		"BY", // Belarus
		"LT", // Lithuania
		"LV", // Latvia
		"EE", // Estonia
		"RU", // Russia
		"KZ", // Kazakhstan
		"UZ", // Uzbekistan
		"TM", // Turkmenistan
		"KG", // Kyrgyzstan
		"TJ", // Tajikistan
		"AF", // Afghanistan
		"PK", // Pakistan
		"BD", // Bangladesh
		"LK", // Sri Lanka
		"MV", // Maldives
		"NP", // Nepal
		"BT", // Bhutan
		"MM", // Myanmar
		"TH", // Thailand
		"LA", // Laos
		"VN", // Vietnam
		"KH", // Cambodia
		"MY", // Malaysia
		"SG", // Singapore
		"BN", // Brunei
		"ID", // Indonesia
		"PH", // Philippines
		"TW", // Taiwan
		"HK", // Hong Kong
		"MO", // Macau
		"KR", // South Korea
		"KP", // North Korea
		"MN", // Mongolia
		"KZ", // Kazakhstan
		"UZ", // Uzbekistan
		"TM", // Turkmenistan
		"KG", // Kyrgyzstan
		"TJ", // Tajikistan
		"AF", // Afghanistan
		"PK", // Pakistan
		"BD", // Bangladesh
		"LK", // Sri Lanka
		"MV", // Maldives
		"NP", // Nepal
		"BT", // Bhutan
		"MM", // Myanmar
		"TH", // Thailand
		"LA", // Laos
		"VN", // Vietnam
		"KH", // Cambodia
		"MY", // Malaysia
		"SG", // Singapore
		"BN", // Brunei
		"ID", // Indonesia
		"PH", // Philippines
		"TW", // Taiwan
		"HK", // Hong Kong
		"MO", // Macau
		"KR", // South Korea
		"KP", // North Korea
		"MN", // Mongolia
	}
}

// ValidateCountryCodeWithLibrary validates a country code using the phonenumbers library
func ValidateCountryCodeWithLibrary(countryCode string) (bool, error) {
	if countryCode == "" {
		return true, nil
	}

	// Normalize country code to uppercase
	countryCode = strings.ToUpper(countryCode)

	// Test with a dummy phone number
	dummyNumber := "1234567890"
	_, err := phonenumbers.Parse(dummyNumber, countryCode)

	if err != nil {
		return false, fmt.Errorf("invalid country code '%s': %v", countryCode, err)
	}

	return true, nil
}
