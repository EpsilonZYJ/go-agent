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

func ReadLine() bool {
	termMu.Lock()
	defer termMu.Unlock()
	return termScanner.Scan()
}

func ScannerErr() error {
	return termScanner.Err()
}

func LineText() string {
	return termScanner.Text()
}
