package store

import (
	"context"
	ipb "engram/internal/proto"
	pb "engram/proto/v1"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

func (r *RedisStore) eventKey(agentID string) string {
	return fmt.Sprintf("agent:%s:events", agentID)
}

func (r *RedisStore) metaKey(agentID string) string {
	return fmt.Sprintf("agent:%s:meta", agentID)
}

func (r *RedisStore) workerKey(workerID string) string {
	return fmt.Sprintf("worker:%s:heartbeat", workerID)
}

func (r *RedisStore) idempotencyKey(key string) string {
	return fmt.Sprintf("idempotency:%s", key)
}

// === EventStore ===

func (r *RedisStore) AppendEvent(ctx context.Context, agentID string, event *ipb.Event) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	_, err = r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: r.eventKey(agentID),
		Values: map[string]interface{}{
			"seq":  event.Sequence,
			"data": data,
		},
	}).Result()

	return err
}

func (r *RedisStore) LoadEvents(ctx context.Context, agentID string, fromSeq uint64) ([]*ipb.Event, error) {
	msgs, err := r.client.XRange(ctx, r.eventKey(agentID), "-", "+").Result()
	if err != nil {
		return nil, err
	}

	var events []*ipb.Event
	for _, msg := range msgs {
		data, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		event := &ipb.Event{}
		if err := proto.Unmarshal([]byte(data), event); err != nil {
			continue
		}

		if event.Sequence >= fromSeq {
			events = append(events, event)
		}
	}

	return events, nil
}

func (r *RedisStore) GetNextSequence(ctx context.Context, agentID string) (uint64, error) {
	length, err := r.client.XLen(ctx, r.eventKey(agentID)).Result()
	if err != nil {
		return 1, nil
	}
	return uint64(length) + 1, nil
}

// === MetadataStore ===

func (r *RedisStore) GetMetadata(ctx context.Context, agentID string) (*ipb.AgentMetadata, error) {
	data, err := r.client.Get(ctx, r.metaKey(agentID)).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("agent not found")
	}
	if err != nil {
		return nil, err
	}

	meta := &ipb.AgentMetadata{}
	if err := proto.Unmarshal(data, meta); err != nil {
		return nil, err
	}

	return meta, nil
}

func (r *RedisStore) SetMetadata(ctx context.Context, agentID string, meta *ipb.AgentMetadata) error {
	meta.UpdatedAt = timestamppb.Now()
	data, err := proto.Marshal(meta)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, r.metaKey(agentID), data, 0)
	if meta.Status == "waiting_for_step" {
		pipe.SAdd(ctx, "agents:pending", agentID)
	} else {
		pipe.SRem(ctx, "agents:pending", agentID)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStore) GetAgent(ctx context.Context, agentID string) (*pb.Agent, error) {
	data, err := r.client.Get(ctx, "agent:v1:"+agentID).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("agent not found")
	}
	if err != nil {
		return nil, err
	}

	agent := &pb.Agent{}
	if err := proto.Unmarshal(data, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

func (r *RedisStore) SetAgent(ctx context.Context, agentID string, agent *pb.Agent) error {
	agent.UpdatedAt = timestamppb.Now()
	data, err := proto.Marshal(agent)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, "agent:v1:"+agentID, data, 0).Err()
}

func (r *RedisStore) DeleteAgent(ctx context.Context, agentID string) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, r.metaKey(agentID))
	pipe.Del(ctx, "agent:v1:"+agentID)
	pipe.Del(ctx, r.eventKey(agentID))
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStore) AtomicClaim(ctx context.Context, agentID, workerID string) (*ipb.AgentMetadata, error) {
	// Use WATCH/MULTI/EXEC for atomic claim
	key := r.metaKey(agentID)

	txf := func(tx *redis.Tx) error {
		data, err := tx.Get(ctx, key).Bytes()
		if err != nil {
			return err
		}

		meta := &ipb.AgentMetadata{}
		if err := proto.Unmarshal(data, meta); err != nil {
			return err
		}

		if meta.Status != "waiting_for_step" || meta.RunningOnWorker != "" {
			return fmt.Errorf("cannot claim: status=%s, worker=%s", meta.Status, meta.RunningOnWorker)
		}

		meta.Status = "processing"
		meta.RunningOnWorker = workerID
		meta.StepStartedAt = timestamppb.Now()
		meta.UpdatedAt = timestamppb.Now()

		newData, err := proto.Marshal(meta)
		if err != nil {
			return err
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, newData, 0)
			pipe.SRem(ctx, "agents:pending", agentID)
			return nil
		})

		return err
	}

	for i := 0; i < 3; i++ {
		err := r.client.Watch(ctx, txf, key)
		if err == nil {
			return r.GetMetadata(ctx, agentID)
		}
		if err == redis.TxFailedErr {
			continue
		}
		return nil, err
	}

	return nil, fmt.Errorf("failed to claim after retries")
}

