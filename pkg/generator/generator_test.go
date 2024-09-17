package generator_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/stretchr/testify/assert"
)

func TestGenerate_ErrorWritingToFile(t *testing.T) {
	invalidPath := "/invalid_path/output.txt"

	tmpl := "{{ .Name }}"
	input := generator.GenerateInput{
		BindData:     struct{ Name string }{"John"},
		Template:     tmpl,
		TemplateName: "testTemplate",
		OutputPath:   invalidPath,
	}

	err := generator.Generate(input, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed create file")
}

func TestGenerate_ErrorParsingTemplate(t *testing.T) {
	input := generator.GenerateInput{
		BindData:     nil,
		Template:     "{{ .invalid}",
		TemplateName: "testTemplate",
		OutputPath:   "test_output.txt",
	}

	err := generator.Generate(input, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing")
}

func TestGenerate_ErrorExecutingTemplate(t *testing.T) {
	tmpl := "{{ .Name }}"
	input := generator.GenerateInput{
		BindData:     struct{}{},
		Template:     tmpl,
		TemplateName: "testTemplate",
		OutputPath:   "test_output.txt",
	}

	err := generator.Generate(input, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error execute template")
}

// Test for FileWriter Write function
func TestFileWriter_Write_ErrorCreatingFile(t *testing.T) {
	invalidPath := "/invalid_path/output.txt" // Simulating an invalid path

	fw := &generator.FileWriter{FilePath: invalidPath}
	_, err := fw.Write([]byte("test content"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed create file")
}

func TestFileWriter_Write_ErrorFormattingCode(t *testing.T) {
	validPath := "test_err_output_formatting.txt"

	fw := &generator.FileWriter{FilePath: validPath}

	input := generator.GenerateInput{
		BindData:     struct{ Name string }{"invalid code"},
		Template:     "{{ .Name }}",
		TemplateName: "testTemplate",
	}

	err := generator.Generate(input, fw)
	assert.Contains(t, err.Error(), "error format code")
}
