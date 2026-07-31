// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package session

import (
	"bufio"
	"os"
	"sync"

	"go-agent/internal/logs"
)

var (
	termScanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	termMu      sync.Mutex
)

func ReadLine() (string, error) {
	termMu.Lock()
	defer termMu.Unlock()
	termScanner.Scan()
	if err := termScanner.Err(); err != nil {
		logs.Warn("terminal input scan error", "err", err)
	}
	return termScanner.Text(), termScanner.Err()
}
