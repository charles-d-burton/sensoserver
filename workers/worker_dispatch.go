package workers

import "log"

var WorkerQueue chan chan WorkRequest

func StartDispatcher(nworkers int, ApiKey string) {
	// First, initialize the channel we are going to put the workers' work channels into.
	WorkerQueue = make(chan chan WorkRequest, nworkers)

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		log.Println("Starting worker", i+1)
		worker := NewWorker(i+1, WorkerQueue)
		worker.Start(ApiKey)
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				//log.Println("Received work requeust")
				go func() {
					//Grab a worker and send work to it
					worker := <-WorkerQueue

					//log.Println("Dispatching work request")
					worker <- work
				}()
			}
		}
	}()
}
