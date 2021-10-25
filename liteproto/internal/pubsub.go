package internal

import (
	"sync"

	"github.com/drone/liteproto/liteproto"
)

// PubSub is a simple implementation of publisher/subscriber interface
// that stores all subscribers into a in-memory map.
// It doesn't support horizontal scaling of servers.
type PubSub struct {
	subscribers map[string]chan<- liteproto.TaskResponse
	sync.Mutex
}

func (q *PubSub) Subscribe(id string) (responseCh <-chan liteproto.TaskResponse, err error) {
	q.Lock()
	defer q.Unlock()

	if q.subscribers == nil {
		q.subscribers = make(map[string]chan<- liteproto.TaskResponse)
	}

	_, ok := q.subscribers[id]
	if ok {
		err = liteproto.ErrAlreadySubscribed
		return
	}

	ch := make(chan liteproto.TaskResponse, 10)
	q.subscribers[id] = ch

	responseCh = ch

	return
}

func (q *PubSub) Unsubscribe(id string) (err error) {
	q.Lock()
	defer q.Unlock()

	_, ok := q.subscribers[id]
	if !ok {
		err = liteproto.ErrNotSubscribed
		return
	}

	delete(q.subscribers, id)

	return
}

func (q *PubSub) Publish(response liteproto.TaskResponse) (err error) {
	q.Lock()
	defer q.Unlock()

	for id, responseCh := range q.subscribers {
		if id != response.ID {
			continue
		}

		responseCh <- response
	}

	return
}
