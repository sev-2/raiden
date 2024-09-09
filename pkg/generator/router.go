package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/valyala/fasthttp"
)

var RouterLogger hclog.Logger = logger.HcLog().Named("generator.router")

// ----- Define type, variable and constant -----
type (
	GenerateRouteItem struct {
		Type       string
		Path       string
		Methods    string
		Controller string
		Model      string
		Storage    string
	}

	GenerateRouterData struct {
		Imports []string
		Package string
		Routes  []GenerateRouteItem
	}

	FoundRoute struct {
		Package string
		Name    string
		Tag     string
		Methods []string
		Model   string
		Storage string
	}
)

const (
	RouterFilename = "route.go"
	RouterDir      = "internal/bootstrap"
	RouterTemplate = `// Code generated by raiden-cli; DO NOT EDIT.
package {{ .Package }}
{{if gt (len .Imports) 0 }}
import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{end }}
func RegisterRoute(server *raiden.Server) {
	server.RegisterRoute([]*raiden.Route{
		{{- range .Routes}}
		{
			Type:       {{ .Type }},
			Path:       {{ .Path }},
			{{- if ne .Methods ""}}
			Methods:    {{ .Methods }},
			{{- end}}
			Controller: &{{ .Controller }},
			{{- if ne .Model "" }}
			Model:      {{ .Model }},
			{{- end}}
			{{- if ne .Storage "" }}
			Storage:      &{{ .Storage }},
			{{- end}}
		},
		{{- end}}
	})
}
`
)

// Generate route configuration file
func GenerateRoute(basePath string, projectName string, generateFn GenerateFn) error {
	routePath := filepath.Join(basePath, RouterDir)
	RouterLogger.Trace("create bootstrap folder if not exist", routePath)
	if exist := utils.IsFolderExists(routePath); !exist {
		if err := utils.CreateFolder(routePath); err != nil {
			return err
		}
	}

	controllerPath := filepath.Join(basePath, ControllerDir)

	// scan all controller
	routes, err := WalkScanControllers(controllerPath)
	if err != nil {
		return err
	}

	input, err := createRouteInput(projectName, routePath, routes)
	if err != nil {
		return err
	}

	RouterLogger.Debug("generate route", "path", input.OutputPath)
	return generateFn(input, nil)
}

