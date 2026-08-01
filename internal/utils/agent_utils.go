// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package utils

import (
	"fmt"
	"strings"
)

// PrintAgentOutput prints each non-empty agent output to stdout, prefixed
// with a green "Agent:" label, followed by a trailing blank line.
func PrintAgentOutput(textOuts []strings.Builder) {
	for _, textOut := range textOuts {
		if textOut.Len() > 0 {
			fmt.Println("\033[32mAgent: \n\n \033[0m" + textOut.String())
		}
	}
	fmt.Println()
}
