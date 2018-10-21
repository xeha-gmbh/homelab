package shared

import (
	"encoding/json"
	"strings"
)

var (
	successWithNothing = func(f CombinedJsonOutputHandler) (interface{}, error) {
		return nil, nil
	}
)

type CombinedJsonOutputHandler func(map[string]interface{}) (interface{}, error)

type OutputProcessor func(CombinedJsonOutputHandler) (interface{}, error)

type JsonHandler func(raw []byte, err error) OutputProcessor

func HandleOutput(printer MessagePrinter) JsonHandler {
	return func(raw []byte, err error) OutputProcessor {
		printer.Debug(string(raw), map[string]interface{}{})
		return HandledJson(raw, err)
	}
}

func HandledJson(raw []byte, err error) func(f CombinedJsonOutputHandler) (interface{}, error) {
	if len(raw) == 0 {
		if err == nil {
			return successWithNothing
		} else {
			return func(f CombinedJsonOutputHandler) (interface{}, error) {
				return nil, err
			}
		}
	}

	r := make(map[string]interface{})
	_ = json.Unmarshal([]byte(lastOutput(raw)), &r)
	return func(f CombinedJsonOutputHandler) (interface{}, error) {
		return f(r)
	}
}

func lastOutput(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	return s[strings.LastIndexByte(s, '\n')+1:]
}
