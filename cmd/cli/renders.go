package main

import (
	"fmt"

	"github.com/apolo96/gitfresh"
)

var Verbose bool = true

func renderVerbose(text string) {
	if Verbose {
		println(text)
	}
}

func renderRepos(rp []*gitfresh.GitRepository, fresh bool) {
	for _, r := range rp {
		url := fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Name)
		if fresh {
			url = url + "/settings/hooks"
		}
		fmt.Printf("Repository: %-25s | URL: %-20s\n", r.Name, url)
	}
}
