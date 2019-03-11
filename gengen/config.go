package gengen

import (
	"io/ioutil"
	"reflect"
	"time"

	hjson "github.com/hjson/hjson-go"
	"github.com/mitchellh/mapstructure"
)

func readConfig(filename string) (map[string]interface{}, error) {
	switch filename {
	case "@beego", "@bee", "@beego.json", "@bee.json":
		return beeConfig, nil
	case "@gin", "@gin.json":
		return ginConfig, nil
	case "@echo", "@echo.json":
		return echoConfig, nil
	case "@loong", "@loong.json":
		return loongConfig, nil
	}

	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var values map[string]interface{}
	if err := hjson.Unmarshal(bs, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func readStyleConfig(filename string) (*DefaultStye, error) {
	values, err := readConfig(filename)
	if err != nil {
		return nil, err
	}

	style := NewEchoStye()
	if err := toStruct(style, values); err != nil {
		return nil, err
	}

	style.reinit(values)
	return style, nil
}

func readImports(row map[string]interface{}) map[string]string {
	if row == nil {
		return nil
	}
	o := row["imports"]
	if o == nil {
		return nil
	}

	switch v := o.(type) {
	case map[string]string:
		return v
	case map[string]interface{}:
		result := map[string]string{}
		for key, value := range v {
			if s, ok := value.(string); ok && s != "" {
				result[key] = s
			}
		}
		return result
	default:
		return nil
	}
}

func decodeHook(from reflect.Kind, to reflect.Kind, v interface{}) (interface{}, error) {
	if from == reflect.String && to == reflect.Bool {
		s := v.(string)
		return s == "on" || s == "yes" || s == "enabled", nil
	}
	return v, nil
}

func toStruct(rawVal interface{}, row map[string]interface{}) (err error) {
	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(decodeHook,
			stringToTimeHookFunc(time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02 15:04:05Z07:00",
				"2006-01-02 15:04:05",
				"2006-01-02")),
		Metadata:         nil,
		Result:           rawVal,
		TagName:          "json",
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(row)
}

func stringToTimeHookFunc(layouts ...string) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}
		s := data.(string)
		if s == "" {
			return time.Time{}, nil
		}
		for _, layout := range layouts {
			t, err := time.Parse(layout, s)
			if err == nil {
				return t, nil
			}
		}
		// Convert it by parsing
		return data, nil
	}
}
