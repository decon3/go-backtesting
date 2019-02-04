package quotes

import (
	"testing"
	"errors"
	"encoding/csv"
)


func testload(t *Testing, err error) {
	var quotes = []Quote {
		Quote { Date: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 120, Low: 88, Close: 110, Volume: 100000 },
		Quote { Date: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC), Open: 109, High: 121, Low: 100, Close: 111, Volume: 200000 },
		Quote { Date: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC), Open: 99, High: 110, Low: 99, Close: 110, Volume: 100000 },
		Quote { Date: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC), Open: 120, High: 113, Low: 99, Close: 100, Volume: 200000 },
	}
	qd := convert(quotes)
	t.FailNow()
}
