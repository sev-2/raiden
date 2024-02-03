package raiden

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/valyala/fasthttp"
)

// ----- Define type, variable and constant -----
type (
	RpcSecurityType string
	RpcParam        struct {
		Name    string
		Type    string
		Default *string
		Value   any
	}

	Rpc struct {
		Model         any
		ModelKey      string
		Schema        string
		Name          string
		Params        []RpcParam
		Definition    string
		Security      RpcSecurityType
		ReturnType    string
		Hash          string
		CompleteQuery string
		IsDeprecated  bool
	}

	RpcRegistry struct {
		BaseUrl  string
		BasePath string
		Locker   sync.RWMutex
		Data     map[string]*Rpc
	}

	RpcCompareResult struct {
		IsDifferent   bool
		DefinitionRpc *Rpc
		ActualRpc     *Rpc
	}

	RpcMigrateStatus struct {
		Error error
		Name  string
	}

	RpcCompareResults []RpcCompareResult
)

var (
	rpcRegistryInstance = RpcRegistry{
		BasePath: "rest/v1/rpc",
		Data:     make(map[string]*Rpc),
		Locker:   sync.RWMutex{},
	}

	defaultParamPrefix = "in"
)

const (
	SecurityTypeDefiner RpcSecurityType = "DEFINER"
	SecurityTypeInvoker RpcSecurityType = "INVOKER"
	RpcTemplate                         = `CREATE OR REPLACE FUNCTION :function_name(:params) RETURNS :return_type LANGUAGE plpgsql :security AS $function$ :definition $function$`
)

// ----- Rpc Functionality  -----
func NewRpc(name string) *Rpc {
	return &Rpc{
		Name: name,
	}
}

func (rpc *Rpc) SetName(name string) *Rpc {
	rpc.Name = name
	return rpc
}

func (rpc *Rpc) SetModel(model any, key string) *Rpc {
	rpc.Model = model
	rpc.ModelKey = key

	modelType := reflect.TypeOf(rpc.Model)
	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}

	// get metadata
	if metadataField, found := modelType.FieldByName("Metadata"); found {
		if schemaTag := metadataField.Tag.Get("schema"); schemaTag != "" {
			rpc.Schema = schemaTag
		}
	}

	if rpc.Schema == "" {
		rpc.Schema = "public"
	}

	return rpc
}

func (rpc *Rpc) SetParam(name string, paramType string, defaultValue *string) *Rpc {
	rpc.Params = append(rpc.Params, RpcParam{
		Name:    name,
		Type:    paramType,
		Default: defaultValue,
	})

	return rpc
}

func (rpc *Rpc) SetParamStruct(input any) *Rpc {
	inputType := reflect.TypeOf(input)
	inputValue := reflect.ValueOf(input)
	if inputType.Kind() == reflect.Pointer {
		inputType = inputType.Elem()
		inputValue = inputValue.Elem()
	}

	for i := 0; i < inputType.NumField(); i++ {
		field, fieldValue := inputType.Field(i), inputValue.Field(i)
		param := RpcParam{
			Name:  strings.ToLower(utils.ToSnakeCase(field.Name)),
			Value: fieldValue.Interface(),
		}

		rpcParamType, defaultValue, suffixType := field.Type.String(), "", ""

		if strings.Contains(rpcParamType, "*") {
			defaultValue = "null"
			rpcParamType = strings.TrimLeft(rpcParamType, "*")
		}

		if strings.Contains(rpcParamType, "[]") {
			suffixType = "[]"
			rpcParamType = strings.Trim(rpcParamType, "[]")
		}

		if defaultValue != "" {
			param.Default = &defaultValue
		}

		// set param type
		pgType := postgres.ToPostgresType(rpcParamType)
		param.Type = pgType + suffixType

		defaultTag := field.Tag.Get("default")
		if defaultTag != "" {
			switch pgType {
			case "text", "varchar":
				defaultTag = fmt.Sprintf("'%s'", defaultTag)
				param.Default = &defaultTag
			default:
				param.Default = &defaultTag
			}
		}

		rpc.Params = append(rpc.Params, param)
	}
	return rpc
}

