package internal

import (
	"sync"

	"github.com/drone/liteproto/liteproto"
)

// PubSub is a simple implementation of publisher/subscriber interface
// that stores all subscribers into a map. It doesn't support horizontal scaling of servers.
type PubSub struct {
	subscribers map[string]chan<- liteproto.Message
	sync.Mutex
}

func (q *PubSub) Subscribe(jobID string) (responseCh <-chan liteproto.Message, err error) {
	q.Lock()
	defer q.Unlock()

	if q.subscribers == nil {
		q.subscribers = make(map[string]chan<- liteproto.Message)
	}

	_, ok := q.subscribers[jobID]
	if ok {
		err = liteproto.ErrAlreadySubscribed
		return
	}

	ch := make(chan liteproto.Message, 10)
	q.subscribers[jobID] = ch

	responseCh = ch

	return
}

func (q *PubSub) Unsubscribe(jobID string) (err error) {
	q.Lock()
	defer q.Unlock()

	_, ok := q.subscribers[jobID]
	if !ok {
		err = liteproto.ErrNotSubscribed
		return
	}

	delete(q.subscribers, jobID)

	return
}

func (q *PubSub) Publish(response liteproto.Message) (err error) {
	q.Lock()
	defer q.Unlock()

	for jobID, responseCh := range q.subscribers {
		if jobID != response.MessageID {
			continue
		}

		responseCh <- response
	}

	return
}
