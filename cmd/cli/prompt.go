package main

import (
	"bufio"
	"os"
	"strings"
)

/* CLI Prompts */
func PromptSecret(label string, required bool) (e string) {
	println("[Press ENTER to submit your response] \n")
	if !required {
		label = "[OPTIONAL] " + label
	}
	println("- " + label)
	r := bufio.NewReader(os.Stdin)
	for {
		print("> ")
		e, _ = r.ReadString('\n')
		if required && len(e) <= 1 {
			println("Entered an empty value, please type a real value")
			continue
		}
		break

	}
	println("")
	return strings.TrimRight(e, "\n")
}