func (rpc *Rpc) SetParams(params []RpcParam) *Rpc {
	rpc.Params = append(rpc.Params, params...)
	return rpc
}

func (rpc *Rpc) SetQuery(query string) *Rpc {
	rpc.Definition = query
	return rpc
}

func (rpc *Rpc) SetSecurity(security RpcSecurityType) *Rpc {
	rpc.Security = security
	return rpc
}

func (rpc *Rpc) buildParam() string {
	var params []string
	for _, p := range rpc.Params {
		param := fmt.Sprintf("%s_%s %s", defaultParamPrefix, p.Name, p.Type)
		if p.Default != nil {
			param += " default " + *p.Default
		}
		params = append(params, param)
	}
	return strings.Join(params, ",")
}

func (rpc *Rpc) paramToMap() map[string]any {
	mapParam := make(map[string]any)
	for _, p := range rpc.Params {
		mapParam[p.Name] = p.Value
	}

	return mapParam
}

func (rpc *Rpc) paramInputToMap() map[string]any {
	mapParam := make(map[string]any)
	for _, p := range rpc.Params {
		key := fmt.Sprintf("%s_%s", defaultParamPrefix, p.Name)
		mapParam[key] = p.Value
	}

	return mapParam
}

func (rpc *Rpc) extractNamedKey() map[string]bool {
	pattern := `:([^\s]+)`

	re := regexp.MustCompile(pattern)

	matches := re.FindAllStringSubmatch(rpc.Definition, -1)

	keys := make(map[string]bool)
	for _, match := range matches {
		if len(match) == 2 {
			keys[match[1]] = true
		}
	}

	return keys
}

func (rpc *Rpc) buildDefinition(modelType reflect.Type) (definition string, err error) {
	mapParam, keys, definition := rpc.paramToMap(), rpc.extractNamedKey(), rpc.Definition
	tableName := strings.ToLower(utils.ToSnakeCase(modelType.Name()))

	for k := range keys {
		if strings.Contains(k, ";") {
			k = strings.ReplaceAll(k, ";", "")
		}

		// check if exist if param
		if _, exist := mapParam[k]; exist {
			nKey := fmt.Sprintf("%s_%s", defaultParamPrefix, k)
			definition = strings.ReplaceAll(definition, ":"+k, nKey)
			continue
		}

		// check if model aliases
		if k == rpc.ModelKey {
			definition = utils.MatchReplacer(definition, ":"+k, tableName)
			continue
		}

		// check is relation
		splitKeys := strings.Split(k, ".")
		if len(splitKeys) > 1 && splitKeys[0] == rpc.ModelKey {
			nestedTableName, err := rpc.findTableName(modelType, splitKeys[1:])
			if err != nil {
				return "", err
			}
			definition = utils.MatchReplacer(definition, ":"+k, utils.ToSnakeCase(nestedTableName))
		}

	}

	return
}

func (rpc Rpc) findTableName(modelType reflect.Type, findKeys []string) (tableName string, err error) {
	// stop condition
	if len(findKeys) == 0 && modelType.Kind() == reflect.Struct {
		tableName = modelType.Name()
		return
	}

	head, tail := findKeys[:1], findKeys[1:]
	field, found := modelType.FieldByName(head[0])
	if !found {
		err = fmt.Errorf("field %s in model %s is not defined", head[0], modelType.Name())
		return
	}

	if field.Type.Kind() == reflect.Struct && len(tail) > 0 {
		return field.Type.Name(), nil
	}

	switch field.Type.Kind() {
	case reflect.Ptr:
		return rpc.findTableName(field.Type.Elem(), tail)
	case reflect.Slice, reflect.Array:
		findType := field.Type.Elem()
		if findType.Kind() == reflect.Ptr {
			findType = findType.Elem()
		}

		return rpc.findTableName(findType, tail)
	case reflect.Struct:
		return rpc.findTableName(field.Type, tail)
	}

	return
}

func (rpc *Rpc) GetDefinition() (definition string, err error) {
	modelType := reflect.TypeOf(rpc.Model)
	return rpc.buildDefinition(modelType)
}

