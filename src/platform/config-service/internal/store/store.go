package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/shopos/config-service/internal/proto"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Store wraps etcd client with the config-service domain operations.
type Store struct {
	client  *clientv3.Client
	prefix  string
	timeout time.Duration
}

func New(addrs []string, prefix string, timeout time.Duration) (*Store, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: addrs,
		// DialTimeout omitted: lazy connection so startup succeeds without etcd.
	})
	if err != nil {
		return nil, fmt.Errorf("etcd connect: %w", err)
	}
	return &Store{client: cli, prefix: prefix, timeout: timeout}, nil
}

func (s *Store) Close() error { return s.client.Close() }

func (s *Store) key(k string) string {
	return s.prefix + "/" + strings.TrimPrefix(k, "/")
}

func (s *Store) Get(ctx context.Context, key string) (*pb.ConfigEntry, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resp, err := s.client.Get(ctx, s.key(key))
	if err != nil {
		return nil, false, err
	}
	if len(resp.Kvs) == 0 {
		return nil, false, nil
	}
	kv := resp.Kvs[0]
	return &pb.ConfigEntry{
		Key:     key,
		Value:   string(kv.Value),
		Version: kv.Version,
	}, true, nil
}

func (s *Store) Set(ctx context.Context, key, value string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.client.Put(ctx, s.key(key), value)
	return err
}

func (s *Store) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.client.Delete(ctx, s.key(key))
	return err
}

func (s *Store) List(ctx context.Context, prefix string) ([]*pb.ConfigEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	etcdPrefix := s.key(prefix)
	resp, err := s.client.Get(ctx, etcdPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	entries := make([]*pb.ConfigEntry, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		k := strings.TrimPrefix(string(kv.Key), s.prefix+"/")
		entries = append(entries, &pb.ConfigEntry{
			Key:     k,
			Value:   string(kv.Value),
			Version: kv.Version,
		})
	}
	return entries, nil
}

// WatchCh opens a watch channel on a key prefix and sends events until ctx is cancelled.
func (s *Store) WatchCh(ctx context.Context, prefix string) (<-chan *pb.WatchEvent, error) {
	etcdPrefix := s.key(prefix)
	watchCh := s.client.Watch(ctx, etcdPrefix, clientv3.WithPrefix())
	out := make(chan *pb.WatchEvent, 32)

	go func() {
		defer close(out)
		for resp := range watchCh {
			for _, ev := range resp.Events {
				k := strings.TrimPrefix(string(ev.Kv.Key), s.prefix+"/")
				out <- &pb.WatchEvent{
					Entry: &pb.ConfigEntry{
						Key:     k,
						Value:   string(ev.Kv.Value),
						Version: ev.Kv.Version,
					},
					Deleted: ev.Type == clientv3.EventTypeDelete,
				}
			}
		}
	}()
	return out, nil
}
