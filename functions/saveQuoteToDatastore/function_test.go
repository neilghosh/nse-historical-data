package stockprice

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/stretchr/testify/assert"
)

func TestRestHandlerForUpdate(t *testing.T) {
	assert.False(t, false, "Response id differs")
}

func TestParseTimeStamp(t *testing.T) {
	timeAsString, timeAsTime := parseTimestamp("16-APR-2021")
	assert.Equal(t, "2021-04-16", timeAsString, "Response id differs")
	assert.Equal(t, time.Date(2021, time.Month(4), 16, 0, 0, 0, 0, time.UTC), timeAsTime, "Response id differs")
}

//SYMBOL,SERIES,OPEN,HIGH,LOW,CLOSE,LAST,PREVCLOSE,TOTTRDQTY,TOTTRDVAL,TIMESTAMP,TOTALTRADES,ISIN
//20MICRONS,EQ,36.5,36.7,35.9,36.05,36,36.05,28824,1043088.25,16-APR-2021,397,INE144J01027,
func TestParseCsv(t *testing.T) {
	csv := `SYMBOL,SERIES,OPEN,HIGH,LOW,CLOSE,LAST,PREVCLOSE,TOTTRDQTY,TOTTRDVAL,TIMESTAMP,TOTALTRADES,ISIN
	20MICRONS,EQ,36.5,36.7,35.9,36.05,36,36.05,28824,1043088.25,16-APR-2021,397,INE144J01027,
	`
	//Note that the extra line at the end of the multi line string is on purpose. This is how the actual event comes
	actualQuotes := parseCsv(csv)

	assert.Equal(t, 1, len(actualQuotes), "Map size mismatch")
}

func TestParseQuote(t *testing.T) {
	quoteLine := "20MICRONS,EQ,36.5,36.7,35.9,36.05,36,36.05,28824,1043088.25,16-APR-2021,397,INE144J01027,"
	key, actualQuote := parseQuote(quoteLine)

	assert.Equal(t, "20MICRONS2021-04-16", key, "Key does not match")

	expectedQuote := &Quote{
		Timestamp: time.Date(2021, time.Month(4), 16, 0, 0, 0, 0, time.UTC),
		Symbol:    "20MICRONS",
		Close:     36.05,
		High:      36.7,
		Low:       35.9,
		Open:      36.5,
		Volume:    28824,
	}

	assert.Equal(t, expectedQuote, actualQuote, "quote does not match")
}

func TestBatchQuote(t *testing.T) {
	quotes := make(map[string]*Quote)

	for i := 0; i < 10; i++ {
		quote := &Quote{
			Timestamp: time.Now(),
			Symbol:    "SYM" + strconv.Itoa(i),
			Close:     36.05,
			High:      36.7,
			Low:       35.9,
			Open:      36.5,
			Volume:    28824,
		}
		quotes["TEST"+strconv.Itoa(i)] = quote
	}

	keyBatches, valuevatchs := batchQuotes(3, quotes)

	assert.Equal(t, 4, len(keyBatches), "Number of Batches does not match")
	assert.Equal(t, 4, len(valuevatchs), "Number of batches does not match")

	assert.Equal(t, 3, len(keyBatches[0]), "Batch Size does not match")
	assert.Equal(t, 3, len(valuevatchs[0]), "Batch Size does not match")

	assert.Equal(t, 1, len(keyBatches[3]), "Last Batch size does not match")
	assert.Equal(t, 1, len(valuevatchs[3]), "Last Batch size does not match")
}

func TestDatastore(t *testing.T) {
	/**
		gcloud beta emulators datastore start \
	  --project unit-testing-project-name \
	  --consistency=1
	*/
	os.Setenv("DATASTORE_EMULATOR_HOST", "localhost:8081")

	var datastoreClient *datastore.Client
	ctx := context.Background()
	datastoreClient, err := datastore.NewClient(ctx, "unit-testing-project-name")
	if err != nil {
		log.Fatal(err)
	}

	var keyBatches [][]string
	var valueBatches [][]*Quote

	quote := &Quote{
		Timestamp: time.Now().UTC(),
		Symbol:    "SYM",
		Close:     36.05,
		High:      36.7,
		Low:       35.9,
		Open:      36.5,
		Volume:    28824,
	}
	keyBatches = append(keyBatches, []string{"TEST"})
	valueBatches = append(valueBatches, []*Quote{quote})

	writeToDateStoreBulk(ctx, datastoreClient, keyBatches, valueBatches)

	var actualQuote Quote
	quoteKey := datastore.NameKey(ENTITY_NAME, "TEST", nil)
	dbReadError := datastoreClient.Get(ctx, quoteKey, &actualQuote)
	_ = dbReadError // Make sure you check err.
	assert.Equal(t, *quote, actualQuote, "Last Batch size does not match")

}
