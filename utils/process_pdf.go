package utils

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shahrishi501/creditcardpaymentpdfparser/gemini"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

func ProcessPDF(pdfPath, password string) (string, error) {
	
	f, err := os.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("error opening PDF: %w", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return "", fmt.Errorf("error creating reader: %w", err)
	}

	if password != "" {
		auth, err := pdfReader.Decrypt([]byte(password))
		if err != nil {
			return "", fmt.Errorf("error decrypting PDF: %w", err)
		}
		if !auth {
			return "", fmt.Errorf("invalid password")
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("error getting page count: %w", err)
	}

	var allText strings.Builder

	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			log.Printf("Error loading page %d: %v", i, err)
			continue
		}

		ex, err := extractor.New(page)
		if err != nil {
			log.Printf("Error creating extractor for page %d: %v", i, err)
			continue
		}

		text, err := ex.ExtractText()
		if err != nil {
			log.Printf("Error extracting text from page %d: %v", i, err)
			continue
		}

		allText.WriteString(text)
		allText.WriteString("\n")
	}

	// Call Gemini AI Model
	result := gemini.AnalyzePDFWithGemini(
		allText.String(),
		`You are an expert financial document parser.

        Your task is to analyze the following credit card statement text and extract key structured information. 
        The text may contain formatting artifacts, OCR errors, or unusual spacing; be flexible and infer meaning from context.

        Return your answer strictly as a JSON object with these five fields:

        {
        "card_last_4": "string | null",
        "card_variant": "string | null",
        "billing_cycle": "string | null",
        "payment_due_date": "string | null",
        "total_due_amount": "string | null",
        "transactions": [
            {
                "date": "string | null",
                "description": "string | null",
                "amount": "string | null"
            }
        ] | null
        }

        Guidelines:
        - Extract only the **last four digits** of the card number.
        - The **card variant** is typically like "SBI Card PRIME", "HDFC Regalia", "Axis Flipkart Card", etc.
        - **Billing cycle** or statement period may appear as “Billing Cycle: 12 Dec 2024 10 Jan 2025”.
        - **Payment due date** might appear as “Payment Due Date: 29 Dec 2024” or similar; extract full date as string.
        - **Total due amount** or “Total Amount Due” should be captured as a number string, e.g. "₹5,342.00".
        - For **transactions**, extract each transaction's date, description, and amount. Dates may be in formats like "12/12", "12-Dec", etc.
        - If a field is not found, return null for that field.
        - Do not include any commentary or explanations outside the JSON.

        Now analyze this statement text and return the structured data.`,)

	return result, nil
}