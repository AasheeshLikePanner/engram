package store

import (
	"context"
	ipb "engram/internal/proto"
	pb "engram/proto/v1"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	bpb "github.com/dgraph-io/badger/v4/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BadgerStore struct {
	db *badger.DB
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db at %s: %w", path, err)
	}

	return &BadgerStore{db: db}, nil
}

func (s *BadgerStore) eventKey(agentID string, seq uint64) string {
	return fmt.Sprintf("agent:v1:%s:event:%08d", agentID, seq)
}

func (s *BadgerStore) metaKey(agentID string) string {
	return fmt.Sprintf("agent:v1:%s:meta", agentID)
}

func (s *BadgerStore) workerKey(workerID string) string {
	return fmt.Sprintf("worker:%s:heartbeat", workerID)
}

func (s *BadgerStore) idempotencyKey(key string) string {
	return fmt.Sprintf("idempotency:%s", key)
}

// === EventStore ===

func (s *BadgerStore) AppendEvent(ctx context.Context, agentID string, event *ipb.Event) error {
	key := s.eventKey(agentID, event.Sequence)
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func (s *BadgerStore) LoadEvents(ctx context.Context, agentID string, fromSeq uint64) ([]*ipb.Event, error) {
	var events []*ipb.Event
	prefix := []byte(fmt.Sprintf("agent:v1:%s:event:", agentID))

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				event := &ipb.Event{}
				if err := proto.Unmarshal(val, event); err != nil {
					return err
				}
				if event.Sequence >= fromSeq {
					events = append(events, event)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return events, err
}

func (s *BadgerStore) SubscribeEvents(ctx context.Context, agentID string, handler func(*ipb.Event)) error {
	prefix := []byte(fmt.Sprintf("agent:v1:%s:event:", agentID))
	return s.db.Subscribe(ctx, func(kvs *badger.KVList) error {
		for _, kv := range kvs.Kv {
			event := &ipb.Event{}
			if err := proto.Unmarshal(kv.Value, event); err != nil {
				continue
			}
			handler(event)
		}
		return nil
	}, []bpb.Match{{Prefix: prefix}})
}

func (s *BadgerStore) GetNextSequence(ctx context.Context, agentID string) (uint64, error) {
	meta, err := s.GetMetadata(ctx, agentID)
	if err == nil {
		return meta.HeadSequence + 1, nil
	}

	var count uint64 = 0
	prefix := []byte(fmt.Sprintf("agent:v1:%s:event:", agentID))

	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			count++
		}
		return nil
	})

	return count + 1, nil
}

// === MetadataStore ===

func (s *BadgerStore) GetMetadata(ctx context.Context, agentID string) (*ipb.AgentMetadata, error) {
	meta := &ipb.AgentMetadata{}
	key := s.metaKey(agentID)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return proto.Unmarshal(val, meta)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, fmt.Errorf("agent not found")
	}

	return meta, err
}

func (s *BadgerStore) SetMetadata(ctx context.Context, agentID string, meta *ipb.AgentMetadata) error {
	meta.UpdatedAt = timestamppb.Now()
	key := s.metaKey(agentID)
	data, err := proto.Marshal(meta)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func (s *BadgerStore) GetAgent(ctx context.Context, agentID string) (*pb.Agent, error) {
	agent := &pb.Agent{}
	key := fmt.Sprintf("agent:v1:%s", agentID)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return proto.Unmarshal(val, agent)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, fmt.Errorf("agent not found")
	}

	return agent, err
}

func (s *BadgerStore) SetAgent(ctx context.Context, agentID string, agent *pb.Agent) error {
	agent.UpdatedAt = timestamppb.Now()
	key := fmt.Sprintf("agent:v1:%s", agentID)
	data, err := proto.Marshal(agent)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func (s *BadgerStore) DeleteAgent(ctx context.Context, agentID string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		txn.Delete([]byte(s.metaKey(agentID)))
		txn.Delete([]byte(fmt.Sprintf("agent:v1:%s", agentID)))

		// Delete events
		prefix := []byte(fmt.Sprintf("agent:v1:%s:event:", agentID))
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			txn.Delete(it.Item().KeyCopy(nil))
		}
		return nil
	})
}

