package bdc

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	baseURL    string = "https://api.bill.com/api/v2/"
	loginURL   string = baseURL + "Login.json"
	pageMax    int    = 999
	workersMax int    = 3 // API has a max concurrent thread count of 3
)

const (
	customerSuffix        = "Customer.json"
	vendorSuffix          = "Vendor.json"
	invoiceSuffix         = "Invoice.json"
	billSuffix            = "Bill.json"
	paymentMadeSuffix     = "BillPay.json"
	paymentReceivedSuffix = "ReceivedPay.json"
	locationSuffix        = "Location.json"
	classSuffix           = "ActgClass.json"
	itemSuffix            = "Item.json"
)

type credentials struct {
	UserName string
	Password string
	OrgID    string
	DevKey   string
}

// A Client for making authenticated API calls
type Client struct {
	sessionID       string
	devKey          string
	Reports         reports
	Customer        customerResource
	Vendor          vendorResource
	Invoice         invoiceResource
	Bill            billResource
	PaymentMade     paymentMadeResource
	PaymentReceived paymentReceivedResource
	Location        locationResource
	Class           classResource
	Item            itemResource
}

type resultError struct {
	result []byte
	page   int
	err    error
}

// Shared fields across all resources
type resourceFields struct {
	suffix string
	client *Client
}

type baseResponse struct {
	Data []map[string]interface{} `json:"response_data"`
}

const timeFormat = "2006-01-02T15:04:05.999-0700"

type resourceType string

// Resource type options
const (
	Locations        resourceType = "Locations"
	Classes                       = "Classes"
	Customers                     = "Customers"
	Vendors                       = "Vendors"
	Invoices                      = "Invoices"
	Bills                         = "Bills"
	BillPayments                  = "BillPayments"
	Payments                      = "Payments"
	Items                         = "Items"
	CustomerAccounts              = "CustomerAccounts"
)

var client = new(Client)

// GetClient returns an authenticated client. Will reuse the existing client if available.
// Must provide the path to a JSON file containing complete Bill.com credentials.
func GetClient(path string) (*Client, error) {
	var creds credentials
	sid, err := login(path, &creds)
	if err != nil {
		return nil, fmt.Errorf("Unable to create new Client: %s", err)
	}
	if client.sessionID == "" {
		client = &Client{sessionID: sid, devKey: creds.DevKey}
	}

	client.Reports = reports{client: client}
	client.Customer = customerResource{resourceFields{suffix: customerSuffix, client: client}}
	client.Vendor = vendorResource{resourceFields{suffix: vendorSuffix, client: client}}
	client.Invoice = invoiceResource{resourceFields{suffix: invoiceSuffix, client: client}}
	client.Bill = billResource{resourceFields{suffix: billSuffix, client: client}}
	client.PaymentMade = paymentMadeResource{resourceFields{suffix: paymentMadeSuffix, client: client}}
	client.PaymentReceived = paymentReceivedResource{resourceFields{suffix: paymentReceivedSuffix, client: client}}
	client.Location = locationResource{resourceFields{suffix: locationSuffix, client: client}}
	client.Class = classResource{resourceFields{suffix: classSuffix, client: client}}
	client.Item = itemResource{resourceFields{suffix: itemSuffix, client: client}}
	return client, nil
}

// make an HTTP request
func makeRequest(endpoint string, body io.Reader) ([]byte, error) {
	url := baseURL + endpoint
	resp, err := http.Post(url, "application/x-www-form-urlencoded", body)
	if err != nil {
		return nil, fmt.Errorf("Unable to send Post request to %s: %s", url, err)
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read resp body from %s: %s", url, err)
	}
	err = handleError(r, url)
	if err != nil {
		return nil, fmt.Errorf("Unable to get data from page: %v", err)
	}
	return r, nil
}
