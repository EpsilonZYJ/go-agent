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

// 后台 worker 不直接写终端，只把提示存入 notices；由主 goroutine 在打印下一个
// 提示符前统一取出打印，避免异步输出打断用户正在输入的行。
var (
	noticeMu sync.Mutex
	notices  []string
)

// addNotice 后台只记录，不打印。
func addNotice(s string) {
	noticeMu.Lock()
	notices = append(notices, s)
	noticeMu.Unlock()
}

// DrainNotices 取出并清空待显示通知；主 goroutine 在打印下一个提示符前调用。
func DrainNotices() []string {
	noticeMu.Lock()
	defer noticeMu.Unlock()
	out := notices
	notices = nil
	return out
}

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
