package eventlog

import (
	"context"
	"log"
	"os"
	"sync"
	"time"
)

// Logger — асинхронный логгер событий.
// Публичный API: Log(msg) — не блокирует; Close(ctx) — дожидается слива.
type Logger struct {
	ch        chan string
	done      chan struct{} // закрывается, когда воркер завершился
	closeOnce sync.Once
}

// New создаёт логгер и запускает воркер.
func New(path string, delay time.Duration) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	l := &Logger{
		ch:   make(chan string, 1000),
		done: make(chan struct{}),
	}

	go l.worker(f, delay)
	return l, nil
}

// Log отправляет сообщение в канал. Не блокирует (пока буфер не переполнен).
func (l *Logger) Log(msg string) {
	defer func() {
		_ = recover()
	}()
	l.ch <- msg
}

// Close сообщает воркеру завершиться и ждёт, пока он допишет всё из буфера.
func (l *Logger) Close(ctx context.Context) {
	l.closeOnce.Do(func() {
		close(l.ch)
	})

	select {
	case <-l.done:
	case <-ctx.Done():
		log.Printf("eventlog: close timeout: %v", ctx.Err())
	}
}

// worker — единственный, кто пишет в файл.
func (l *Logger) worker(f *os.File, delay time.Duration) {
	defer close(l.done)
	defer f.Close()

	// range по каналу завершится, когда канал закроют (после Close).
	for msg := range l.ch {
		time.Sleep(delay)

		line := time.Now().Format(time.RFC3339) + " " + msg + "\n"
		if _, err := f.WriteString(line); err != nil {
			log.Printf("eventlog: write error: %v", err)
			continue
		}

		log.Printf("eventlog: written: %s", msg)
	}
	log.Println("eventlog: worker stopped")
}