func (rpc *Rpc) BuildQuery() (err error) {
	// check if model is set
	if rpc.Model == nil {
		err = fmt.Errorf("rpc : %s - model can`t be null, set model using BindModel(alias string, model any", rpc.Name)
		return
	}

	// define model type
	modelType := reflect.TypeOf(rpc.Model)
	// check if model type is pointer and rewrite with reflect type from element
	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}

	// start generate query
	var q string

	// define new query
	q = strings.ReplaceAll(RpcTemplate, ":function_name", fmt.Sprintf("%s.%s", rpc.Schema, rpc.Name))

	// set param
	q = strings.ReplaceAll(q, ":params", rpc.buildParam())

	// set return type
	if rpc.ReturnType == "" {
		tableName := strings.ToLower(utils.ToSnakeCase(modelType.Name()))
		rpc.ReturnType = fmt.Sprintf("setof %s", tableName)
	}
	q = strings.ReplaceAll(q, ":return_type", rpc.ReturnType)

	// set security
	if rpc.Security == "" {
		rpc.Security = SecurityTypeDefiner
	}

	if rpc.Security == SecurityTypeDefiner {
		q = strings.ReplaceAll(q, ":security", "SECURITY DEFINER")
	} else {
		q = strings.ReplaceAll(q, ":security", "")
	}

	// set definitions
	definition, err := rpc.buildDefinition(modelType)
	if err != nil {
		return err
	}
	q = strings.ReplaceAll(q, ":definition", definition)

	// clean up space
	q = strings.ToLower(strings.ReplaceAll(q, "\n", ""))
	q = strings.ReplaceAll(q, "  ", " ")

	// hash query
	rpc.Hash = utils.HashString(q)

	// set generated statement
	rpc.CompleteQuery = q

	return
}

func (rpc Rpc) attachAuthHeader(inReq *fasthttp.Request) func(*fasthttp.Request) {
	return func(outReq *fasthttp.Request) {
		if authHeader := inReq.Header.Peek("Authorization"); len(authHeader) > 0 {
			outReq.Header.AddBytesV("Authorization", authHeader)
		}

		if apiKey := inReq.Header.Peek("apiKey"); len(apiKey) > 0 {
			outReq.Header.AddBytesV("apiKey", apiKey)
		}
	}
}

func (rpc *Rpc) sendRequest(body []byte, reqInterceptor func(req *fasthttp.Request)) ([]byte, error) {
	// check is rpc exist in registry
	if !rpcRegistryInstance.IsExist(rpc.Name) {
		return nil, &ErrorResponse{
			StatusCode: fasthttp.StatusNotFound,
			Details:    fmt.Sprintf("rpc %s is not registered", rpc.Name),
			Message:    "rpc not found",
		}
	}

	// set url
	apiUrl := fmt.Sprintf("%s/%s/%s", rpcRegistryInstance.BaseUrl, rpcRegistryInstance.BasePath, rpc.Name)
	resData, err := utils.SendRequest(fasthttp.MethodPost, apiUrl, body, reqInterceptor)

	if err != nil {
		sendErr, isHaveData := err.(utils.SendRequestError)
		if isHaveData {
			var errResponse ErrorResponse
			if err := json.Unmarshal(sendErr.Body, &errResponse); err == nil {
				return nil, &errResponse
			}
		}

		return nil, &ErrorResponse{
			StatusCode: fasthttp.StatusInternalServerError,
			Details:    err.Error(),
			Message:    "fail request to upstream",
		}
	}

	return resData, nil
}

func (rpc *Rpc) Execute(incomingReq *fasthttp.Request, dest any) error {
	body, err := json.Marshal(rpc.paramInputToMap())
	if err != nil {
		return &ErrorResponse{
			StatusCode: fasthttp.StatusBadRequest,
			Details:    err,
			Message:    "payload is not valid",
		}
	}

	resData, err := rpc.sendRequest(body, rpc.attachAuthHeader(incomingReq))
	if err != nil {
		return err
	}

	// sample data
	if err := json.Unmarshal(resData, dest); err != nil {
		return &ErrorResponse{
			StatusCode: fasthttp.StatusInternalServerError,
			Details:    err,
			Message:    "invalid marshall response data",
		}
	}

	return nil
}

