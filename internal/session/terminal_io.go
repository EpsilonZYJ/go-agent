// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package session

import (
	"bufio"
	"os"
	"sync"
)

var (
	termScanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	termMu      sync.Mutex
)

func ReadLine() (string, error) {
	termMu.Lock()
	defer termMu.Unlock()
	termScanner.Scan()
	return termScanner.Text(), termScanner.Err()
}
