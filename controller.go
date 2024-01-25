package raiden

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/valyala/fasthttp"
)

type (
	ControllerType      string
	ControllerOptionKey string
	ControllerOptions   struct {
		Type   ControllerType
		Method string
		Path   string
	}

	Controller struct {
		Options ControllerOptions
		Handler RouteHandlerFn
	}

	ControllerRegistry struct {
		ControllerComments map[string]string
		Controllers        []Controller
	}
)

const (
	ControllerTypeHttpHandler ControllerType = "http-handler"
	ControllerTypeRpc         ControllerType = "rpc"
	ControllerTypeFunction    ControllerType = "function"

	ControllerOptionTypeKey  ControllerOptionKey = "type"
	ControllerOptionRouteKey ControllerOptionKey = "route"
)

// ----- Controller Functionality -----
func NewControllerRegistry() *ControllerRegistry {
	registry := &ControllerRegistry{
		ControllerComments: map[string]string{},
	}
	registry.Register(HealthController)
	return registry
}

func (cr *ControllerRegistry) Register(routeHandlers ...RouteHandlerFn) {
	var controller []Controller
	for i := range routeHandlers {
		h := routeHandlers[i]
		if c := cr.getController(h); c.Handler != nil {
			controller = append(controller, cr.getController(h))
		}
	}
	cr.Controllers = append(cr.Controllers, controller...)
}

func (cr *ControllerRegistry) getController(h RouteHandlerFn) Controller {
	controller := Controller{
		Handler: nil,
	}

	funcPtr := reflect.ValueOf(h).Pointer()
	funcInfo := runtime.FuncForPC(funcPtr)

	funcNameArr := strings.Split(funcInfo.Name(), ".")
	funcName := funcNameArr[len(funcNameArr)-1]

	comment, ok := cr.ControllerComments[funcName]
	if !ok {
		file, _ := funcInfo.FileLine(funcPtr)
		mapCommentArr := GetHandlerComment(file)
		if len(mapCommentArr) == 0 {
			return controller
		}

		for _, fc := range mapCommentArr {
			cr.ControllerComments[fc["fn"]] = fc["comment"]
		}

		comment, ok = cr.ControllerComments[funcName]
		if !ok {
			return controller
		}
	}

	options := parseToOptions(comment)
	return Controller{
		Options: options,
		Handler: h,
	}
}

func GetHandlerComment(file string) []map[string]string {
	fSet := token.NewFileSet()
	node, err := parser.ParseFile(fSet, file, nil, parser.ParseComments)
	if err != nil {
		return nil
	}

	var commentsList []map[string]string
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			comments := extractComments(funcDecl.Doc)
			if len(comments) > 0 {
				commentsList = append(commentsList, map[string]string{
					"fn":      funcDecl.Name.Name,
					"comment": comments,
				})
			}
		}
	}
	return commentsList
}

func extractComments(doc *ast.CommentGroup) string {
	var comments []string
	if doc != nil {
		for _, comment := range doc.List {
			if strings.Contains(comment.Text, "@") {
				comments = append(comments, strings.TrimPrefix(comment.Text, "//"))
			}
		}
	}
	return strings.Join(comments, "\n")
}

func parseToOptions(comment string) ControllerOptions {
	trimmedComment := strings.ReplaceAll(comment, "\n", " ")

	var filteredComments []string
	for _, v := range strings.Split(trimmedComment, "@") {
		if len(v) == 0 {
			continue
		}

		filteredComments = append(filteredComments, v)
	}

	// create options
	options := ControllerOptions{}

	// loop for assign to options
	for _, fc := range filteredComments {
		splitedComments := strings.Split(fc, " ")
		switch splitedComments[0] {
		case string(ControllerOptionTypeKey):
			if len(splitedComments) >= 2 {
				options.Type = ControllerType(splitedComments[1])

				// setup method by type
				switch options.Type {
				case ControllerTypeFunction:
					options.Method = fasthttp.MethodPost

				}
			}
		case string(ControllerOptionRouteKey):
			if len(splitedComments) > 2 {
				options.Method = strings.ToUpper(splitedComments[1])
				options.Path = splitedComments[2]
			} else if len(splitedComments) == 2 {
				options.Path = splitedComments[1]
			}
		}
	}

	return options
}

// ----- Helper Functionality -----
func UnmarshalRequest[T any](ctx Context) (*T, error) {
	reqType := reflect.TypeOf((*T)(nil)).Elem()
	reqPtr := reflect.New(reqType).Interface()
	reqValue := reflect.ValueOf(reqPtr).Elem()

	if reqType.Kind() != reflect.Struct {
		return nil, errors.New("marshall request must passing struct type")
	}

	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)

		tagPath, tagQuery := field.Tag.Get("path"), field.Tag.Get("query")
		if field.Tag.Get("json") != "" {
			continue
		}

		var value string
		if tagPath != "" {
			tagValue := ctx.FastHttpRequestContext().UserValue(tagPath)
			if tagValueString, isString := tagValue.(string); isString {
				value = tagValueString
			}
		} else if tagQuery != "" {
			value = string(ctx.FastHttpRequestContext().Request.URI().QueryArgs().Peek(tagQuery))
		} else {
			continue
		}

		if err := setPayloadValue(reqValue.Field(i), value); err != nil {
			return nil, err
		}
	}

	if err := json.Unmarshal(ctx.FastHttpRequestContext().Request.Body(), reqPtr); err != nil {
		return nil, err
	}

	return reqPtr.(*T), nil
}

func UnmarshalRequestAndValidate[T any](ctx Context, requestValidators ...ValidatorRequest) (*T, error) {
	payload, err := UnmarshalRequest[T](ctx)
	if err != nil {
		return nil, err
	}

	if err := Validate(payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func setPayloadValue(fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("%s : must be integer value", fieldValue.Type().Name())
		}
		fieldValue.SetInt(int64(intValue))
	default:
		return fmt.Errorf("%s : unsupported field type %s", fieldValue.Type().Name(), fieldValue.Kind())
	}

	return nil
}

// ----- Handlers -----

// @type http-handler
// @route GET /health
func HealthController(ctx Context) Presenter {
	responseData := map[string]any{
		"message": "server up",
	}
	return ctx.SendJson(responseData)
}

func ProxyHandler(
	targetURL *url.URL,
	requestInterceptor func(req *fasthttp.Request),
	responseInterceptor func(resp *fasthttp.Response) error,
) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		req, res := &ctx.Request, &ctx.Response

		req.SetRequestURI(targetURL.String())
		req.URI().SetScheme(targetURL.Scheme)
		req.URI().SetHost(targetURL.Host)

		logger.Infof("Proxying to : %s %s\n", req.Header.Method(), req.URI().FullURI())

		if requestInterceptor != nil {
			requestInterceptor(req)
		}

		if err := fasthttp.Do(req, res); err != nil {
			logger.Error(err)
			return
		}

		if responseInterceptor != nil {
			if err := responseInterceptor(res); err != nil {
				logger.Error(err)
				return
			}
		}
	}
}
