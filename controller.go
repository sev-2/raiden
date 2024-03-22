package raiden

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/valyala/fasthttp"
)

type (
	// The `Controller` interface defines a set of methods that a controller in the Raiden framework
	// should implement. These methods correspond to different HTTP methods (GET, POST, PUT, PATCH,
	// DELETE, OPTIONS, HEAD) and are used to handle incoming requests and generate responses. Each method
	// has a "Before" and "After" counterpart, which can be used to perform pre-processing and
	// post-processing tasks respectively.
	Controller interface {
		BeforeAll(ctx Context) error
		AfterAll(ctx Context) error

		AfterGet(ctx Context) error
		BeforeGet(ctx Context) error
		Get(ctx Context) error

		AfterPost(ctx Context) error
		BeforePost(ctx Context) error
		Post(ctx Context) error

		AfterPut(ctx Context) error
		BeforePut(ctx Context) error
		Put(ctx Context) error

		AfterPatch(ctx Context) error
		BeforePatch(ctx Context) error
		Patch(ctx Context) error

		AfterDelete(ctx Context) error
		BeforeDelete(ctx Context) error
		Delete(ctx Context) error

		AfterOptions(ctx Context) error
		BeforeOptions(ctx Context) error
		Options(ctx Context) error

		AfterHead(ctx Context) error
		BeforeHead(ctx Context) error
		Head(ctx Context) error
	}

	// The `ControllerBase` struct is a base struct that implements the `Controller` interface. It
	// provides default implementations for all the methods defined in the interface. These default
	// implementations return a `NotImplemented` error, indicating that the corresponding handler method
	// is not implemented in the actual controller. The actual controller can embed the `ControllerBase`
	// struct and override the methods as needed.
	ControllerBase struct{}
)

func (*ControllerBase) BeforeAll(ctx Context) error {
	return nil
}

func (*ControllerBase) AfterAll(ctx Context) error {
	return nil
}

func (*ControllerBase) BeforeGet(ctx Context) error {
	return nil
}

func (*ControllerBase) Get(ctx Context) error {
	return ctx.SendErrorWithCode(fasthttp.StatusNotImplemented, errors.New("handler not implemented"))
}

func (*ControllerBase) AfterGet(ctx Context) error {
	return nil
}

func (*ControllerBase) BeforePost(ctx Context) error {
	return nil
}

func (*ControllerBase) Post(ctx Context) error {
	return ctx.SendErrorWithCode(fasthttp.StatusNotImplemented, errors.New("handler not implemented"))
}

func (*ControllerBase) AfterPost(ctx Context) error {
	return nil
}

func (*ControllerBase) BeforePut(ctx Context) error {
	return nil
}

func (*ControllerBase) Put(ctx Context) error {
	return ctx.SendErrorWithCode(fasthttp.StatusNotImplemented, errors.New("handler not implemented"))
}

func (*ControllerBase) AfterPut(ctx Context) error {
	return nil
}

func (*ControllerBase) BeforePatch(ctx Context) error {
	return nil
}

func (*ControllerBase) Patch(ctx Context) error {
	return ctx.SendErrorWithCode(fasthttp.StatusNotImplemented, errors.New("handler not implemented"))
}

func (*ControllerBase) AfterPatch(ctx Context) error {
	return nil
}

func (*ControllerBase) BeforeDelete(ctx Context) error {
	return nil
}

func (*ControllerBase) Delete(ctx Context) error {
	return ctx.SendErrorWithCode(fasthttp.StatusNotImplemented, errors.New("handler not implemented"))
}

func (*ControllerBase) AfterDelete(ctx Context) error {
	return nil
}

func (*ControllerBase) BeforeOptions(ctx Context) error {
	return nil
}

func (*ControllerBase) Options(ctx Context) error {
	return ctx.SendErrorWithCode(fasthttp.StatusNotImplemented, errors.New("handler not implemented"))
}

func (*ControllerBase) AfterOptions(ctx Context) error {
	return nil
}

func (*ControllerBase) BeforeHead(ctx Context) error {
	return nil
}