func WalkScanControllers(controllerPath string) ([]GenerateRouteItem, error) {
	RouterLogger.Trace("scan all controller", "path", controllerPath)

	routes := make([]GenerateRouteItem, 0)
	err := filepath.Walk(controllerPath, func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".go") {
			RouterLogger.Trace("collect routes", "path", path)
			rs, e := getRoutes(path)
			if e != nil {
				return e
			}

			for _, r := range rs {
				if r.Path != "" && len(r.Methods) > 0 {
					RouterLogger.Trace("found controller", "controller", r.Controller)
					routes = append(routes, r)
				}
			}

		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return routes, nil
}

func getRoutes(filePath string) (r []GenerateRouteItem, err error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return r, err
	}

	// Traverse the AST to find the struct with the Http attribute
	foundRouteMap := make(map[string]*FoundRoute)
	ast.Inspect(file, func(node ast.Node) bool {
		switch t := node.(type) {
		case *ast.TypeSpec:
			if t.Name != nil && t.Type != nil {
				// Check if it's a struct
				if st, ok := t.Type.(*ast.StructType); ok {
					foundRoute := &FoundRoute{}
					if fr, exist := foundRouteMap[t.Name.Name]; exist {
						foundRoute = fr
					}

					// Check if it has the Http / Model attribute
					foundField := false
					for _, field := range st.Fields.List {
						for _, fName := range field.Names {

							if fName != nil && fName.Name == "Http" && field.Tag != nil {
								tag := strings.Trim(field.Tag.Value, "`")
								foundRoute.Name = t.Name.Name
								foundRoute.Tag = tag

								foundRouteMap[t.Name.Name] = foundRoute
								foundField = true
								continue
							}

							if fName != nil && fName.Name == "Model" {
								switch fType := field.Type.(type) {
								case *ast.StarExpr:
									if se, ok := fType.X.(*ast.SelectorExpr); ok {
										foundRoute.Model = fmt.Sprintf("%s.%s{}", se.X, se.Sel.Name)
									}
								case *ast.SelectorExpr:
									foundRoute.Model = fmt.Sprintf("%s.%s{}", fType.X, fType.Sel.Name)
								case *ast.Ident:
									foundRoute.Model = fmt.Sprintf("%s{}", fType.Name)
								}

								foundRouteMap[t.Name.Name] = foundRoute
								foundField = true
								continue
							}

							if fName != nil && fName.Name == "Storage" {
								switch fType := field.Type.(type) {
								case *ast.StarExpr:
									if se, ok := fType.X.(*ast.SelectorExpr); ok {
										foundRoute.Storage = fmt.Sprintf("%s.%s{}", se.X, se.Sel.Name)
									}
								case *ast.SelectorExpr:
									foundRoute.Storage = fmt.Sprintf("%s.%s{}", fType.X, fType.Sel.Name)
								case *ast.Ident:
									foundRoute.Storage = fmt.Sprintf("%s{}", fType.Name)
								}

								foundRouteMap[t.Name.Name] = foundRoute
								foundField = true
								continue
							}
						}
					}

					// stop walk
					if foundField {
						return foundField
					}
				}
			}
		case *ast.FuncDecl:
			if t == nil || t.Recv == nil || t.Recv.List[0].Type == nil {
				return true
			}
			startExp, isStartExp := t.Recv.List[0].Type.(*ast.StarExpr)
			if isStartExp {
				structName := fmt.Sprintf("%s", startExp.X)
				route, isExist := foundRouteMap[structName]
				if isExist {
					switch strings.ToUpper(t.Name.Name) {
					case fasthttp.MethodGet:
						route.Methods = append(route.Methods, "fasthttp.MethodGet")
					case fasthttp.MethodPost:
						route.Methods = append(route.Methods, "fasthttp.MethodPost")
					case fasthttp.MethodPatch:
						route.Methods = append(route.Methods, "fasthttp.MethodPatch")
					case fasthttp.MethodPut:
						route.Methods = append(route.Methods, "fasthttp.MethodPut")
					case fasthttp.MethodDelete:
						route.Methods = append(route.Methods, "fasthttp.MethodDelete")
					case fasthttp.MethodOptions:
						route.Methods = append(route.Methods, "fasthttp.MethodOptions")
					case fasthttp.MethodHead:
						route.Methods = append(route.Methods, "fasthttp.MethodHead")
					}
				}
			}
		}
		return true
	})

	if len(foundRouteMap) == 0 {
		return r, nil
	}

	// bind package name
	fileDir := filepath.Dir(filePath)
	for _, m := range foundRouteMap {
		_, m.Package = filepath.Split(fileDir)
	}

	for _, m := range foundRouteMap {
		rNew, err := generateRoute(m)
		if err != nil {
			return r, err
		}
		r = append(r, rNew)
	}

	return
}

