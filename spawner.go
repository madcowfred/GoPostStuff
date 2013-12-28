package main

import (
	"sync"
)

func Spawner() {
	log.Debug("Spawner started")

	var wg sync.WaitGroup

	for name, server := range Config.Server {
		c := make(chan Article)

		log.Info("[%s] Starting %d connections...", name, server.Connections)

		for i := 0; i < server.Connections; i++ {
			// Increment the WaitGroup
			wg.Add(1)

			// Launch a goroutine for this connection
			go func(c chan Article) {
				// Decrement the counter when the goroutine completes
				defer wg.Done()

				// Create server connection
				// nntp := SimpleNNTP()

				// Connect
				// if server.TLS {
				// 	nntp.DialTLS(server.Address, server.Port)
				// } else {
				// 	nntp.Dial(server.Address, server.Port)
				// }

				// Authenticate if required
				// if server.Username && server.Password {
				// 	nntp.Auth(server.Username, server.Password)
				// }

				// Begin consuming
				for article := range c {
					log.Info("[%s] %+v", name, article)
				}
			}(c)

			log.Debug("[%s] Started connection #%d", name, i+1)
		}
	}

	// Wait for all connections to complete
	wg.Wait()
}