func (*ControllerBase) Head(ctx Context) error {
	return ctx.SendErrorWithCode(fasthttp.StatusNotImplemented, errors.New("handler not implemented"))
}

func (*ControllerBase) AfterHead(ctx Context) error {
	return nil
}

// ----- Rest Controller -----
type RestController struct {
	Controller
	ModelName string
}

// AfterAll implements Controller.
// Subtle: this method shadows the method (Controller).AfterAll of RestController.Controller.
func (rc RestController) AfterAll(ctx Context) error {
	return rc.Controller.AfterAll(ctx)
}

// AfterDelete implements Controller.
// Subtle: this method shadows the method (Controller).AfterDelete of RestController.Controller.
func (rc RestController) AfterDelete(ctx Context) error {
	return rc.Controller.AfterDelete(ctx)
}

// AfterGet implements Controller.
// Subtle: this method shadows the method (Controller).AfterGet of RestController.Controller.
func (rc RestController) AfterGet(ctx Context) error {
	return rc.Controller.AfterGet(ctx)
}

// AfterHead implements Controller.
// Subtle: this method shadows the method (Controller).AfterHead of RestController.Controller.
func (rc RestController) AfterHead(ctx Context) error {
	return rc.Controller.AfterHead(ctx)
}

// AfterOptions implements Controller.
// Subtle: this method shadows the method (Controller).AfterOptions of RestController.Controller.
func (rc RestController) AfterOptions(ctx Context) error {
	return rc.Controller.AfterOptions(ctx)
}

// AfterPatch implements Controller.
// Subtle: this method shadows the method (Controller).AfterPatch of RestController.Controller.
func (rc RestController) AfterPatch(ctx Context) error {
	return rc.Controller.AfterPatch(ctx)
}

// AfterPost implements Controller.
// Subtle: this method shadows the method (Controller).AfterPost of RestController.Controller.
func (rc RestController) AfterPost(ctx Context) error {
	return rc.Controller.AfterPost(ctx)
}

// AfterPut implements Controller.
// Subtle: this method shadows the method (Controller).AfterPut of RestController.Controller.
func (rc RestController) AfterPut(ctx Context) error {
	return rc.Controller.AfterPut(ctx)
}

// BeforeAll implements Controller.
func (rc RestController) BeforeAll(ctx Context) error {
	// Implement validation
	queryParam := ctx.RequestContext().QueryArgs().String()
	decodedStr, err := url.QueryUnescape(queryParam)
	if err != nil {
		return ctx.SendError(err.Error())
	}

	countAsterisk := countCharOccurrences(decodedStr, "*")
	if countAsterisk > 4 {
		return ctx.SendErrorWithCode(400, errors.New("asterisk usage exceeds the maximum limit, use a maximum of 4 asterisks for better performance, if you need a complex query use rpc"))
	}

	countBracket := countBracketPairs(decodedStr)
	if countBracket > 5 {
		return ctx.SendErrorWithCode(400, errors.New("table usage exceeds the maximum limit, use a maximum of 5 table for better performance, if you need a complex query use rpc"))
	}

	return rc.Controller.BeforeAll(ctx)
}

// BeforeDelete implements Controller.
// Subtle: this method shadows the method (Controller).BeforeDelete of RestController.Controller.
func (rc RestController) BeforeDelete(ctx Context) error {
	return rc.Controller.BeforeDelete(ctx)
}

// BeforeGet implements Controller.
// Subtle: this method shadows the method (Controller).BeforeGet of RestController.Controller.
func (rc RestController) BeforeGet(ctx Context) error {
	return rc.Controller.BeforeGet(ctx)
}

// BeforeHead implements Controller.
// Subtle: this method shadows the method (Controller).BeforeHead of RestController.Controller.
func (rc RestController) BeforeHead(ctx Context) error {
	return rc.Controller.BeforeHead(ctx)
}

// BeforeOptions implements Controller.
// Subtle: this method shadows the method (Controller).BeforeOptions of RestController.Controller.
func (rc RestController) BeforeOptions(ctx Context) error {
	return rc.Controller.BeforeOptions(ctx)
}

