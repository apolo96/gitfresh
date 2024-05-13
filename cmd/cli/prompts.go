package main

import (
	"bufio"
	"os"
	"strings"
)

/* CLI Prompts */
func PromptSecret(label string, required bool) (s string) {
	println("[Press ENTER to submit your response] \n")
	if !required {
		label = "[OPTIONAL] " + label
	}
	println("- " + label)
	r := bufio.NewReader(os.Stdin)
	for {
		print("> ")
		s, _ = r.ReadString('\n')
		if required && len(s) <= 1 {
			println("Empty value, please type a real value")
			continue
		}
		break

	}
	println("")
	return strings.TrimRight(s, "\n")
}

func PromptConfirm(label string, value string) bool {
	println("[Press ENTER to submit your response] \n")
	println("- " + label + " GitWorkDir=" + value)
	r := bufio.NewReader(os.Stdin)
	for {
		print("> ")
		s, _ := r.ReadString('\n')
		s = strings.ToLower(strings.TrimSpace(s))
		if len(s) <= 1 {
			println("Empty value, please type a real value")
			continue
		}
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
		break
	}
	return true
}
