package stockprice

import (
	"testing"
	"time"

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
	20MICRONS,EQ,36.5,36.7,35.9,36.05,36,36.05,28824,1043088.25,16-APR-2021,397,INE144J01027,`

	actualQuotes := parseCsv(csv)

	assert.Equal(t, 1, len(actualQuotes), "Map size mismatch")
}

func TestParseQuote(t *testing.T) {
	quoteLine := "20MICRONS,EQ,36.5,36.7,35.9,36.05,36,36.05,28824,1043088.25,16-APR-2021,397,INE144J01027,"
	key, actualQuote := parseQuote(quoteLine)

	assert.Equal(t, "20MICRONS2021-04-16", key, "Key does not match")

	expectedQuote := &quote{
		timestamp: time.Date(2021, time.Month(4), 16, 0, 0, 0, 0, time.UTC),
		symbol:    "20MICRONS",
		close:     36.05,
		high:      36.7,
		low:       35.9,
		open:      36.5,
		volume:    28824,
	}

	assert.Equal(t, expectedQuote, actualQuote, "quote does not match")
}
