package main

import (
	"github.com/madcowfred/gopoststuff/simplenntp"
	"sync"
)

func Spawner() {
	log.Debug("Spawner started")

	var wg sync.WaitGroup

	for name, server := range Config.Server {
		c := make(chan Article)

		log.Info("[%s] Starting %d connections", name, server.Connections)

		for i := 0; i < server.Connections; i++ {
			connID := i + 1

			// Increment the WaitGroup
			wg.Add(1)

			// Launch a goroutine for this connection
			go func(c chan Article) {
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
				for article := range c {
					log.Info("[%s] %+v", name, article)
				}
			}(c)
		}
	}

	// Wait for all connections to complete
	wg.Wait()
}
