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

func FlushMemories() {
	memWg.Wait()
}
