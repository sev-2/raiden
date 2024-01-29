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
	// controller contract and capabilities
	// executed order
	// Before -> Handler -> After
	Controller interface {
		After(ctx Context) error
		Before(ctx Context) error
		Handler(ctx Context) Presenter
	}
	ControllerBase struct{}
)

// `Before` function executed before `Handler“ function executed
// overwrite function in embedded struct for custom logic
// if error is not null, process will be intercepted
// and return response error message to client with json format
func (*ControllerBase) Before(ctx Context) error {
	return nil
}

// `Handler` function is main logic of controller
// Overwrite function in embedded struct for custom logic
func (*ControllerBase) Handler(ctx Context) Presenter {
	return ctx.SendJsonErrorWithCode(fasthttp.StatusNotImplemented, errors.New("handler not implemented"))
}

// `After“ function executed before `Handler“ function executed
// Overwrite function in embedded struct for custom logic
// if error is not null, process will be continue processed to client,
// error message only print to standard output
func (*ControllerBase) After(ctx Context) error {
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
	if err := json.Unmarshal(ctx.Request.Body(), payloadPtr); err != nil {
		return err
	}

	if err := Validate(payloadPtr); err != nil {
		logger.Debugf("validate error : %v", err)
		return err
	}

	filedValue := controllerValue.FieldByName("Payload")
	filedValue.Set(reflect.ValueOf(payloadPtr))
	return nil
}

// set value to reflect value field
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

type HealthRequest struct{}
type HealthResponse struct {
	Message string `json:"message"`
}
type HealthController struct {
	ControllerBase
	Payload *HealthRequest
	Result  HealthResponse
}

func (c *HealthController) Handler(ctx Context) Presenter {
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