func (s *BadgerStore) AtomicClaim(ctx context.Context, agentID, workerID string) (*ipb.AgentMetadata, error) {
	key := []byte(s.metaKey(agentID))
	var claimedMeta *ipb.AgentMetadata

	err := s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		meta := &ipb.AgentMetadata{}
		err = item.Value(func(val []byte) error {
			return proto.Unmarshal(val, meta)
		})
		if err != nil {
			return err
		}

		if meta.Status != "waiting_for_step" || meta.RunningOnWorker != "" {
			return fmt.Errorf("cannot claim: status=%s, worker=%s", meta.Status, meta.RunningOnWorker)
		}

		meta.Status = "processing"
		meta.RunningOnWorker = workerID
		meta.StepStartedAt = timestamppb.Now()
		meta.UpdatedAt = timestamppb.Now()

		data, err := proto.Marshal(meta)
		if err != nil {
			return err
		}

		claimedMeta = meta
		return txn.Set(key, data)
	})

	return claimedMeta, err
}

func (s *BadgerStore) UpdateStatus(ctx context.Context, agentID, status string) error {
	meta, err := s.GetMetadata(ctx, agentID)
	if err != nil {
		return err
	}

	meta.Status = status
	if status != "processing" {
		meta.RunningOnWorker = ""
	}

	return s.SetMetadata(ctx, agentID, meta)
}

// === IdempotencyStore ===

func (s *BadgerStore) GetToolResult(ctx context.Context, key string) (string, bool, error) {
	var result string
	found := false

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(s.idempotencyKey(key)))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			result = string(val)
			found = true
			return nil
		})
	})

	return result, found, err
}

func (s *BadgerStore) SetToolResult(ctx context.Context, key, result string) error {
	entry := badger.NewEntry([]byte(s.idempotencyKey(key)), []byte(result)).WithTTL(24 * time.Hour)
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(entry)
	})
}

// === WorkerRegistry ===

func (s *BadgerStore) RegisterWorker(ctx context.Context, workerID string) error {
	return s.Heartbeat(ctx, workerID)
}

func (s *BadgerStore) Heartbeat(ctx context.Context, workerID string) error {
	key := s.workerKey(workerID)
	now := fmt.Sprintf("%d", time.Now().Unix())
	entry := badger.NewEntry([]byte(key), []byte(now)).WithTTL(30 * time.Second)

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(entry)
	})
}

func (s *BadgerStore) GetDeadWorkers(ctx context.Context, timeout int64) ([]string, error) {
	var dead []string
	prefix := []byte("worker:")
	now := time.Now().Unix()

	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.Key())

			if !strings.HasSuffix(key, ":heartbeat") {
				continue
			}

			err := item.Value(func(val []byte) error {
				ts := int64(0)
				fmt.Sscanf(string(val), "%d", &ts)
				if now-ts > timeout {
					parts := strings.Split(key, ":")
					if len(parts) >= 2 {
						dead = append(dead, parts[1])
					}
				}
				return nil
			})
			if err != nil {
				continue
			}
		}
		return nil
	})

	return dead, nil
}

func (s *BadgerStore) UnregisterWorker(ctx context.Context, workerID string) error {
	key := s.workerKey(workerID)
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// === Utility ===

func (s *BadgerStore) ListAgentIDs(ctx context.Context) ([]string, error) {
	var ids []string
	prefix := []byte("agent:v1:")

	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		seen := make(map[string]bool)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := string(it.Item().Key())
			parts := strings.Split(key, ":")
			if len(parts) >= 3 {
				agentID := parts[2]
				if !seen[agentID] {
					seen[agentID] = true
					ids = append(ids, agentID)
				}
			}
		}
		return nil
	})

	return ids, nil
}

func (s *BadgerStore) Close() error {
	return s.db.Close()
}

func (s *BadgerStore) DB() *badger.DB {
	return s.db
}
