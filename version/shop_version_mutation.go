package version

import (
	"fmt"
	"reflect"
	"strings"

	"chay/json"

	"github.com/dop251/goja"
)

type Action string

const (
	Add       Action = "ADD"
	Remove    Action = "REMOVE"
	Update    Action = "UPDATE"
	ChangeTag Action = "CHANGE_TAG"
)

type Mutation struct {
	Action    Action      `json:"action,omitempty"`
	Name      string      `json:"name,omitempty"`
	NewName   string      `json:"newName,omitempty"`
	Value     any         `json:"value,omitempty"`
	ValueFrom string      `json:"valueFrom,omitempty"`
	JoinValue []string    `json:"joinValue,omitempty"`
	TypeFrom  string      `json:"typeFrom,omitempty"`
	Type      string      `json:"type,omitempty"`
	Condition *Condition  `json:"condition,omitempty"`
	Fields    []*Mutation `json:"fields,omitempty"`
	Params    []string    `json:"params"`
	Operation string      `json:"operation"`
}

func (m Mutation) value(data map[string]any) (any, error) {
	var (
		value any
		err   error
	)
	switch {
	case m.Value != nil:
		value = m.Value
	case len(m.ValueFrom) != 0:
		value, err = m.valueFrom(data)
	case len(m.JoinValue) != 0:
		value, err = m.joinValues(data)
	case len(m.Operation) != 0:
		value, err = m.calculateValue(data)
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

// get the value of an attribute in property
// data: { settings { background { desktop {attachment : "123px"}}}}
// valueFrom: "settings.background.desktop.attachment"
// result: "123px" || support any type
func (m Mutation) valueFrom(data map[string]any) (any, error) {
	return json.GetValueFromJSONMap(strings.Join([]string{m.TypeFrom, m.ValueFrom}, "."), data)
}

// Combine all values of 1 or more attributes
// data: { settings { background { desktop {attachment : "123px", iconSpacing: "12px", color: "#AE0000"}}}}
// joinValues: ["settings.background.desktop.attachment", "settings.background.desktop.iconSpacing", "settings.background.desktop.color"]
// result: "123px/12px/#AE0000" || just support for string value
func (m Mutation) joinValues(data map[string]any) (string, error) {
	valuesStr := make([]string, 0, len(m.JoinValue))
	values, err := json.GetValuesFromJSONMap(m.JoinValue, data)
	if err != nil {
		return "", err
	}

	for _, value := range values {
		if value != nil && reflect.TypeOf(value).Kind() == reflect.String {
			valuesStr = append(valuesStr, value.(string))
		}
	}

	return strings.Join(valuesStr, "/"), nil
}

// calculate value using operation and params ( support js function )
// operation: "settings.desktop.height + settings.desktop.width" params: "setting": { "desktop": { "width": 50, "height": 60}
// result: 110
func (m Mutation) calculateValue(data map[string]any) (any, error) {
	vm := goja.New()
	if m.Params != nil {
		for i, param := range m.Params {
			paramName := fmt.Sprintf("param%v", i)
			m.Operation = strings.ReplaceAll(m.Operation, param, paramName)
			value, err := json.GetValueFromJSONMap(param, data)
			if err != nil {
				return nil, err
			}
			err = vm.Set(paramName, value)
			if err != nil {
				return nil, fmt.Errorf("failed to set param to vm %w", err)
			}
		}
	}
	runString, err := vm.RunString(m.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate operation: %w", err)
	}
	return runString.Export(), nil
}
