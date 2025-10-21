package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shahrishi501/creditcardpaymentpdfparser/models"
)

func ExtractCreditCardInfo(text string) models.CreditCardInfo {
	info := models.CreditCardInfo{}

	// DEBUG: Print first 500 characters to see what we're working with
	fmt.Println("=== DEBUG: First 500 chars of extracted text ===")
	if len(text) > 500 {
		fmt.Println(text[:900])
	} else {
		fmt.Println(text)
	}
	fmt.Println("=== END DEBUG ===")

	// Normalize text: collapse multiple spaces/newlines
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Card Variant - more patterns
	cardVariantPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(visa|mastercard|amex|american express|discover|rupay)\b`),
		regexp.MustCompile(`(?i)(visa|mastercard|amex|american express|discover|rupay)\s+credit\s+card`),
	}
	for _, pattern := range cardVariantPatterns {
		if match := pattern.FindStringSubmatch(text); len(match) > 1 {
			info.CardVariant = strings.ToUpper(match[1])
			fmt.Printf("DEBUG: Found card variant: %s\n", info.CardVariant)
			break
		}
	}

	// Card Number (last 2–4 digits) - more flexible patterns
	cardPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)card\s*(?:no\.?|number)?\s*[:\-]?\s*(?:x{2,}|X{2,}|\*{2,})\s*(\d{2,4})`),
		regexp.MustCompile(`(?i)credit\s*card\s*ending\s*(?:in\s*)?(\d{2,4})`),
		regexp.MustCompile(`(?i)(?:x{4,}|X{4,}|\*{4,})\s*(\d{4})`),
		regexp.MustCompile(`(?i)card\s*ending\s*(\d{4})`),
		regexp.MustCompile(`\d{4}\s+\d{4}\s+\d{4}\s+(\d{4})`), // Full card format
		regexp.MustCompile(`(?i)a\/c\s*no[:\.\s]*(?:x+|\*+)\s*(\d{4})`),
	}
	for _, pattern := range cardPatterns {
		if match := pattern.FindStringSubmatch(text); len(match) > 1 {
			info.CardLast4Digits = match[1]
			fmt.Printf("DEBUG: Found card last digits: %s\n", info.CardLast4Digits)
			break
		}
	}

	// Billing Cycle / Statement Period - more flexible
	billingPatterns := []*regexp.Regexp{
		// Format: 01 Jan 2024 to 31 Jan 2024
		regexp.MustCompile(`(?i)(?:statement|billing)\s*(?:period|cycle|date)\s*[:\-]?\s*(\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4})\s*(?:to|-)\s*(\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4})`),
		// Format: 01/01/2024 to 31/01/2024
		regexp.MustCompile(`(?i)(?:statement|billing)\s*(?:period|cycle|date)\s*[:\-]?\s*(\d{1,2}[\/\-]\d{1,2}[\/\-]\d{2,4})\s*(?:to|-)\s*(\d{1,2}[\/\-]\d{1,2}[\/\-]\d{2,4})`),
		// Format without keywords
		regexp.MustCompile(`(?i)from[:\s]+(\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4})\s*(?:to|-)\s*(\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4})`),
	}
	for _, pattern := range billingPatterns {
		if match := pattern.FindStringSubmatch(text); len(match) > 2 {
			info.BillingCycle = fmt.Sprintf("%s to %s", strings.TrimSpace(match[1]), strings.TrimSpace(match[2]))
			fmt.Printf("DEBUG: Found billing cycle: %s\n", info.BillingCycle)
			break
		}
	}

	// Payment Due Date - more flexible
	dueDatePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)payment\s+due\s+(?:date|by)[:\s]*(\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4})`),
		regexp.MustCompile(`(?i)due\s+date[:\s]*(\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4})`),
		regexp.MustCompile(`(?i)pay\s+by[:\s]*(\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4})`),
		regexp.MustCompile(`(?i)payment\s+due[:\s]*(\d{1,2}[\/\-]\d{1,2}[\/\-]\d{2,4})`),
	}
	for _, pattern := range dueDatePatterns {
		if match := pattern.FindStringSubmatch(text); len(match) > 1 {
			info.PaymentDueDate = strings.TrimSpace(match[1])
			fmt.Printf("DEBUG: Found payment due date: %s\n", info.PaymentDueDate)
			break
		}
	}

	// Total Balance / Amount Due - more flexible
	balancePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:total\s*)?(?:amount\s*)?(?:due|outstanding|payable)[:\s]*(?:rs\.?|₹|inr)?\s*([0-9,]+\.?\d*)`),
		regexp.MustCompile(`(?i)outstanding\s*balance[:\s]*(?:rs\.?|₹|inr)?\s*([0-9,]+\.?\d*)`),
		regexp.MustCompile(`(?i)total\s*balance[:\s]*(?:rs\.?|₹|inr)?\s*([0-9,]+\.?\d*)`),
		regexp.MustCompile(`(?i)minimum\s*(?:amount\s*)?due[:\s]*(?:rs\.?|₹|inr)?\s*([0-9,]+\.?\d*)`),
		regexp.MustCompile(`(?i)current\s*(?:outstanding|balance)[:\s]*(?:rs\.?|₹|inr)?\s*([0-9,]+\.?\d*)`),
	}
	for _, pattern := range balancePatterns {
		if match := pattern.FindStringSubmatch(text); len(match) > 1 {
			info.TotalBalance = strings.ReplaceAll(strings.TrimSpace(match[1]), ",", "")
			fmt.Printf("DEBUG: Found total balance: %s\n", info.TotalBalance)
			break
		}
	}

	return info
}

func DisplayCreditCardInfo(info models.CreditCardInfo) {
	fmt.Println("\n=== CREDIT CARD INFORMATION ===")
	fmt.Printf("Card Variant: %s\n", getValueOrNA(info.CardVariant))
	fmt.Printf("Card Last 4 Digits: %s\n", getValueOrNA(info.CardLast4Digits))
	fmt.Printf("Billing Cycle: %s\n", getValueOrNA(info.BillingCycle))
	fmt.Printf("Payment Due Date: %s\n", getValueOrNA(info.PaymentDueDate))
	fmt.Printf("Total Balance: %s\n", getValueOrNA(info.TotalBalance))

	fmt.Println("\n=== TRANSACTIONS ===")
	if len(info.Transactions) > 0 {
		for i, transaction := range info.Transactions {
			fmt.Printf("%d. %s\n", i+1, transaction)
		}
	} else {
		fmt.Println("No transactions found")
	}
}

func getValueOrNA(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

// Helper function to print all regex matches for debugging
func DebugRegexMatches(text string, pattern string, description string) {
	fmt.Printf("\n=== DEBUG: Testing pattern for %s ===\n", description)
	fmt.Printf("Pattern: %s\n", pattern)
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) > 0 {
		fmt.Printf("Found %d matches:\n", len(matches))
		for i, match := range matches {
			fmt.Printf("  Match %d: %v\n", i+1, match)
		}
	} else {
		fmt.Println("No matches found")
	}
	fmt.Println("=== END DEBUG ===")
}