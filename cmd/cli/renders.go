package main

import (
	"fmt"

	"github.com/apolo96/gitfresh"
)

func renderVerbose(text string) {
	println(text)
}

func renderRepos(rp []*gitfresh.Repository) {
	for _, r := range rp {
		url := fmt.Sprintf("https://github.com/apolo96/%s/settings/hooks", r.Name)
		fmt.Printf("Repository: %-30s | URL: %-20s\n", r.Name, url)
	}
}