// BeforePatch implements Controller.
// Subtle: this method shadows the method (Controller).BeforePatch of RestController.Controller.
func (rc RestController) BeforePatch(ctx Context) error {
	return rc.Controller.BeforePatch(ctx)
}

// BeforePost implements Controller.
// Subtle: this method shadows the method (Controller).BeforePost of RestController.Controller.
func (rc RestController) BeforePost(ctx Context) error {
	return rc.Controller.BeforePost(ctx)
}

// BeforePut implements Controller.
// Subtle: this method shadows the method (Controller).BeforePut of RestController.Controller.
func (rc RestController) BeforePut(ctx Context) error {
	return rc.Controller.BeforePut(ctx)
}

// Delete implements Controller.
func (rc RestController) Delete(ctx Context) error {
	return RestProxy(ctx, rc.ModelName)
}

// Get implements Controller.
func (rc RestController) Get(ctx Context) error {
	return RestProxy(ctx, rc.ModelName)
}

// Head implements Controller.
// Subtle: this method shadows the method (Controller).Head of RestController.Controller.
func (rc RestController) Head(ctx Context) error {
	return rc.Controller.Head(ctx)
}

// Options implements Controller.
// Subtle: this method shadows the method (Controller).Options of RestController.Controller.
func (rc RestController) Options(ctx Context) error {
	return rc.Controller.Options(ctx)
}

// Patch implements Controller.
func (rc RestController) Patch(ctx Context) error {
	return RestProxy(ctx, rc.ModelName)
}

// Post implements Controller.
func (rc RestController) Post(ctx Context) error {
	return RestProxy(ctx, rc.ModelName)
}

// Put implements Controller.
func (rc RestController) Put(ctx Context) error {
	return RestProxy(ctx, rc.ModelName)
}

// ----- Helper Functionality -----

// Marshall request data (path param, query and body data) to Payload data in
// actual controller
//
// Example :
//
//	type Request {
//			Search 		string	`query:"q"`
//			Resource 	string	`path:"resource" validate:"required"`
//	}
//
//	Controller {
//			raiden.ControllerBase
//			Payload	*Request
//	}
//
// Example Request :
// GET /hello/{resource}?q="some-resource"
//
// base on example above this code will auto marshall data from fasthttp.Request to Request struct
// and validate all data is appropriate base on validate tag
func MarshallAndValidate(ctx *fasthttp.RequestCtx, controller any) error {
	controllerType := reflect.TypeOf(controller).Elem()
	controllerValue := reflect.ValueOf(controller).Elem()

	payloadField, isPayloadFound := controllerType.FieldByName("Payload")
	if !isPayloadFound {
		return fmt.Errorf("field Payload is not exist in %s", controllerType.Name())
	}

	payloadType := payloadField.Type.Elem()
	payloadPtr := reflect.New(payloadType).Interface()
	payloadValue := reflect.ValueOf(payloadPtr).Elem()

	for i := 0; i < payloadType.NumField(); i++ {
		field := payloadType.Field(i)

		tagPath, tagQuery := field.Tag.Get("path"), field.Tag.Get("query")

		// handle marshall json with json.unmarshal in next process
		if field.Tag.Get("json") != "" {
			continue
		}

		var value string
		if tagPath != "" {
			tagValue := ctx.UserValue(tagPath)
			if tagValueString, isString := tagValue.(string); isString {
				value = tagValueString
			}
		} else if tagQuery != "" {
			value = string(ctx.Request.URI().QueryArgs().Peek(tagQuery))
		} else {
			continue
		}

		// bind value to struct attribute
		if err := setPayloadValue(payloadValue.Field(i), value); err != nil {
			return err
		}
	}

	// unmarshal data from request body to payload
	// only marshall to field with tag json
	requestBody := ctx.Request.Body()
	if requestBody != nil && string(requestBody) != "" {
		if err := json.Unmarshal(requestBody, payloadPtr); err != nil {
			return err
		}
	}

	// validate marshalled payload
	if err := Validate(payloadPtr); err != nil {
		return err
	}

	// set value to controller payload
	filedValue := controllerValue.FieldByName("Payload")
	filedValue.Set(reflect.ValueOf(payloadPtr))
	return nil
}