func generateRoute(foundRoute *FoundRoute) (GenerateRouteItem, error) {
	var r GenerateRouteItem

	r.Controller = fmt.Sprintf("%s.%s{}", foundRoute.Package, foundRoute.Name)
	r.Model = foundRoute.Model
	r.Storage = foundRoute.Storage

	tagItems := strings.Split(foundRoute.Tag, " ")
	r.Methods = GenerateArrayDeclaration(reflect.ValueOf(foundRoute.Methods), true)

	for _, tagItem := range tagItems {
		items := strings.Split(tagItem, ":")
		if len(items) != 2 {
			continue
		}

		trimmedItem := strings.Trim(items[1], "\"")
		switch items[0] {
		case "path":
			r.Path = fmt.Sprintf("%q", trimmedItem)
		case "type":

			// validate method
			// exclude for rest and storage controller
			// because automatically register by route
			if len(foundRoute.Methods) == 0 && trimmedItem != string(raiden.RouteTypeRest) && trimmedItem != string(raiden.RouteTypeStorage) {
				return r, fmt.Errorf("controller %s, required to set method handler. available method Get, Post, Put, Patch, Delete, and Option", foundRoute.Name)
			}

			if trimmedItem == string(raiden.RouteTypeRest) && r.Model == "" {
				return r, fmt.Errorf("controller %s, required to set model because have rest type", foundRoute.Name)
			}

			switch trimmedItem {
			case string(raiden.RouteTypeFunction):
				if len(foundRoute.Methods) > 1 {
					return r, fmt.Errorf("controller %s with type function,only allowed set 1 method and only allowed setup with post method", foundRoute.Name)
				}

				if len(foundRoute.Methods) == 1 && foundRoute.Methods[0] != "fasthttp.MethodPost" {
					return r, fmt.Errorf("controller %s with type function,only allowed setup with Post method", foundRoute.Name)
				}

				r.Type = "raiden.RouteTypeFunction"
			case string(raiden.RouteTypeCustom):
				r.Type = "raiden.RouteTypeCustom"
			case string(raiden.RouteTypeRpc):
				if len(foundRoute.Methods) > 1 {
					return r, fmt.Errorf("controller %s with type rpc,only allowed set 1 method and only allowed setup with post method", foundRoute.Name)
				}

				if len(foundRoute.Methods) == 1 && foundRoute.Methods[0] != "fasthttp.MethodPost" {
					return r, fmt.Errorf("controller %s with type rpc,only allowed setup with Post method", foundRoute.Name)
				}

				r.Type = "raiden.RouteTypeRpc"
			case string(raiden.RouteTypeRest):
				r.Type = "raiden.RouteTypeRest"
			case string(raiden.RouteTypeRealtime):
				r.Type = "raiden.RouteTypeRealtime"
			case string(raiden.RouteTypeStorage):
				r.Type = "raiden.RouteTypeStorage"
			default:
				return r, fmt.Errorf(
					"%s.%s : unsupported route type %s, available type are %s, %s, %s, %s, %s and %s ",
					foundRoute.Package, foundRoute.Name, items[1],
					raiden.RouteTypeFunction, raiden.RouteTypeCustom, raiden.RouteTypeRpc,
					raiden.RouteTypeRest, raiden.RouteTypeRealtime, raiden.RouteTypeStorage,
				)
			}
		}
	}
	return r, nil
}

func createRouteInput(projectName string, routePath string, routes []GenerateRouteItem) (input GenerateInput, err error) {
	// set file path
	filePath := filepath.Join(routePath, RouterFilename)

	// set imports path
	imports := []string{
		fmt.Sprintf("%q", "github.com/sev-2/raiden"),
	}

	if len(routes) > 0 {
		routeImportPath := fmt.Sprintf("%s/internal/controllers", utils.ToGoModuleName(projectName))
		imports = append(imports, fmt.Sprintf("%q", routeImportPath))
	}

	isHaveModel := false
	isHaveMethods := false
	isHaveStorage := false
	for i := range routes {
		r := routes[i]

		if r.Model != "" && !isHaveModel {
			isHaveModel = true
		}

		if r.Methods != "" && r.Methods != "[]string{}" && !isHaveMethods {
			isHaveMethods = true
		}

		if r.Storage != "" && !isHaveStorage {
			isHaveStorage = true
		}
	}

	if isHaveModel {
		modelImportPath := fmt.Sprintf("%s/internal/models", utils.ToGoModuleName(projectName))
		imports = append(imports, fmt.Sprintf("%q", modelImportPath))
	}

	if isHaveMethods {
		imports = append(imports, fmt.Sprintf("%q", "github.com/valyala/fasthttp"))
	}

	if isHaveStorage {
		storageImportPath := fmt.Sprintf("%s/internal/storages", utils.ToGoModuleName(projectName))
		imports = append(imports, fmt.Sprintf("%q", storageImportPath))
	}

	// set passed parameter
	sort.Strings(imports)
	sort.Slice(routes, func(i, j int) bool {
		iRunes := []rune(routes[i].Path)
		jRunes := []rune(routes[j].Path)

		max := len(iRunes)
		if max > len(jRunes) {
			max = len(jRunes)
		}

		for idx := 0; idx < max; idx++ {
			ir := iRunes[idx]
			jr := jRunes[idx]

			lir := unicode.ToLower(ir)
			ljr := unicode.ToLower(jr)

			if lir != ljr {
				return lir < ljr
			}

			if ir != jr {
				return ir < jr
			}
		}

		return len(iRunes) < len(jRunes)
	})

	data := GenerateRouterData{
		Package: "bootstrap",
		Imports: imports,
		Routes:  routes,
	}

	input = GenerateInput{
		BindData:     data,
		Template:     RouterTemplate,
		TemplateName: "routerTemplate",
		OutputPath:   filePath,
	}

	return
}
