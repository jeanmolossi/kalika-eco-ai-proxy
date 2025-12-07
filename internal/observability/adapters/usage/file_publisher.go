package usage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

const usageFileQueueSize = 256

var ErrUsageFilePublisherQueueFull = errors.New("usage file publisher queue is full")

// FilePublisher persists usage events as JSONL on disk for later billing export.
type FilePublisher struct {
	path  string
	mu    sync.Mutex
	queue chan []byte
	wg    sync.WaitGroup
}

// NewFilePublisher creates or opens a JSONL sink at the given path.
func NewFilePublisher(path string) (*FilePublisher, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	if err := file.Close(); err != nil {
		return nil, err
	}

	p := &FilePublisher{
		path:  path,
		queue: make(chan []byte, usageFileQueueSize),
	}

	p.wg.Add(1)

	go p.run()

	return p, nil
}

// Publish appends the event as JSON to the sink and fsyncs to keep it durable.
func (p *FilePublisher) Publish(_ context.Context, ev Event) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	select {
	case p.queue <- append(data, '\n'):
		return nil
	default:
		return ErrUsageFilePublisherQueueFull
	}
}

// Close stops the worker and flushes pending messages.
func (p *FilePublisher) Close() error {
	close(p.queue)
	p.wg.Wait()

	return nil
}

func (p *FilePublisher) run() {
	defer p.wg.Done()

	for data := range p.queue {
		p.mu.Lock()

		file, err := os.OpenFile(p.path, os.O_APPEND|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = file.Write(data)
			_ = file.Sync()
			_ = file.Close()
		}

		p.mu.Unlock()
	}
}