func (r *RedisStore) UpdateStatus(ctx context.Context, agentID, status string) error {
	meta, err := r.GetMetadata(ctx, agentID)
	if err != nil {
		return err
	}

	meta.Status = status
	if status != "processing" {
		meta.RunningOnWorker = ""
	}

	pipe := r.client.Pipeline()
	if status == "waiting_for_step" {
		pipe.SAdd(ctx, "agents:pending", agentID)
	} else {
		pipe.SRem(ctx, "agents:pending", agentID)
	}

	data, err := proto.Marshal(meta)
	if err != nil {
		return err
	}
	pipe.Set(ctx, r.metaKey(agentID), data, 0)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStore) ListPendingAgents(ctx context.Context) ([]string, error) {
	return r.client.SMembers(ctx, "agents:pending").Result()
}

// === IdempotencyStore ===

func (r *RedisStore) GetToolResult(ctx context.Context, key string) (string, bool, error) {
	result, err := r.client.Get(ctx, r.idempotencyKey(key)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return result, true, nil
}

func (r *RedisStore) SetToolResult(ctx context.Context, key, result string) error {
	return r.client.Set(ctx, r.idempotencyKey(key), result, 24*time.Hour).Err()
}

// === WorkerRegistry ===

func (r *RedisStore) RegisterWorker(ctx context.Context, workerID string) error {
	return r.client.SAdd(ctx, "workers:active", workerID).Err()
}

func (r *RedisStore) Heartbeat(ctx context.Context, workerID string) error {
	now := time.Now().Unix()
	pipe := r.client.Pipeline()
	pipe.Set(ctx, r.workerKey(workerID), now, 30*time.Second)
	pipe.SAdd(ctx, "workers:active", workerID)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStore) GetDeadWorkers(ctx context.Context, timeout int64) ([]string, error) {
	workers, err := r.client.SMembers(ctx, "workers:active").Result()
	if err != nil {
		return nil, err
	}

	var dead []string
	now := time.Now().Unix()

	for _, workerID := range workers {
		lastBeat, err := r.client.Get(ctx, r.workerKey(workerID)).Result()
		if err == redis.Nil {
			dead = append(dead, workerID)
			continue
		}
		if err != nil {
			continue
		}

		ts, _ := strconv.ParseInt(lastBeat, 10, 64)
		if now-ts > timeout {
			dead = append(dead, workerID)
		}
	}

	return dead, nil
}

func (r *RedisStore) UnregisterWorker(ctx context.Context, workerID string) error {
	pipe := r.client.Pipeline()
	pipe.SRem(ctx, "workers:active", workerID)
	pipe.Del(ctx, r.workerKey(workerID))
	_, err := pipe.Exec(ctx)
	return err
}

// === Subscription for real-time streaming ===

func (r *RedisStore) SubscribeEvents(ctx context.Context, agentID string, handler func(*ipb.Event)) error {
	lastID := "0"

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		streams, err := r.client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{r.eventKey(agentID), lastID},
			Count:   10,
			Block:   1 * time.Second,
		}).Result()

		if err == redis.Nil {
			continue
		}
		if err != nil {
			return err
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				lastID = msg.ID
				data, ok := msg.Values["data"].(string)
				if !ok {
					continue
				}

				event := &ipb.Event{}
				if err := proto.Unmarshal([]byte(data), event); err != nil {
					continue
				}

				handler(event)
			}
		}
	}
}

// === Utility ===

func (r *RedisStore) Close() error {
	return r.client.Close()
}

func (r *RedisStore) ListAgentIDs(ctx context.Context) ([]string, error) {
	keys, err := r.client.Keys(ctx, "agent:*:meta").Result()
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) >= 2 {
			ids = append(ids, parts[1])
		}
	}

	return ids, nil
}
