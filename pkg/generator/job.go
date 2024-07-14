package generator

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

var JobLogger hclog.Logger = logger.HcLog().Named("generator.Job")

// ----- Define type, variable and constant -----
type JobFieldAttribute struct {
	Field string
	Type  string
	Tag   string
}

type GenerateJobData struct {
	Imports           []string
	JobName           string
	SnakeCasedJobName string
	Package           string
}

const (
	JobTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}

type {{ .JobName }}Job struct {
	raiden.JobBase
}

func (j *{{ .JobName }}Job) Name() string {
	return "{{ .SnakeCasedJobName }}"
}

func (j *{{ .JobName }}Job) Before(ctx raiden.JobContext, jobID uuid.UUID, jobName string) {
	raiden.Info("before execute", "name", jobName)
}

func (j *{{ .JobName }}Job) After(ctx raiden.JobContext, jobID uuid.UUID, jobName string) {
	raiden.Info("after execute", "name", jobName)
}

func (j *{{ .JobName }}Job) AfterErr(ctx raiden.JobContext, jobID uuid.UUID, jobName string, err error) {
	raiden.Error("after execute with error", "message", err.Error())
}

func (j *{{ .JobName }}Job) Duration() gocron.JobDefinition {
	return gocron.DurationJob(5 * time.Minute)
}

func (j *{{ .JobName }}Job) Task(ctx raiden.JobContext) error {
	fmt.Printf("Hello at %s \n", time.Now().String())

	return nil
}
`
)

// ----- Generate Job -----
func GenerateJob(file string, data GenerateJobData, generateFn GenerateFn) error {
	input := GenerateInput{
		BindData:     data,
		Template:     JobTemplate,
		TemplateName: "jobName",
		OutputPath:   file,
	}
	JobLogger.Debug("generate job", "path", input.OutputPath)
	return generateFn(input, nil)
}

// ----- Generate hello world -----
func GenerateHelloWorldJob(basePath string, generateFn GenerateFn) (err error) {
	JobPath := filepath.Join(basePath, JobDir)
	JobLogger.Trace("create jobs folder if not exist", "path", JobPath)
	if exist := utils.IsFolderExists(JobPath); !exist {
		if err := utils.CreateFolder(JobPath); err != nil {
			return err
		}
	}
	return createHelloWorldJob(JobPath, generateFn)
}

func createHelloWorldJob(JobPath string, generateFn GenerateFn) error {
	// set file path
	filePath := filepath.Join(JobPath, fmt.Sprintf("%s.go", "hello"))

	// set imports path
	imports := []string{
		fmt.Sprintf("%q", "fmt"),
		fmt.Sprintf("%q", "time"),

		fmt.Sprintf("%q", "github.com/go-co-op/gocron/v2"),
		fmt.Sprintf("%q", "github.com/google/uuid"),

		fmt.Sprintf("%q", "github.com/sev-2/raiden"),
	}

	data := GenerateJobData{
		Package:           "jobs",
		JobName:           "HelloWorld",
		SnakeCasedJobName: "hello_world_job",
		Imports:           imports,
	}

	return GenerateJob(filePath, data, generateFn)
}
