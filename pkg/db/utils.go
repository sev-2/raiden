package db

import (
	"log"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource"
)

func findModel(targetName string) interface{} {
	for _, m := range resource.RegisteredModels {
		t := reflect.TypeOf(m)

		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if t.Name() == targetName {
			return m
		}
	}

	return nil
}

func findModelByTable(table string) string {
	for _, m := range resource.RegisteredModels {

		t := reflect.TypeOf(m)

		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Name == "Metadata" {
				tableName := field.Tag.Get("tableName")
				if tableName == table {
					parts := strings.Split(t.String(), ".")
					return parts[len(parts)-1]
				}
			}
		}
	}

	return ""
}

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
