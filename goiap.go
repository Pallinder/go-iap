// Package goiap implements the ability to easily validate a receipt with apples verifyReceipt service
package goiap

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type Receipt struct {
	OriginalPurchaseDatePst string `json:"original_purchase_date_pst"`
	OriginalTransactionId   string `json:"original_transaction_id"`
	OriginalPurchaseDateMs  string `json:"original_purchase_date_ms"`
	TransactionId           string `json:"transaction_id"`
	Quantity                string `json:"quantity"`
	ProductId               string `json:"product_id"`
	Bvrs                    string `json:"bvrs"`
	PurchaseDateMs          string `json:"purchase_date_ms"`
	PurchaseDate            string `json:"purchase_date"`
	OriginalPurchaseDate    string `json:"original_purchase_date"`
	PurchaseDatePst         string `json:"purchase_date_pst"`
	Bid                     string `json:"bid"`
	ItemId                  string `json:"item_id"`
}

type receiptRequestData struct {
	Receiptdata string `json:"receipt-data"`
}

const (
	appleSandboxURL    string = "https://sandbox.itunes.apple.com/verifyReceipt"
	appleProductionURL string = "https://buy.itunes.apple.com/verifyReceipt"
)

type Error struct {
	error
}

// Given receiptData (base64 encoded) it tries to connect to either the sandbox (useSandbox true) or
// apples ordinary service (useSandbox false) to validate the receipt. Returns either a receipt struct or an error.
func VerifyReceipt(receiptData string, useSandbox bool) (*Receipt, error) {
	receipt, err := sendReceiptToApple(receiptData, verificationURL(useSandbox))
	return receipt, err
}

// Selects the proper url to use when talking to apple based on if we should use the sandbox environment or not
func verificationURL(useSandbox bool) string {

	if useSandbox {
		return appleSandboxURL
	}
	return appleProductionURL
}

// Sends the receipt to apple, returns the receipt or an error upon completion
func sendReceiptToApple(receiptData, url string) (*Receipt, error) {
	requestData, err := json.Marshal(receiptRequestData{receiptData})

	if err != nil {
		return nil, err
	}

	toSend := bytes.NewBuffer(requestData)

	resp, err := http.Post(url, "application/json", toSend)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var responseData struct {
		Status         float64  `json:"status"`
		ReceiptContent *Receipt `json:"receipt"`
	}

	responseData.ReceiptContent = new(Receipt)

	err = json.Unmarshal(body, &responseData)

	if err != nil {
		return nil, err
	}

	if responseData.Status != 0 {
		return nil, verificationError(responseData.Status)
	}

	return responseData.ReceiptContent, nil
}

// Generates the correct error based on a status error code
func verificationError(errCode float64) error {
	var errorMessage string

	switch errCode {
	case 21000:
		errorMessage = "The App Store could not read the JSON object you provided."
		break
	case 21002:
		errorMessage = "The data in the receipt-data property was malformed."
		break

	case 21003:
		errorMessage = "The receipt could not be authenticated."
		break

	case 21004:
		errorMessage = "The shared secret you provided does not match the shared secret on file for your account."
		break

	case 21005:
		errorMessage = "The receipt server is not currently available."
		break
	case 21006:
		errorMessage = "This receipt is valid but the subscription has expired. When this status code is returned to your server, " +
			"the receipt data is also decoded and returned as part of the response."
		break
	case 21007:
		errorMessage = "This receipt is a sandbox receipt, but it was sent to the production service for verification."
		break
	case 21008:
		errorMessage = "This receipt is a production receipt, but it was sent to the sandbox service for verification."
		break
	default:
		errorMessage = "An unknown error ocurred"
		break
	}

	return &Error{errors.New(errorMessage)}
}
