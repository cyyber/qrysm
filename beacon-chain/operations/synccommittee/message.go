package synccommittee

import (
	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/container/queue"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

// SaveSyncCommitteeMessage saves a sync committee message in to a priority queue.
// The priority queue capped at syncCommitteeMaxQueueSize contributions.
func (s *Store) SaveSyncCommitteeMessage(msg *qrysmpb.SyncCommitteeMessage) error {
	if msg == nil {
		return errNilMessage
	}

	s.messageLock.Lock()
	defer s.messageLock.Unlock()

	item, err := s.messageCache.PopByKey(syncCommitteeKey(msg.Slot))
	if err != nil {
		return err
	}

	copied := qrysmpb.CopySyncCommitteeMessage(msg)
	// Messages exist in the queue. Append instead of insert new.
	if item != nil {
		messages, ok := item.Value.([]*qrysmpb.SyncCommitteeMessage)
		if !ok {
			return errors.New("not typed []qrysmpb.SyncCommitteeMessage")
		}

		idx := -1
		for i, msg := range messages {
			if msg.ValidatorIndex == copied.ValidatorIndex {
				idx = i
				break
			}
		}
		if idx >= 0 {
			// Override the existing messages with a new one
			messages[idx] = copied
		} else {
			// Append the new message
			messages = append(messages, copied)
			savedSyncCommitteeMessageTotal.Inc()
		}

		return s.messageCache.Push(&queue.Item{
			Key:      syncCommitteeKey(msg.Slot),
			Value:    messages,
			Priority: int64(msg.Slot),
		})
	}

	// Message does not exist. Insert new.
	if err := s.messageCache.Push(&queue.Item{
		Key:      syncCommitteeKey(msg.Slot),
		Value:    []*qrysmpb.SyncCommitteeMessage{copied},
		Priority: int64(msg.Slot),
	}); err != nil {
		return err
	}
	savedSyncCommitteeMessageTotal.Inc()

	// Trim messages in queue down to syncCommitteeMaxQueueSize.
	if s.messageCache.Len() > syncCommitteeMaxQueueSize {
		if _, err := s.messageCache.Pop(); err != nil {
			return err
		}
	}

	return nil
}

// SyncCommitteeMessages returns sync committee messages by slot from the priority queue.
// Upon retrieval, the message is removed from the queue.
func (s *Store) SyncCommitteeMessages(slot primitives.Slot) ([]*qrysmpb.SyncCommitteeMessage, error) {
	s.messageLock.RLock()
	defer s.messageLock.RUnlock()

	item := s.messageCache.RetrieveByKey(syncCommitteeKey(slot))
	if item == nil {
		return nil, nil
	}

	messages, ok := item.Value.([]*qrysmpb.SyncCommitteeMessage)
	if !ok {
		return nil, errors.New("not typed []qrysmpb.SyncCommitteeMessage")
	}

	return messages, nil
}
