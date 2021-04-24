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
var ENTITY_NAME = `StockPrices`

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type Quote struct {
	Timestamp time.Time `json:"timestamp"`
	Symbol    string    `json:"symbol"`
	Close     float32   `json:"close"`
	High      float32   `json:"high"`
	Low       float32   `json:"low"`
	Open      float32   `json:"open"`
	Volume    int       `json:"volume"`
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
	keyBatches, valueBatches := batchQuotes(500, quotes)
	writeToDateStoreBulk(ctx, datastoreClient, keyBatches, valueBatches)
	//writeToDateStore(ctx, quotes)
	return nil
}

func parseCsv(csvQuotes string) map[string]*Quote {
	quotes := make(map[string]*Quote)
	quoteLines := strings.Split(csvQuotes, "\n")
	log.Println("Number of lines parsed ", len(quoteLines))
	log.Println("Sample line ", quoteLines[1])

	for i, quoteLine := range quoteLines {
		//Skip headers & the last new line
		if i == 0 || strings.TrimSpace(quoteLine) == "" {
			continue
		} else {
			key, quote := parseQuote(quoteLine)
			quotes[key] = quote
		}
	}
	return quotes
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func batchQuotes(limit int, quotes map[string]*Quote) ([][]string, [][]*Quote) {
	//fetch all the keys
	var keyBatches [][]string
	var valueBatches [][]*Quote

	keys := make([]string, 0, len(quotes))
	values := make([]*Quote, 0, len(quotes))

	for k, v := range quotes {
		keys = append(keys, k)
		values = append(values, v)
	}

	for i := 0; i < len(keys); i += limit {
		keyBatch := keys[i:min(i+limit, len(keys))]
		valueBatch := values[i:min(i+limit, len(keys))]
		keyBatches = append(keyBatches, keyBatch)
		valueBatches = append(valueBatches, valueBatch)
	}
	return keyBatches, valueBatches
}

func writeToDateStoreBulk(ctx context.Context, datastoreClient *datastore.Client, keyBatches [][]string, valueBatches [][]*Quote) {
	log.Println("Writting batches:", len(keyBatches))
	log.Println("Sample key ", keyBatches[0][0], valueBatches[0][0])

	for i := 0; i < len(keyBatches); i++ {
		var keys []*datastore.Key
		for j := 0; j < len(keyBatches[i]); j++ {
			keys = append(keys, datastore.NameKey(ENTITY_NAME, keyBatches[i][j], nil))
		}

		if _, err := datastoreClient.PutMulti(ctx, keys, valueBatches[i]); err != nil {
			log.Fatal(err)
		}
		log.Println("Written entry to database :", len(keys))
	}
}

// func writeToDateStore(ctx context.Context, quotes map[string]*quote) {
// 	for key, quote := range quotes {
// 		//fmt.Printf("key[%s] value[%s]\n", k, v)
// 		datastoreKey := datastore.NameKey(ENTITY_NAME, key, nil)
// 		datastoreKey, _ = datastoreClient.Put(ctx, datastoreKey, quote)
// 		log.Println("Written entry to database :", datastoreKey)
// 	}
// }

func parseQuote(quoteLine string) (string, *Quote) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred while parsing:", len(quoteLine), quoteLine, err)
		}
	}()
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

	quote := &Quote{
		Timestamp: t,
		Symbol:    symbol,
		Close:     float32(close),
		High:      float32(high),
		Low:       float32(low),
		Open:      float32(open),
		Volume:    int(volume),
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
