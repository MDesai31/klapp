// Package sms sends text messages to workers via the Twilio REST API.
package sms

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const apiURL = "https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json"

// Send sends an SMS with the given body to the given phone number (E.164
// format, e.g. +15551234567). Credentials and the sending number are read
// from the TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, and TWILIO_FROM_NUMBER
// environment variables.
func Send(to, body string) error {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	from := os.Getenv("TWILIO_FROM_NUMBER")

	switch {
	case accountSID == "":
		return fmt.Errorf("sms: TWILIO_ACCOUNT_SID is not set")
	case authToken == "":
		return fmt.Errorf("sms: TWILIO_AUTH_TOKEN is not set")
	case from == "":
		return fmt.Errorf("sms: TWILIO_FROM_NUMBER is not set")
	}

	form := url.Values{
		"To":   {to},
		"From": {from},
		"Body": {body},
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(apiURL, accountSID), strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("sms: building request: %w", err)
	}
	req.SetBasicAuth(accountSID, authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sms: sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)

		var apiErr struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("sms: twilio returned %d: %s", resp.StatusCode, apiErr.Message)
		}
		return fmt.Errorf("sms: twilio returned %d: %s", resp.StatusCode, respBody)
	}

	return nil
}
