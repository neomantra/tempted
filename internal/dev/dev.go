package dev

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var debugSet = os.Getenv("TEMPTED_DEBUG")

// dev
func Debug(msg string) {
	if debugSet != "" {
		f, err := tea.LogToFile("tempted.log", "")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		log.Printf("%q", msg)
		defer f.Close()
	}
}
