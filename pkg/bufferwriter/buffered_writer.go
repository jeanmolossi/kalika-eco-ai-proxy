package bufferwriter

import (
	"bufio"
	"io"
	"sync"
	"time"
)

// BufferedWriter is a thread-safe buffered writer for logs.
// It uses a single goroutine to serialize all writes and buffer data
// before flushing to the underlying writer.
type BufferedWriter struct {
	ch   chan []byte
	done chan struct{}

	wg    sync.WaitGroup
	errMu sync.Mutex
	err   error
}

type BufferedWriterOptions struct {
	ChannelSize   int           // number of queued messages
	BufferSize    int           // bufio size
	FlushInterval time.Duration // periodic flush
}

func NewBufferedWriter(dst io.Writer, opt BufferedWriterOptions) *BufferedWriter {
	if opt.ChannelSize <= 0 {
		opt.ChannelSize = 1024
	}

	if opt.BufferSize <= 0 {
		const bufSize = 64 << 10

		opt.BufferSize = bufSize
	}

	if opt.FlushInterval <= 0 {
		opt.FlushInterval = time.Second
	}

	//nolint:varnamelen // short name used just at this scope
	bw := &BufferedWriter{
		ch:   make(chan []byte, opt.ChannelSize),
		done: make(chan struct{}),
	}

	bw.wg.Go(func() {
		bufw := bufio.NewWriterSize(dst, opt.BufferSize)

		ticker := time.NewTicker(opt.FlushInterval)
		defer ticker.Stop()

		flush := func() {
			if err := bufw.Flush(); err != nil {
				bw.setErr(err)
			}
		}

		for {
			select {
			case data, ok := <-bw.ch:
				if !ok {
					// draing ended, flush and exit
					flush()
					return
				}

				if len(data) > bufw.Available() {
					// flush before write to avoid break logs
					flush()
				}

				if _, err := bufw.Write(data); err != nil {
					bw.setErr(err)
				}
			case <-ticker.C:
				flush()
			}
		}
	})

	return bw
}

func (bw *BufferedWriter) setErr(err error) {
	if err == nil {
		return
	}

	bw.errMu.Lock()
	defer bw.errMu.Unlock()

	if bw.err == nil {
		bw.err = err
	}
}

// Write is thread-safe. It copies the slice and enqueues it.
func (bw *BufferedWriter) Write(p []byte) (int, error) {
	// copy to avoid data races if caller reuses buffer
	cp := make([]byte, len(p))
	copy(cp, p)

	select {
	case bw.ch <- cp:
		return len(p), nil
	case <-bw.done:
		// writer is closing/closed
		return 0, io.ErrClosedPipe
	}
}

// Close waits for all logs to be flushed.
func (bw *BufferedWriter) Close() error {
	close(bw.ch)
	close(bw.done)
	bw.wg.Wait()

	bw.errMu.Lock()
	defer bw.errMu.Unlock()

	return bw.err
}
