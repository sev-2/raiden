package raiden

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
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

	options := marshallComment(comment)
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

func marshallComment(comment string) ControllerOptions {
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
		case string(ControllerOptionRouteKey):
			if len(splitedComments) > 2 {
				options.Method = strings.ToUpper(splitedComments[1])
				options.Path = splitedComments[2]
			} else if len(splitedComments) == 2 {
				options.Path = splitedComments[1]
			}
		case string(ControllerOptionTypeKey):
			if len(splitedComments) >= 2 {
				options.Type = ControllerType(splitedComments[1])

				switch options.Type {
				case ControllerTypeFunction:
					options.Path = fasthttp.MethodPost
				}
			}
		}
	}

	return options
}

// @type http-handler
// @route GET /health
func HealthHandler(ctx *Context) Presenter {
	responseData := map[string]any{
		"message": "server up",
	}
	return ctx.SendData(responseData)
}

func ProxyHandler(
	url *url.URL,
	requestInterceptor func(req *http.Request),
	responseInterceptor func(r *http.Response) error,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = url.Scheme
			req.URL.Host = url.Host
			logger.Infof("Proxying to : %s %s\n", utils.GetColoredHttpMethod(req.Method), req.URL.String())

			// intercept request
			if requestInterceptor != nil {
				requestInterceptor(req)
			}
		}

		if responseInterceptor != nil {
			proxy.ModifyResponse = responseInterceptor
		}

		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
