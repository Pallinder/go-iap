// Package goiap implements the ability to easily validate a receipt with apples verifyReceipt service
package goiap

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// Receipt is information returned by Apple
//
// Documentation: https://developer.apple.com/library/ios/releasenotes/General/ValidateAppStoreReceipt/Chapters/ReceiptFields.html#//apple_ref/doc/uid/TP40010573-CH106-SW10
type Receipt struct {
	BundleId                   string            `json:"bundle_id"`
	ApplicationVersion         string            `json:"application_version"`
	InApp                      []PurchaseReceipt `json:"in_app"`
	OriginalApplicationVersion string            `json:"original_application_version"`
}

type PurchaseReceipt struct {
	Quantity                  string `json:"quantity"`
	ProductId                 string `json:"product_id"`
	TransactionId             string `json:"transaction_id"`
	OriginalTransactionId     string `json:"original_transaction_id"`
	PurchaseDate              string `json:"purchase_date"`
	OriginalPurchaseDate      string `json:"original_purchase_date"`
	ExpiresDate               string `json:"expires_date"`
	AppItemId                 string `json:"app_item_id"`
	VersionExternalIdentifier string `json:"version_external_identifier"`
	WebOrderLineItemId        string `json:"web_order_line_item_id"`
}

type receiptRequestData struct {
	Receiptdata string `json:"receipt-data"`
}

const (
	appleSandboxURL    string = "https://sandbox.itunes.apple.com/verifyReceipt"
	appleProductionURL string = "https://buy.itunes.apple.com/verifyReceipt"
)

// Simple interface to get the original error code from the error object
type ErrorWithCode interface {
	Code() float64
}

type Error struct {
	error
	errCode float64
}

// Simple method to get the original error code from the error object
func (e *Error) Code() float64 {
	return e.errCode
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

// Error codes as they returned by the App Store
const (
	UnreadableJSON       = 21000
	MalformedData        = 21002
	AuthenticationError  = 21003
	UnmatchedSecret      = 21004
	ServerUnavailable    = 21005
	SubscriptionExpired  = 21006
	SandboxReceiptOnProd = 21007
	ProdReceiptOnSandbox = 21008
)

// Generates the correct error based on a status error code
func verificationError(errCode float64) error {
	var errorMessage string

	switch errCode {
	case UnreadableJSON:
		errorMessage = "The App Store could not read the JSON object you provided."
		break
	case MalformedData:
		errorMessage = "The data in the receipt-data property was malformed."
		break

	case AuthenticationError:
		errorMessage = "The receipt could not be authenticated."
		break

	case UnmatchedSecret:
		errorMessage = "The shared secret you provided does not match the shared secret on file for your account."
		break

	case ServerUnavailable:
		errorMessage = "The receipt server is not currently available."
		break
	case SubscriptionExpired:
		errorMessage = "This receipt is valid but the subscription has expired. When this status code is returned to your server, " +
			"the receipt data is also decoded and returned as part of the response."
		break
	case SandboxReceiptOnProd:
		errorMessage = "This receipt is a sandbox receipt, but it was sent to the production service for verification."
		break
	case ProdReceiptOnSandbox:
		errorMessage = "This receipt is a production receipt, but it was sent to the sandbox service for verification."
		break
	default:
		errorMessage = "An unknown error ocurred"
		break
	}

	return &Error{errors.New(errorMessage), errCode}
}
