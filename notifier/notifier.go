package notifier

import (
	"bytes"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

//client function to receive events
/**
 * either httpCode, or errBody would be valid and not both at a time
 */
type onMessage func(event MessageEvent, messageId int, errBody string)

const (
	maxThreads     = 100
	requestTimeout = 10
)

type MessageEvent string

const (
	TimeoutEvent      MessageEvent = "TimeoutEvent"
	SuccessEvent      MessageEvent = "SuccessEvent"
	CompletedEvent    MessageEvent = "CompletedEvent"
	HttpErrorEvent    MessageEvent = "HttpErrorEvent"
	RuntimeErrorEvent MessageEvent = "RuntimeErrorEvent"
)

type MessageEvt struct {
	event MessageEvent
	item  int
	error error
}
type Notifier struct {
	url        string
	data       []string
	interval   int
	processed  int
	onMessage  onMessage
	httpClient *http.Client
	eventChan  chan MessageEvt
	interrupt  chan os.Signal
	limiter    chan struct{}
	done       chan struct{}
	mtx        sync.Mutex
}

func NewNotifier(_url string, data []string, interval int, messageFunc onMessage) *Notifier {
	_, err := url.ParseRequestURI(_url)
	if err != nil {
		panic(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	return &Notifier{
		url:        _url,
		data:       data,
		interval:   interval,
		onMessage:  messageFunc,
		httpClient: &http.Client{Timeout: requestTimeout * time.Second},
		eventChan:  make(chan MessageEvt),
		limiter:    make(chan struct{}, maxThreads),
		interrupt:  interrupt,
		processed:  0,
		mtx:        sync.Mutex{},
		done:       make(chan struct{}),
	}
}

func (n *Notifier) ProcessMessages() {


	//we would sleep the main go routine when (n.interval >=1), its necessary to process the events in a new routine so the client does
	//not wait to get event callbacks (via onMessage)
	waitForEvents := make(chan struct{})
	go func(waitForEvents chan struct{}) {
	loop:
		for {
			select {
			case <-n.done: //done processing messages
				n.onMessage(CompletedEvent, 0, "")
				break loop
			case ev, ok := <-n.eventChan:
				if ok {
					if ev.error != nil {
						n.onMessage(ev.event, ev.item, ev.error.Error())
					} else {
						n.onMessage(ev.event, ev.item, "")
					}
				}

			case <-n.interrupt:
				close(n.eventChan)
				break loop
			default:
			}

		}

		waitForEvents <- struct{}{}

	}(waitForEvents)


	for i := 0; i < len(n.data); i++ {

		//keep filling the limiter channel buffer till maxThreads then block till the channel is drained
		n.limiter <- struct{}{}

		go func(i int) {
			defer func() {
				//drain the limiter after each request completes, so a new task can run
				<-n.limiter

				//update processed count
				n.mtx.Lock()
				n.processed += 1

				if n.processed == len(n.data) {
					n.done <- struct{}{}
				}
				n.mtx.Unlock()


			}()

			n.processNextMessage(i)//make the next url call

		}(i)

		if n.interval > 0 {
			time.Sleep(time.Duration(n.interval) * time.Second)
		}
	}




	<-waitForEvents
}

func (n *Notifier) processNextMessage(idx int) {

	data := n.data[idx]
	reqBody := []byte(data)

	req, err := http.NewRequest(http.MethodPost, n.url, bytes.NewBuffer(reqBody))
	if err != nil {
		n.eventChan <- MessageEvt{
			event: RuntimeErrorEvent,
			item:  idx,
			error: err,
		}
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			n.eventChan <- MessageEvt{
				event: TimeoutEvent,
				item:  idx,
				error: err,
			}
		} else {
			n.eventChan <- MessageEvt{
				event: RuntimeErrorEvent,
				item:  idx,
				error: err,
			}
		}
		return
	}
	if resp.StatusCode != http.StatusOK {
		n.eventChan <- MessageEvt{
			event: HttpErrorEvent,
			item:  idx,
			error: errors.New(strconv.Itoa(resp.StatusCode)),
		}
		return
	}
	n.eventChan <- MessageEvt{
		event: SuccessEvent,
		item:  idx,
		error: nil,
	}
	defer resp.Body.Close()

}
