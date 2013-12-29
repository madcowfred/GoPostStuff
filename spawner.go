package main

import (
	"github.com/madcowfred/gopoststuff/simplenntp"
	"os"
	"path/filepath"
	"sync"
)

type FileData struct {
	path string
	size int64
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
	for _, fd := range files {
		log.Debug("%+v", fd)
	}

	for name, server := range Config.Server {
		log.Info("[%s] Starting %d connections", name, server.Connections)

		// Make a channel to stuff Articles into
		c := make(chan *Article, server.Connections)

		// Start a goroutine to generate articles
		wg.Add(1)
		go func(c chan *Article, files []FileData) {
			log.Debug("[%s] Article generator started", name)

			defer wg.Done()

			mc := NewMmapCache()

			for filenum, fd := range files {
				log.Debug("fd: %+v", fd)
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
						FileName:  fd.path,
					}

					a := NewArticle(md.data[start:end], ad, "test")
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
		}(c, files)

		// Start a goroutine for each individual connection
		for i := 0; i < server.Connections; i++ {
			connID := i + 1

			// Increment the WaitGroup counter
			wg.Add(1)
			go func(c chan *Article) {
				// Decrement the counter when the goroutine completes
				defer wg.Done()

				// Connect
				log.Debug("[%s:%02d] Connecting...", name, connID)
				conn, err := simplenntp.Dial(server.Address, server.Port, server.TLS)
				if err != nil {
					log.Critical("[%s] Error while connecting: %s", name, err)
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

				// Begin consuming
				for _ = range c {
					log.Info("[%s:%02d] Got an article!", name, connID)
				}

				// Close the connection
				log.Debug("[%s:%02d] Closing connection", name, connID)
				err = conn.Quit()
				if err != nil {
					log.Warning("[%s:%02d] Error while closing connection: %s", name, connID, err)
				}
			}(c)
		}
	}

	// Wait for all connections to complete
	wg.Wait()
}

func min(a, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}
