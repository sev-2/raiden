package raiden

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"

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