// The function `setPayloadValue` sets the value of a field in a struct based on its type.
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

// ----- Native Handlers -----

type HealthRequest struct{}
type HealthResponse struct {
	Message string `json:"message"`
}
type HealthController struct {
	ControllerBase
	Payload *HealthRequest
	Result  HealthResponse
}

func (c *HealthController) Get(ctx Context) error {
	responseData := map[string]any{
		"message": "server up",
	}
	return ctx.SendJson(responseData)
}

// RestHandler
func RestProxy(appCtx Context, modelName string) error {
	// Create a new request object
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Copy the original request to the new request object
	appCtx.RequestContext().Request.CopyTo(req)

	proxyUrl := fmt.Sprintf("%s/rest/v1/%s", appCtx.Config().SupabasePublicUrl, strings.ToLower(modelName))
	queryParam := appCtx.RequestContext().Request.URI().QueryString()
	if len(queryParam) > 0 {
		proxyUrl = fmt.Sprintf("%s?%s", proxyUrl, queryParam)
	}

	req.SetRequestURI(proxyUrl)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	Infof("Proxying to : %s %s", req.Header.Method(), req.URI().FullURI())
	if err := fasthttp.Do(req, resp); err != nil {
		return err
	}

	resp.Header.VisitAll(func(k, v []byte) {
		appCtx.RequestContext().Response.Header.SetBytesKV(k, v)
	})

	appCtx.RequestContext().Response.SetStatusCode(resp.StatusCode())
	appCtx.RequestContext().Response.SetBody(resp.Body())

	return nil
}

// Default Proxy Handler
func ProxyHandler(
	targetURL *url.URL,
	basePath string,
	requestInterceptor func(req *fasthttp.Request),
	responseInterceptor func(resp *fasthttp.Response) error,
) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Create a new request object
		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)

		// Copy the original request to the new request object
		ctx.Request.CopyTo(req)

		paths := strings.Split(req.URI().String(), basePath)
		if len(paths) < 2 {
			ctx.Request.Header.SetContentType("application/json")
			ctx.SetBodyString("{ \"message\" : \"invalid path\"}")
			return
		}

		proxyUrl := fmt.Sprintf("%s%s%s", targetURL.String(), basePath, paths[1])
		req.SetRequestURI(proxyUrl)

		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		Infof("Proxying to : %s %s", req.Header.Method(), req.URI().FullURI())

		if requestInterceptor != nil {
			requestInterceptor(req)
		}

		if err := fasthttp.Do(req, resp); err != nil {
			logger.Error(err)
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
			errResponse := fmt.Sprintf("{ \"messages\": %q}", err)
			ctx.Response.SetBodyString(errResponse)
			return
		}

		if responseInterceptor != nil {
			if err := responseInterceptor(resp); err != nil {
				logger.Error(err)
				ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
				errResponse := fmt.Sprintf("{ \"messages\": %q}", err)
				ctx.Response.SetBodyString(errResponse)
				return
			}
		}
		// Copy the response headers and body back to the original request context
		resp.Header.VisitAll(func(k, v []byte) {
			ctx.Response.Header.SetBytesKV(k, v)
		})

		ctx.Response.SetStatusCode(resp.StatusCode())
		ctx.Response.SetBody(resp.Body())
	}
}

func countCharOccurrences(str string, char string) int {
	re := regexp.MustCompile(regexp.QuoteMeta(char))
	return len(re.FindAllStringIndex(str, -1))
}

func countBracketPairs(str string) int {
	stack := 0
	count := 0
	for _, char := range str {
		if char == '(' {
			stack++
			if stack == 2 {
				count++
			}
		} else if char == ')' {
			stack--
			if stack == 0 {
				count++
			}
		}
	}
	return count
}
