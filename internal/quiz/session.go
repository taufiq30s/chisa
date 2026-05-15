package quiz

import (
	"sort"
	"sync"
	"time"
)

// QueueEntry holds information about a user who has raised their hand.
type QueueEntry struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Avatar    string    `json:"avatar"`
	Timestamp time.Time `json:"timestamp"`
}

// Session holds the live state of one quiz session.
// All exported methods are safe for concurrent use.
type Session struct {
	mu            sync.Mutex
	Active        bool
	ChannelID     string
	GuildID       string
	Queue         []QueueEntry
	Promoted      map[string]bool
	TotalPromoted int
}

func newSession() *Session {
	return &Session{
		Promoted: make(map[string]bool),
	}
}

// AddToQueue inserts entry into the queue if the user is not already present.
// Entries are kept sorted by Timestamp ascending (earliest hand raise first).
func (s *Session) AddToQueue(entry QueueEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.Queue {
		if e.UserID == entry.UserID {
			return // already queued
		}
	}
	s.Queue = append(s.Queue, entry)
	sort.Slice(s.Queue, func(i, j int) bool {
		return s.Queue[i].Timestamp.Before(s.Queue[j].Timestamp)
	})
}

// RemoveFromQueue removes the user with the given ID from the queue (no-op if absent).
func (s *Session) RemoveFromQueue(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, e := range s.Queue {
		if e.UserID == userID {
			s.Queue = append(s.Queue[:i], s.Queue[i+1:]...)
			return
		}
	}
}

// MarkPromoted records a user as promoted and removes them from the queue.
func (s *Session) MarkPromoted(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Promoted[userID] = true
	s.TotalPromoted++
	for i, e := range s.Queue {
		if e.UserID == userID {
			s.Queue = append(s.Queue[:i], s.Queue[i+1:]...)
			return
		}
	}
}

// IsPromoted reports whether the user has already been promoted this session.
func (s *Session) IsPromoted(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Promoted[userID]
}

// Reset wipes all session state (does not change Active flag).
func (s *Session) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Queue = nil
	s.Promoted = make(map[string]bool)
	s.TotalPromoted = 0
	s.ChannelID = ""
	s.GuildID = ""
}

// Snapshot returns a defensive copy of the current queue.
func (s *Session) Snapshot() []QueueEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Queue) == 0 {
		return []QueueEntry{}
	}
	cp := make([]QueueEntry, len(s.Queue))
	copy(cp, s.Queue)
	return cp
}
