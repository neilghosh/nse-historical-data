package stockprice

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

var datastoreClient *datastore.Client

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type quote struct {
	timestamp time.Time
	symbol    string
	close     float32
	high      float32
	low       float32
	open      float32
	volume    uint32
}

// HelloPubSub consumes a Pub/Sub message.
func HelloPubSub(ctx context.Context, m PubSubMessage) error {
	var csvQuotes = string(m.Data)
	log.Println("Size of pubsub message ", len(csvQuotes))

	// Set this in app.yaml when running in production.
	projectID := "demoneil"

	var err error
	datastoreClient, err = datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}

	quotes := parseCsv(csvQuotes)
	log.Println("Number of quotes parsed ", len(quotes))

	writeToDateStore(ctx, quotes)
	return nil
}

func parseCsv(csvQuotes string) map[string]*quote {
	quotes := make(map[string]*quote)
	quoteLines := strings.Split(csvQuotes, "\n")

	for i, quoteLine := range quoteLines {
		//Skip headers
		if i == 0 {
			continue
		} else {
			key, quote := parseQuote(quoteLine)
			quotes[key] = quote
		}
	}
	return quotes
}

func writeToDateStore(ctx context.Context, quotes map[string]*quote) {
	for key, quote := range quotes {
		//fmt.Printf("key[%s] value[%s]\n", k, v)
		datastoreKey := datastore.NameKey("StockPrices", key, nil)
		datastoreKey, _ = datastoreClient.Put(ctx, datastoreKey, quote)
		log.Printf(fmt.Sprintf("Written entry to database %v", datastoreKey))
	}
}

func parseQuote(quoteLine string) (string, *quote) {
	quoteValue := strings.Split(quoteLine, ",")
	symbol := quoteValue[0]
	//SYMBOL,SERIES,OPEN,HIGH,LOW,CLOSE,LAST,PREVCLOSE,TOTTRDQTY,TOTTRDVAL,TIMESTAMP,TOTALTRADES,ISIN
	//20MICRONS,EQ,36.5,36.7,35.9,36.05,36,36.05,28824,1043088.25,16-APR-2021,397,INE144J01027,
	open, _ := strconv.ParseFloat(quoteValue[2], 32)
	high, _ := strconv.ParseFloat(quoteValue[3], 32)
	low, _ := strconv.ParseFloat(quoteValue[4], 32)
	close, _ := strconv.ParseFloat(quoteValue[5], 32)
	volume, _ := strconv.ParseUint(quoteValue[8], 10, 32)

	timestamp, t := parseTimestamp(quoteValue[10])
	key := symbol + timestamp

	quote := &quote{
		timestamp: t,
		symbol:    symbol,
		close:     float32(close),
		high:      float32(high),
		low:       float32(low),
		open:      float32(open),
		volume:    uint32(volume),
	}

	return key, quote
}

func parseTimestamp(timestamp string) (string, time.Time) {
	t, err := time.Parse("2-Jan-2006", timestamp)

	if err != nil {
		fmt.Println(err)
	}
	return t.Format("2006-01-02"), t
}
