//go:build !solution

package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

func MakeHandler(service interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handleRequest(w, req, service)
	})
}

func handleRequest(w http.ResponseWriter, req *http.Request, service interface{}) {
	input, err := decodeRequest(req)
	if err != nil {
		sendErrorResponse(w, "invalid request format", http.StatusBadRequest)
		return
	}

	targetMethod := reflect.ValueOf(service).MethodByName(input.Method)
	if !validateMethod(targetMethod) {
		sendErrorResponse(w, "method not found", http.StatusNotFound)
		return
	}

	argInstance, err := unmarshalParams(targetMethod, input.Params)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("invalid params: %v", err), http.StatusBadRequest)
		return
	}

	result, err := callMethod(req.Context(), targetMethod, argInstance)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccessResponse(w, result)
}

func decodeRequest(req *http.Request) (struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}, error) {
	var input struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	err := json.NewDecoder(req.Body).Decode(&input)
	return input, err
}

func validateMethod(method reflect.Value) bool {
	if !method.IsValid() {
		return false
	}
	methodType := method.Type()
	return methodType.NumIn() == 2 && methodType.NumOut() == 2
}

func unmarshalParams(method reflect.Value, params json.RawMessage) (interface{}, error) {
	argType := method.Type().In(1)
	argInstance := reflect.New(argType.Elem()).Interface()
	err := json.Unmarshal(params, argInstance)
	return argInstance, err
}

func callMethod(ctx context.Context, method reflect.Value, argInstance interface{}) (interface{}, error) {
	results := method.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(argInstance),
	})

	if errVal := results[1]; !errVal.IsNil() {
		return nil, errVal.Interface().(error)
	}

	return results[0].Interface(), nil
}

func sendSuccessResponse(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"result": result,
	})
}

func sendErrorResponse(w http.ResponseWriter, msg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}

func Call(ctx context.Context, endpoint, method string, reqPayload, respPayload interface{}) error {
	payload := map[string]interface{}{
		"method": method,
		"params": reqPayload,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := createHTTPRequest(ctx, endpoint, bodyBytes)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return processResponse(resp, respPayload)
}

func createHTTPRequest(ctx context.Context, endpoint string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func processResponse(resp *http.Response, respPayload interface{}) error {
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp.StatusCode, respData)
	}

	return decodeSuccessResponse(respData, respPayload)
}

func handleErrorResponse(statusCode int, respData []byte) error {
	var errPayload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(respData, &errPayload); err == nil && errPayload.Error != "" {
		return errors.New(errPayload.Error)
	}
	return fmt.Errorf("unexpected status code: %d", statusCode)
}

func decodeSuccessResponse(respData []byte, respPayload interface{}) error {
	var rpcResp struct {
		Result interface{} `json:"result"`
		Error  string      `json:"error"`
	}
	rpcResp.Result = respPayload

	if err := json.Unmarshal(respData, &rpcResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResp.Error != "" {
		return errors.New(rpcResp.Error)
	}

	return nil
}
