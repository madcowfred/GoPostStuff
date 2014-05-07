package main

import (
	"fmt"
	"github.com/madcowfred/gopoststuff/simplenntp"
	"time"
)

func StatusLogger(ticker *time.Ticker, tdchan chan *simplenntp.TimeData) {
	var tds []*simplenntp.TimeData
	totalPosted := int64(0)

	for t := range ticker.C {
		stamp := t.UnixNano() / 1e6
		tds = append(tds, &simplenntp.TimeData{stamp, 0})

		// Fetch any new TimeData entries
		var breakNow bool
		for {
			breakNow = false

			select {
			case td := <-tdchan:
				// New item, add it to our list
				tds = append(tds, td)
				totalPosted += int64(td.Bytes)
			default:
				// Nothing else in the channel, done for now
				breakNow = true
			}

			if breakNow {
				break
			}
		}

		// Calculate current speed
		if len(tds) > 0 {
			active := float64(tds[len(tds)-1].Milliseconds - tds[0].Milliseconds) / 1000
			totalBytes := 0
			for _, td := range tds {
				totalBytes += td.Bytes
			}

			speed := float64(totalBytes) / float64(active) / 1024

			// Total posted
			posted := float64(totalPosted) / 1024 / 1024

			// Print it
			fmt.Printf("Posted \033[1m%.1f\033[0mMiB - Current speed: \033[1m%.1f\033[0mKiB/s             \r", posted, speed)
			//log.Debug("Current speed: %.1fKB/s", speed)
		}

		// Trim slice to only use the last 5 seconds
		earliest := stamp - 5000
		start := 0
		for i, td := range tds {
			if td.Milliseconds >= earliest {
				start = i
				break
			}
		}

		tds = tds[start:]
	}
}