func (rpc *Rpc) ExecuteWithParam(incomingReq *fasthttp.Request, params map[string]any, dest any) error {
	mapInputs := rpc.paramInputToMap()
	for k := range mapInputs {
		filterKey := fmt.Sprintf("%s_", defaultParamPrefix)
		kCheck := strings.ReplaceAll(k, filterKey, "")

		if pv, isExist := params[kCheck]; isExist {
			mapInputs[k] = pv
		}
	}

	body, err := json.Marshal(mapInputs)
	if err != nil {
		return &ErrorResponse{
			StatusCode: fasthttp.StatusBadRequest,
			Details:    err,
			Message:    "payload is not valid",
		}
	}

	resData, err := rpc.sendRequest(body, rpc.attachAuthHeader(incomingReq))
	if err != nil {
		return err
	}

	// sample data
	if err := json.Unmarshal(resData, dest); err != nil {
		return &ErrorResponse{
			StatusCode: fasthttp.StatusInternalServerError,
			Details:    err,
			Message:    "invalid marshall response data",
		}
	}

	return nil
}

// ----- Rpc Registry Struct Functionality -----
func (rr *RpcRegistry) Register(rpc ...*Rpc) {
	rr.Locker.Lock()
	defer rr.Locker.Unlock()
	for _, r := range rpc {
		rr.Data[r.Name] = r
	}
}

func (rr *RpcRegistry) IsExist(name string) bool {
	rr.Locker.RLock()
	defer rr.Locker.RUnlock()

	_, isExist := rr.Data[name]
	return isExist
}

func (rr *RpcRegistry) Delete(name string) {
	rr.Locker.RLock()
	defer rr.Locker.RUnlock()

	delete(rr.Data, name)
}

func (rr *RpcRegistry) Compare(actualRpcMap map[string]*Rpc) (result RpcCompareResults, err error) {
	rr.Locker.RLock()
	defer rr.Locker.RUnlock()

	for k, definitionRpc := range rr.Data {
		// build definition query and hash
		err = definitionRpc.BuildQuery()
		if err != nil {
			return
		}

		// load actual rpc base on defined rpc
		actualRpc, isExist := actualRpcMap[k]
		if !isExist {
			// not exist in database
			result = append(result, RpcCompareResult{
				IsDifferent:   true,
				ActualRpc:     nil,
				DefinitionRpc: definitionRpc,
			})
			continue
		}

		// handle if exist in database
		result = append(result, RpcCompareResult{
			IsDifferent:   definitionRpc.Hash != actualRpc.Hash,
			ActualRpc:     actualRpc,
			DefinitionRpc: definitionRpc,
		})
	}
	return
}

// ----- RpcCompareResults -----
func (results RpcCompareResults) PrintResult() {
	for i := range results {
		rs := results[i]
		if rs.IsDifferent {
			print := color.New(color.FgHiBlack).PrintfFunc()
			print("*** Found different rpc : %s", rs.DefinitionRpc.Name)
			print = color.New(color.FgGreen).PrintfFunc()
			print("\n// Defined rpc :  \n%s\n", rs.DefinitionRpc.CompleteQuery)
			print = color.New(color.FgRed).PrintfFunc()
			if rs.ActualRpc != nil {
				print("\n// Actual rpc : \n%s \n", rs.ActualRpc.CompleteQuery)
			} else {
				print("\n// Actual rpc : \nnot exist in database \n")
			}
			print = color.New(color.FgHiBlack).PrintfFunc()
			print("\n*** End Found different rpc \n")
		}
	}
}

func (results RpcCompareResults) ExtractResult() (newRpc []*Rpc, conflictRpc RpcCompareResults) {
	newRpc = make([]*Rpc, 0)
	for i := range results {
		rs := results[i]
		if rs.ActualRpc == nil {
			newRpc = append(newRpc, rs.DefinitionRpc)
		}

		if rs.ActualRpc != nil && rs.DefinitionRpc != nil && rs.IsDifferent {
			conflictRpc = append(conflictRpc, rs)
		}
	}
	return
}

