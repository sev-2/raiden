package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

var GeneratorLogger hclog.Logger = logger.HcLog().Named("generator")

// ----- Define type, variable and constant -----
type GenerateInput struct {
	BindData     any
	Template     string
	TemplateName string
	OutputPath   string
	FuncMap      []template.FuncMap
}

type GenerateFn func(input GenerateInput, writer io.Writer) error

// ----- Generate functionality  -----
func DefaultWriter(filePath string) (*os.File, error) {
	file, err := utils.CreateFile(filePath, true)
	if err != nil {
		return nil, fmt.Errorf("failed create file %s : %v", filePath, err)
	}

	return file, nil
}

type FileWriter struct {
	FilePath string
	file     *os.File
}

func (fw *FileWriter) Write(p []byte) (int, error) {
	if fw.file == nil {
		file, err := utils.CreateFile(fw.FilePath, true)
		if err != nil {
			return 0, fmt.Errorf("failed create file %s : %v", fw.FilePath, err)
		}
		fw.file = file
		defer fw.Close()
	}

	formattedCode, err := format.Source(p)
	if err != nil {
		return 0, fmt.Errorf("error format code : %v", err)
	}

	return fw.file.Write(formattedCode)
}

// Close closes the underlying file
func (fw *FileWriter) Close() error {
	if fw.file != nil {
		return fw.file.Close()
	}
	return nil
}

func Generate(input GenerateInput, writer io.Writer) error {
	// set default writer
	if writer == nil {
		file, err := DefaultWriter(input.OutputPath)
		if err != nil {
			return err
		}
		writer = file
	}

	tmplInstance := template.New(input.TemplateName)
	for _, tm := range input.FuncMap {
		tmplInstance.Funcs(tm)
	}

	tmpl, err := tmplInstance.Parse(input.Template)
	if err != nil {
		return fmt.Errorf("error parsing : %v", err)
	}

	var renderedCode bytes.Buffer
	err = tmpl.Execute(&renderedCode, input.BindData)
	if err != nil {
		return fmt.Errorf("error execute template : %v", err)
	}

	_, err = writer.Write(renderedCode.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func CreateInternalFolder(basePath string) (err error) {
	internalFolderPath := filepath.Join(basePath, "internal")
	GeneratorLogger.Trace("create internal folder if not exist", "path", internalFolderPath)
	if exist := utils.IsFolderExists(internalFolderPath); !exist {
		if err := utils.CreateFolder(internalFolderPath); err != nil {
			return err
		}
	}
	return nil
}

func GenerateArrayDeclaration(value reflect.Value, withoutQuote bool) string {
	var arrayValues []string
	for i := 0; i < value.Len(); i++ {
		if withoutQuote {
			arrayValues = append(arrayValues, fmt.Sprintf("%s", value.Index(i).Interface()))
		} else {
			arrayValues = append(arrayValues, fmt.Sprintf("%q", value.Index(i).Interface()))
		}
	}
	return "[]string{" + strings.Join(arrayValues, ", ") + "}"
}

func getStructByBaseName(filePath string, baseStructName string) (r []string, err error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return r, err
	}

	// Traverse the AST to find the struct with the Http attribute
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if len(st.Fields.List) == 0 {
				continue
			}

			for _, f := range st.Fields.List {
				if se, isSe := f.Type.(*ast.SelectorExpr); isSe && se.Sel.Name == baseStructName {
					r = append(r, typeSpec.Name.Name)
					continue
				}
			}

		}
	}

	return
}
