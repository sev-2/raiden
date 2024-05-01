package policies

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden/pkg/resource/migrator"
)

func PrintMigratesDiff(items []MigrateItem) {
	newData := []string{}
	deleteData := []string{}
	updateData := []string{}

	for i := range items {
		item := items[i]

		var name string
		if item.NewData.Name != "" {
			name = item.NewData.Name
		} else if item.OldData.Name != "" {
			name = item.OldData.Name
		}

		switch item.Type {
		case migrator.MigrateTypeCreate:
			newData = append(newData, fmt.Sprintf("- %s", name))
		case migrator.MigrateTypeUpdate:
			updateData = append(updateData, fmt.Sprintf("- %s", name))
		case migrator.MigrateTypeDelete:
			deleteData = append(deleteData, fmt.Sprintf("- %s", name))
		}
	}

	if len(newData) > 0 {
		Logger.Debug("List New Policy", "policy", fmt.Sprintf("\n %s", strings.Join(newData, "\n")))
	}

	if len(updateData) > 0 {
		Logger.Debug("List Updated Policy", "policy", fmt.Sprintf("\n%s", strings.Join(updateData, "\n")))
	}

	if len(deleteData) > 0 {
		Logger.Debug("List Delete Policy", "policy", fmt.Sprintf("\n %s", strings.Join(deleteData, "\n")))
	}
}
