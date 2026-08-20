// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package memory

import (
	"go-agent/internal/logs"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
)

var memFileMu sync.Mutex

var (
	memTaskCh     = make(chan string, 8)
	memWorkerOnce sync.Once
	memWg         sync.WaitGroup
)

// Background workers never write to the terminal directly; they only append
// messages to notices. The main goroutine drains and prints them right before
// showing the next prompt, so async output can't interrupt the line the user
// is currently typing.
var (
	noticeMu sync.Mutex
	notices  []string
)

// addNotice records a notice in the background without printing it.
func addNotice(s string) {
	noticeMu.Lock()
	notices = append(notices, s)
	noticeMu.Unlock()
}

// DrainNotices returns and clears pending notices; the main goroutine calls
// it right before printing the next prompt.
func DrainNotices() []string {
	noticeMu.Lock()
	defer noticeMu.Unlock()
	out := notices
	notices = nil
	return out
}

// StartMemoryWorker launches (once) the background worker that consumes
// dialogues from memTaskCh, extracting and consolidating memories.
func StartMemoryWorker() {
	logs.Info("Starting memory worker...")
	memWorkerOnce.Do(func() {
		go func() {
			for dialogue := range memTaskCh {
				memFileMu.Lock()
				extractFromDialogue(dialogue)
				ConsolidateMemory()
				memFileMu.Unlock()
				memWg.Done()
			}
		}()
	})
}

// EnqueueExtraction queues a conversation snapshot for background memory
// extraction. It is a no-op when the snapshot renders to an empty dialogue,
// and silently drops the task when the queue is full.
func EnqueueExtraction(snapshot []anthropic.MessageParam) {
	dialogue := buildDialogue(snapshot)
	if dialogue == "" {
		return
	}
	memWg.Add(1)
	select {
	case memTaskCh <- dialogue:
	default:
		memWg.Done()
	}
}

// FlushMemories blocks until all queued extraction tasks have finished.
func FlushMemories() {
	memWg.Wait()
}
