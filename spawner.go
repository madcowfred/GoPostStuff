package main

import (
	"github.com/madcowfred/gopoststuff/simplenntp"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileData struct {
	path string
	size int64
}

type Totals struct {
	start time.Time
	end   time.Time
	bytes int64
}

func Spawner(filenames []string) {
	var wg sync.WaitGroup
	files := make([]FileData, 0)

	log.Debug("Spawner started")

	// Walk any directories and collect files
	for _, filename := range filenames {
		err := filepath.Walk(filename, func(path string, fi os.FileInfo, err error) error {
			if !fi.IsDir() && fi.Size() > 0 {
				files = append(files, FileData{path: path, size: fi.Size()})
			}
			return err
		})
		if err != nil {
			log.Fatalf("Spawner walk error: %s", err)
		}
	}

	// Make a channel to stuff TimeDatas into
	tdchan := make(chan *simplenntp.TimeData, 100000)

	// Iterate over configured servers
	for name, server := range Config.Server {
		log.Info("[%s] Starting %d connections", name, server.Connections)

		// Make a channel to stuff Articles into
		achan := make(chan *Article, server.Connections)

		// Make a channel to stuff Totals into
		tchan := make(chan *Totals, server.Connections)

		// Start a goroutine to generate articles
		wg.Add(1)
		go func(c chan *Article, files []FileData) {
			defer wg.Done()

			log.Debug("[%s] Article generator started", name)

			mc := NewMmapCache()

			for filenum, fd := range files {
				// Open and mmap the file
				md, err := mc.MapFile(fd.path, len(Config.Server))
				if err != nil {
					log.Fatalf("MapFile error: %s", err)
				}

				// Work out how many parts we need
				parts := fd.size / Config.Global.ArticleSize
				rem := fd.size % Config.Global.ArticleSize
				if rem > 0 {
					parts++
				}

				// Build some articles
				for partnum := int64(0); partnum < parts; partnum++ {
					start := partnum * Config.Global.ArticleSize
					end := min((partnum+1)*Config.Global.ArticleSize, fd.size)

					ad := &ArticleData{
						PartNum:   partnum + 1,
						PartTotal: parts,
						PartSize:  end - start,
						PartBegin: start,
						PartEnd:   end,
						FileNum:   filenum + 1,
						FileTotal: len(files),
						FileSize:  fd.size,
						FileName:  filepath.Base(fd.path),
					}

					var subject string
					if *dirSubjectFlag {
						subject = filepath.Base(filepath.Dir(fd.path))
					} else {
						subject = *subjectFlag
					}

					a := NewArticle(md.data[start:end], ad, subject)
					c <- a

					//log.Debug("%s %d = %d -> %d", fd.path, i, start, end)
				}

				if md.Decrement() {
					err = mc.CloseFile(fd.path)
					if err != nil {
						log.Fatalf("CloseFile error: %s", err)
					}
					log.Debug("[%s] Closed file %s", name, fd.path)
				}
			}

			close(c)
		}(achan, files)

		// Start a goroutine for each individual connection
		for i := 0; i < server.Connections; i++ {
			connID := i + 1

			// Increment the WaitGroup counters
			wg.Add(1)
			go func(achan chan *Article, tchan chan *Totals) {
				// Decrement the counter when the goroutine completes
				defer wg.Done()

				// Connect
				log.Debug("[%s:%02d] Connecting...", name, connID)
				conn, err := simplenntp.Dial(server.Address, server.Port, server.TLS, server.InsecureSSL, tdchan)
				if err != nil {
					log.Fatalf("[%s] Error while connecting: %s", name, err)
				}
				log.Debug("[%s:%02d] Connected", name, connID)

				// Authenticate if required
				if len(server.Username) > 0 {
					log.Debug("[%s:%02d] Authenticating...", name, connID)
					err := conn.Authenticate(server.Username, server.Password)
					if err != nil {
						log.Fatalf("[%s:%02d] Error while authenticating: %s", name, connID, err)
					}
					log.Debug("[%s:%02d] Authenticated", name, connID)
				}

				log.Info("[%s:%02d] Ready", name, connID)

				t := Totals{start: time.Now()}

				// Begin consuming
				for article := range achan {
					err := conn.Post(article.Body, Config.Global.ChunkSize)
					if err != nil {
						log.Warning("[%s:%02d] Post error: %s", name, connID, err)
					}

					t.bytes += int64(len(article.Body))
				}

				// Stick our totals struct into the channel
				t.end = time.Now()
				tchan <- &t

				// Close the connection
				log.Debug("[%s:%02d] Closing connection", name, connID)
				err = conn.Quit()
				if err != nil {
					log.Warning("[%s:%02d] Error while closing connection: %s", name, connID, err)
				}
			}(achan, tchan)
		}

		// Start a goroutine to print some stats when done, sigh
		wg.Add(1)
		go func(tchan chan *Totals) {
			defer wg.Done()

			minStart := time.Now()
			var maxEnd time.Time
			var totalBytes int64

			for i := 0; i < server.Connections; i++ {
				t := <-tchan
				if t.start.Before(minStart) {
					minStart = t.start
				}
				if t.end.After(maxEnd) {
					maxEnd = t.end
				}
				totalBytes += t.bytes
			}

			// Calculate and log the result
			dur := maxEnd.Sub(minStart)
			speedKB := float64(totalBytes) / dur.Seconds() / 1024
			totalMB := float64(totalBytes) / 1024 / 1024
			log.Info("[%s] Posted %.1fMiB in %s at %.1fKB/s", name, totalMB, dur.String(), speedKB)
		}(tchan)
	}

	// Start our weird status goroutine
	statusTicker := time.NewTicker(time.Second * 1)
	go StatusLogger(statusTicker, tdchan)

	// Wait for all connections to complete
	wg.Wait()

	// Stop the status ticker
	statusTicker.Stop()
}

// math.Min wants float64s, zzz
func min(a, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}
