package resource_test

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/resource"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestImport(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath:   "test_project",
		AllowedSchema: "public",
	}
	config := &raiden.Config{}
	resource.ImportLogger = logger.HcLog().Named("import")

	err := resource.Import(flags, config)
	assert.Error(t, err)
}

func TestUpdateLocalStateFromImport(t *testing.T) {
	localState := &state.LocalState{}
	stateChan := make(chan any)
	done := resource.UpdateLocalStateFromImport(localState, stateChan)

	close(stateChan)
	err := <-done
	assert.NoError(t, err)
}

func TestPrintImportReport(t *testing.T) {
	report := resource.ImportReport{
		Table:   1,
		Role:    2,
		Rpc:     3,
		Storage: 4,
	}

	resource.PrintImportReport(report, false)
	resource.PrintImportReport(report, true)

	report = resource.ImportReport{}
	resource.PrintImportReport(report, false)
	resource.PrintImportReport(report, true)
}

func TestImportDecorateFunc(t *testing.T) {
	if os.Getenv("TEST_RUN") == "1" {
		data := []objects.Role{
			{Name: "role1"},
			{Name: "role2"},
		}
		findFunc := func(item objects.Role, input generator.GenerateInput) bool {
			if i, ok := input.BindData.(generator.GenerateRoleData); ok {
				return i.Name == item.Name
			}
			return false
		}
		stateChan := make(chan any)
		defer close(stateChan)

		decorateFunc := resource.ImportDecorateFunc(data, findFunc, stateChan)
		input := generator.GenerateInput{
			BindData: generator.GenerateRoleData{Name: "role1"},
		}

		err := decorateFunc(input, nil)
		assert.NoError(t, err)

		select {
		case rs := <-stateChan:
			assert.NotNil(t, rs)
		default:
			t.Error("expected stateChan to have value")
		}
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestImportDecorateFunc")
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	err := cmd.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	err1 := cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err1)
}

func TestFindImportResource(t *testing.T) {
	data := []objects.Role{
		{Name: "role1"},
		{Name: "role2"},
	}
	findFunc := func(item objects.Role, input generator.GenerateInput) bool {
		if i, ok := input.BindData.(generator.GenerateRoleData); ok {
			return i.Name == item.Name
		}
		return false
	}

	input := generator.GenerateInput{
		BindData: generator.GenerateRoleData{Name: "role1"},
	}

	item, found := resource.FindImportResource(data, input, findFunc)
	assert.True(t, found)
	assert.Equal(t, "role1", item.Name)

	input = generator.GenerateInput{
		BindData: generator.GenerateRoleData{Name: "role3"},
	}

	item, found = resource.FindImportResource(data, input, findFunc)
	assert.False(t, found)
}
