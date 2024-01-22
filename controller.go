package raiden

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"reflect"
	"runtime"
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
		Handler RouteHandler
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

// controller registry
func NewControllerRegistry() *ControllerRegistry {
	return &ControllerRegistry{
		ControllerComments: map[string]string{},
	}
}

func (cr *ControllerRegistry) Register(routeHandlers ...RouteHandler) {
	var controller []Controller
	for i := range routeHandlers {
		h := routeHandlers[i]
		if c := cr.getController(h); c.Handler != nil {
			controller = append(controller, cr.getController(h))
		}
	}
	cr.Controllers = append(cr.Controllers, controller...)
}

func (cr *ControllerRegistry) getController(h RouteHandler) Controller {
	controller := Controller{
		Handler: nil,
	}

	funcPtr := reflect.ValueOf(h).Pointer()
	funcInfo := runtime.FuncForPC(funcPtr)

	funcNameArr := strings.Split(funcInfo.Name(), ".")
	funcName := funcNameArr[len(funcNameArr)-1]

	// find in registry
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

// Get Controller Options
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

// @type http-handler
// @route GET /health
func HealthHandler(ctx Context) Presenter {
	responseData := map[string]any{
		"message": "server up",
	}
	return ctx.SendData(responseData)
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
