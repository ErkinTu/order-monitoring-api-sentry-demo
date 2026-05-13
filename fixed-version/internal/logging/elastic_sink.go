package logging

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type elasticSink struct {
	client      *elasticsearch.Client
	indexPrefix string
	timeout     time.Duration
	queue       chan []byte
	dropped     atomic.Uint64
	wg          sync.WaitGroup
	once        sync.Once
}

func newElasticSink(opts Options) (*elasticSink, error) {
	if err := validateElasticOptions(opts); err != nil {
		return nil, err
	}

	queueSize := opts.ElasticQueueSize
	if queueSize <= 0 {
		queueSize = 2000
	}

	timeout := time.Duration(opts.ElasticTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	cfg := elasticsearch.Config{
		CloudID:  strings.TrimSpace(opts.ElasticCloudID),
		APIKey:   normalizeAPIKey(strings.TrimSpace(opts.ElasticAPIKey)),
		Username: strings.TrimSpace(opts.ElasticUsername),
		Password: strings.TrimSpace(opts.ElasticPassword),
	}

	if url := strings.TrimSpace(opts.ElasticURL); url != "" {
		cfg.Addresses = []string{url}
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create elastic client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	infoRes, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("elastic info request failed: %w", err)
	}
	defer infoRes.Body.Close()

	if infoRes.IsError() {
		body, _ := io.ReadAll(io.LimitReader(infoRes.Body, 2048))
		return nil, fmt.Errorf("elastic info status=%s body=%s", infoRes.Status(), strings.TrimSpace(string(body)))
	}

	sink := &elasticSink{
		client:      client,
		indexPrefix: strings.TrimSpace(opts.ElasticIndexPrefix),
		timeout:     timeout,
		queue:       make(chan []byte, queueSize),
	}
	if sink.indexPrefix == "" {
		sink.indexPrefix = "order-monitoring-logs"
	}

	sink.wg.Add(1)
	go sink.worker()

	slog.Info("elastic logging sink enabled", "elastic.index_prefix", sink.indexPrefix, "elastic.queue_size", queueSize)
	return sink, nil
}

func (s *elasticSink) Write(p []byte) (int, error) {
	line := bytes.TrimSpace(p)
	if len(line) == 0 {
		return len(p), nil
	}

	cp := make([]byte, len(line))
	copy(cp, line)

	select {
	case s.queue <- cp:
	default:
		dropped := s.dropped.Add(1)
		if dropped == 1 || dropped%100 == 0 {
			slog.Warn("elastic log queue is full, dropping log records", "dropped.total", dropped)
		}
	}

	return len(p), nil
}

func (s *elasticSink) Close(ctx context.Context) error {
	s.once.Do(func() {
		close(s.queue)
	})

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *elasticSink) worker() {
	defer s.wg.Done()

	for payload := range s.queue {
		s.writeToElastic(payload)
	}
}

func (s *elasticSink) writeToElastic(payload []byte) {
	index := fmt.Sprintf("%s-%s", s.indexPrefix, time.Now().UTC().Format("2006.01.02"))
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	res, err := s.client.Index(
		index,
		bytes.NewReader(payload),
		s.client.Index.WithContext(ctx),
	)
	if err != nil {
		slog.Warn("failed to send log to elastic", "error", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		slog.Warn("elastic index request failed", "status", res.Status(), "body", strings.TrimSpace(string(body)))
	}
}

func normalizeAPIKey(value string) string {
	if value == "" {
		return ""
	}
	if strings.Contains(value, ":") {
		return base64.StdEncoding.EncodeToString([]byte(value))
	}
	return value
}
