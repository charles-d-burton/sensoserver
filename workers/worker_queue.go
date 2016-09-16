package workers

var (
	MaxWorker = 4
	MaxQueue  = 100
	//MaxWorker = os.Getenv("MAX_WORKERS")
	//MaxQueue  = os.Getenv("MAX_QUEUE")
)

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan WorkRequest, 100)

type WorkRequest struct {
	Name    string
	Message string
}

func AddJob(work WorkRequest) {
	WorkQueue <- work
	//log.Println("Work queued")
}
