package notifier

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"sync"
	"testing"
)

func TestBundlingNotifier(t *testing.T) {

	t.Run("ShouldProcessMessages", func(t *testing.T) {
		var wg sync.WaitGroup
		var actualSuccessEventCount *int
		actualSuccessEventCount = new(int)

		expectedSuccessCount := 3
		*actualSuccessEventCount = 0

		url := "https://webhook.site/f30d570f-20ac-4211-a389-2f3696e1fa45" //update as needed

		data := []string{
			"Lorem ipsum dolor sit amet",
			"consectetur adipiscing elit. Nulla quam eros",
			"blandit luctus euismod et, aliquet hendrerit tortor",
		}
		wg.Add(len(data)+1)

		go func(successEventCount *int) {

			receiver := func(event MessageEvent, messageId int, errBody string) {
				if event == SuccessEvent {
					*successEventCount += 1
					assert.Equal(t, "", errBody)
				}
				wg.Done()
			}
			n := NewNotifier(url, data, 0,receiver)
			n.ProcessMessages()
		}(actualSuccessEventCount)

		wg.Wait()
		assert.Equal(t, expectedSuccessCount, *actualSuccessEventCount)

	})

	t.Run("ShouldReturnHttpError", func(t *testing.T) {
		var wg sync.WaitGroup
		var receivedEvent MessageEvent
		var receivedErrCode string

		url := "https://google.com" //google.com does not allow POST, should return 405

		data := []string{
			"Lorem ipsum dolor sit amet",

		}

		wg.Add(len(data)+1)
		go func() {

			receiver := func(event MessageEvent, messageId int, errBody string) {

				switch event {
				case HttpErrorEvent:
					receivedEvent = event
					receivedErrCode = errBody
				default:

				}
				wg.Done()
			}
			n := NewNotifier(url, data,0, receiver)
			n.ProcessMessages()
		}()

		wg.Wait()
		assert.Equal(t, HttpErrorEvent, receivedEvent)
		assert.Equal(t, strconv.Itoa(http.StatusMethodNotAllowed), receivedErrCode)


	})
}
