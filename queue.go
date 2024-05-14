package intercom

import "sync"

type Queue struct {
	Head *Message
	Size int
	mu   sync.Mutex
}

type LinkedList struct {
	next *Message
}

func (q *Queue) Push(messages ...*Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Size += len(messages)
	for idx, _ := range messages {
		var m = messages[idx]
		if q.Head == nil || m.Priority > q.Head.Priority {
			m.next = q.Head
			q.Head = m
		} else {
			current := q.Head
			for current.next != nil && m.Priority < current.next.Priority {
				current = current.next
			}
			m.next = current.next
			current.next = m
		}
	}
}

func (q *Queue) Pop() (*Message, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.Head == nil {
		return nil, false
	}
	q.Size--
	data := q.Head
	q.Head = q.Head.next
	return data, true
}

func (q *Queue) Length() int {
	return q.Size
}

func (q *Queue) All() []*Message {
	var nodes []*Message
	node := q.Head
	for node != nil {
		nodes = append(nodes, node)
		node = node.next
	}
	return nodes
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Head = nil
}
