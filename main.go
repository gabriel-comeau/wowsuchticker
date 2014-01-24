package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	API_URL        string = "http://pubapi.cryptsy.com/api.php?method=singlemarketdata&marketid="
	DOGE_ID        int    = 132
	RETRY_ATTEMPTS int    = 5
	RETRY_DELAY           = 5
	LOG_FILE_NAME  string = ".wst_log"
)

var (
	currentRetryCount int = 0
	logFile           *os.File
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
	openLogFile()
	defer closeLogFile()

	data := makeApiRequest()

	var apiresp ApiResponse

	err := json.Unmarshal(data, &apiresp)
	if err != nil {
		writeLogLine(fmt.Sprintf("Error occurred while parsing api response data: %v\n", err.Error()))
	}

	label := apiresp.Return.Markets["DOGE"].Label
	val := apiresp.Return.Markets["DOGE"].LastTradePrice

	writeLogLine(fmt.Sprintf("SUCCESS: %v : %v\n", label, val))
	fmt.Printf("%v : %v\n", label, val)
}

// Perform the actual HTTP request to the API and return the response
// as a string.  This will do some retrying if it gets a bad gateway error.
func makeApiRequest() []byte {
	reqUrl := fmt.Sprintf("%v%v", API_URL, DOGE_ID)

	resp, err := http.Get(reqUrl)
	if err != nil {
		writeLogLine(fmt.Sprintf("ERROR OCCURRED DURING HTTP REQUEST: %v\n", err.Error()))
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
			writeLogLine(fmt.Sprintf("ERROR STATUS RETURNED: %v\n", status))
			return nil
		}
	}

	body := resp.Body
	bodyText, err := ioutil.ReadAll(body)
	body.Close()

	return bodyText
}

func getLogFilePath() string {
	path := os.Getenv("HOME")
	pathStr := ""
	if path != "" {
		pathStr = path + "/" + LOG_FILE_NAME
	} else {
		pathStr = LOG_FILE_NAME
	}

	return pathStr

}

// Sets the logFile file pointer to the log file
func openLogFile() {

	filePtr, err := os.OpenFile(getLogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Couldn't open logfile for writing:", err.Error())
		return
	} else {
		logFile = filePtr
	}
}

// Closes the logFile if it is open
func closeLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}

func writeLogLine(text string) {
	if logFile != nil {
		// Add the timestamp to the string
		currentTime := time.Now()
		writeOut := fmt.Sprintf("%v : %v", currentTime, text)
		bytes := []byte(writeOut)
		_, err := logFile.Write(bytes)
		if err != nil {
			fmt.Println("ERROR WRITING TO LOGFILE:", err.Error())
		}
	}
}
