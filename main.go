package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	API_URL        string = "http://pubapi.cryptsy.com/api.php?method=singlemarketdata&marketid="
	DOGE_ID        int    = 132
	RETRY_ATTEMPTS int    = 5
	RETRY_DELAY           = 5
)

var (
	currentRetryCount int = 0
)

// Represents the overall ApiResponse (made up of a market which has many trades)
type ApiResponse struct {
	Success int              `json:"success"`
	Return  MarketCollection `json:"return"`
}

// A wrapper around a map so that I don't need to convert from []interface{} or
// map[string]interface{}
type MarketCollection struct {
	Markets map[string]Market `json:"markets"`
}

// Represents a market on the cryptsy exchange
//
// A lot of these values should be float64s but annoyingly everything comes back as a string
// so conversions are gonna be needed to do anything worthwhile later.
type Market struct {
	MarketId       string  `json:"marketid"`
	Label          string  `json:"label"`
	LastTradePrice string  `json:"lasttradeprice"`
	Volume         string  `json:"volume"`
	LastTradeTime  string  `json:"lasttradetime"`
	PrimaryName    string  `json:"primaryname"`
	PrimaryCode    string  `json:"primarycode"`
	SecondaryName  string  `json:"secondaryname"`
	SecondaryCode  string  `json:"secondarycode"`
	RecentTrades   []Trade `json:"recenttrades"`
	SellOrders     []Order `json:"sellorders"`
	BuyOrders      []Order `json:"buyorders"`
}

// Represents a single trade of DOGE
type Trade struct {
	Id       string
	Time     string
	Price    string
	Quantity string
	Total    string
}

// Represents a single buy or sell order
type Order struct {
	Price    string
	Quantity string
	Total    string
}

func main() {
	data := makeApiRequest()

	var apiresp ApiResponse

	err := json.Unmarshal(data, &apiresp)
	if err != nil {
		fmt.Println("Error occurred while parsing api response data:", err.Error())
	}

	label := apiresp.Return.Markets["DOGE"].Label
	val := apiresp.Return.Markets["DOGE"].LastTradePrice

	fmt.Printf("%v : %v\n", label, val)

}

// Perform the actual HTTP request to the API and return the response
// as a string.  This will do some retrying if it gets a bad gateway error.
func makeApiRequest() []byte {
	reqUrl := fmt.Sprintf("%v%v", API_URL, DOGE_ID)

	resp, err := http.Get(reqUrl)
	if err != nil {
		fmt.Println("ERROR OCCURRED DURING HTTP REQUEST:", err.Error())
	}

	status := resp.StatusCode
	if status != 200 {
		if status == 502 {
			currentRetryCount++
			if currentRetryCount > RETRY_ATTEMPTS {
				return nil
			} else {
				time.Sleep(time.Second * RETRY_DELAY)
				return makeApiRequest()
			}
		} else {
			fmt.Println("ERROR STATUS RETURNED: ", status)
			return nil
		}
	}

	body := resp.Body
	bodyText, err := ioutil.ReadAll(body)
	body.Close()

	return bodyText
}
