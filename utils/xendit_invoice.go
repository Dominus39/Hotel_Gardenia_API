package utils

import (
	"MiniProjectPhase2/entity"
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

// CreateInvoice creates an invoice via Xendit API
func CreateInvoice(booking entity.Booking, user entity.User) (*entity.Invoice, error) {
	apiKey := "xnd_development_EV8LLejJV0hNNoeeNjFIfAr5uTXg99uq2hmvJfw1KKcvoMwYGH0JOyGPq38AP4s"
	apiUrl := "https://api.xendit.co/v2/invoices"

	product := entity.ProductRequest{
		Name:  booking.Room.Name,
		Price: booking.TotalPrice,
	}

	customer := entity.CustomerRequest{
		Name:  user.Name,
		Email: user.Email,
	}

	// Prepare the invoice request body
	bodyRequest := map[string]interface{}{
		"external_id":      "booking-" + strconv.Itoa(booking.ID), // Unique external ID for the booking
		"amount":           product.Price,
		"description":      "Invoice for " + booking.Room.Name,
		"invoice_duration": 86400, // 1 day invoice expiry
		"customer": map[string]interface{}{
			"name":  customer.Name,
			"email": customer.Email,
		},
		"currency": "IDR", // Currency set to IDR
		"items": []interface{}{
			map[string]interface{}{
				"name":     product.Name,
				"quantity": 1,
				"price":    product.Price,
			},
		},
		"should_send_email": true,
	}

	// Marshal the request body to JSON
	reqBody, err := json.Marshal(bodyRequest)
	if err != nil {
		return nil, err
	}

	// Prepare the HTTP client and request
	client := &http.Client{}
	request, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Set API key for authorization and content type
	request.SetBasicAuth(apiKey, "")
	request.Header.Set("Content-Type", "application/json")

	// Send the request to Xendit API
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Parse the response to get the invoice details
	var resInvoice entity.Invoice
	if err := json.NewDecoder(response.Body).Decode(&resInvoice); err != nil {
		return nil, err
	}

	// Return the created invoice
	return &resInvoice, nil
}
