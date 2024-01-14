package utils

import (
	"net/http"

	"github.com/fatih/color"
)

func GetColoredHttpMethod(httpMethod string) string {
	printFunc := color.New(color.FgBlack, color.BgHiWhite).SprintfFunc()
	switch httpMethod {
	case http.MethodGet:
		printFunc = color.New(color.FgWhite, color.BgGreen).SprintfFunc()
	case http.MethodPost:
		printFunc = color.New(color.FgWhite, color.BgHiYellow).SprintfFunc()
	case http.MethodPatch:
		printFunc = color.New(color.FgWhite, color.BgHiBlue).SprintfFunc()
	case http.MethodPut:
		printFunc = color.New(color.FgWhite, color.BgBlue).SprintfFunc()
	case http.MethodDelete:
		printFunc = color.New(color.FgWhite, color.BgHiRed).SprintfFunc()
	}

	return printFunc(" %s ", httpMethod)
}