// ----- Rpc Registry Global Function -----
func SyncRpc(config *Config, rpc ...*Rpc) error {
	// setup client
	switch config.DeploymentTarget {
	case DeploymentTargetCloud:
		if config.ProjectId == "" {
			return errors.New("PROJECT_ID need to be set in environment variable")
		}
		supabase.ConfigureManagementApi(config.SupabaseApiUrl, config.AccessToken)
	case DeploymentTargetSelfHosted:
		supabase.ConfigurationMetaApi(config.SupabaseApiUrl, config.SupabaseApiBaseUrl)
	}

	// register rpc base url
	rpcRegistryInstance.BaseUrl = config.SupabasePublicUrl

	// register defined rpc
	RegisterRpc(rpc...)

	// get function from supabase
	functions, err := supabase.GetFunctions(context.Background(), &config.ProjectId)
	if err != nil {
		Panic(err)
	}

	// convert function to map
	mapFunctions := mapFunctionsToRpc(functions)

	// compare rpc
	result, err := rpcRegistryInstance.Compare(mapFunctions)
	if err != nil {
		return err
	}

	// migrate new rpc to supabase
	newRpc, conflictRpc := result.ExtractResult()

	if len(conflictRpc) > 0 {
		conflictRpc.PrintResult()
		Panic("found conflict rpc please fix this first")
	}

	// migrate data
	if len(newRpc) > 0 {
		Info("start migrate new rpc data")
		rs := migrateNewRpc(&config.ProjectId, newRpc)
		errList, successList := make([]error, 0), make([]string, 0)

		// filter and split error and success result
		for _, r := range rs {
			if r.Error != nil {
				errList = append(errList, r.Error)
				rpcRegistryInstance.Delete(r.Name)
				continue
			}
			successList = append(successList, r.Name)
		}

		// print error result
		if len(errList) > 0 {
			var errArr []string
			for _, e := range errList {
				errArr = append(errArr, fmt.Sprintf("%s\n", e.Error()))
			}
			return errors.New(strings.Join(errArr, " "))
		}

		// print success result
		if len(successList) > 0 {
			Info("success migrate rpc :")
			for _, s := range successList {
				Infof("- %s", s)
			}
		}
	}

	return nil
}

func migrateNewRpc(projectId *string, rpcData []*Rpc) []*RpcMigrateStatus {
	wg, rsChan := sync.WaitGroup{}, make(chan RpcMigrateStatus)
	results := make([]*RpcMigrateStatus, 0)

	go func() {
		wg.Wait()
		close(rsChan)
	}()

	for _, rpc := range rpcData {
		if rpc == nil {
			continue
		}

		wg.Add(1)
		go func(r *Rpc, rChan chan RpcMigrateStatus) {
			defer wg.Done()

			rs := RpcMigrateStatus{Name: r.Name}
			defer func() {
				rChan <- rs
			}()

			_, err := supabase.ExecQuery(context.Background(), projectId, r.CompleteQuery)
			if err != nil {
				rs.Error = err
			}
		}(rpc, rsChan)
	}

	// waiting process
	for rs := range rsChan {
		results = append(results, &rs)
	}
	return results
}

func RegisterRpc(rpc ...*Rpc) {
	rpcRegistryInstance.Register(rpc...)
}

func IsRpcExist(name string) bool {
	return rpcRegistryInstance.IsExist(name)
}

func mapFunctionsToRpc(functions []supabase.Function) map[string]*Rpc {
	mapRpc := make(map[string]*Rpc)
	for i := range functions {
		f := functions[i]
		q := utils.CleanUpString(strings.ToLower(f.CompleteStatement))

		// add space for 'as'
		q = strings.ReplaceAll(q, "as $", " as $")

		mapRpc[f.Name] = &Rpc{
			Name:          f.Name,
			Schema:        f.Schema,
			CompleteQuery: q,
			ReturnType:    strings.ToLower(f.ReturnType),
			Hash:          utils.HashString(q),
		}
	}
	return mapRpc
}
