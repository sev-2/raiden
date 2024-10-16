package db

import (
	"log"
	"os"
	"sort"
	"strings"

	"github.com/sev-2/raiden"
)

func getConfig() *raiden.Config {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return nil
	}

	configFilePath := strings.Join([]string{currentDir, "app.yaml"}, string(os.PathSeparator))

	config, err := raiden.LoadConfig(&configFilePath)
	if err != nil {
		log.Println(err)
		return nil
	}

	return config
}

func keyExist(maps map[string]string, s string) bool {
	for key := range maps {
		if key == s {
			return true
		}
	}
	return false
}

func reverseSortString(n []string) []string {
	sort.Slice(n, func(i, j int) bool {
		return n[i] < n[j]
	})

	return n
}
