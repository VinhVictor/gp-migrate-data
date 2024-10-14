package version

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	ThemeStylePropertyKey = "ThemeStyle"
	Children              = "childrens"
	Tag                   = "tag"
)

func UpgradeSectionComponent(ctx context.Context, component string, dataChange string) (string, error) {
	var (
		oldDataMap    map[string]any
		dataChangeMap map[string]any
		err           error
	)
	err = json.Unmarshal([]byte(component), &oldDataMap)
	if err != nil {
		return "", fmt.Errorf("unmarshal section component: %w", err)
	}

	err = json.Unmarshal([]byte(dataChange), &dataChangeMap)
	if err != nil {
		return "", fmt.Errorf("unmarshal data change: %w", err)
	}

	for k, v := range dataChangeMap {
		// if it is theme style then skip
		// because if dataChangeMap has themeStyle, it will be upgraded in the upgradeThemeStyle function
		if k == ThemeStylePropertyKey {
			continue
		}

		mutations, err := decodeMutations(ctx, v)
		if err != nil {
			return "", fmt.Errorf("decodeMutations: %w", err)
		}

		err = mutateData(ctx, oldDataMap, k, mutations)
		if err != nil {
			return "", fmt.Errorf("update section component: %w", err)
		}
	}

	changedData, err := json.Marshal(oldDataMap)
	if err != nil {
		return "",  fmt.Errorf("marshal section component: %w", err)
	}

	return string(changedData), nil
}

func UpgradeThemeStyle(ctx context.Context, component string, dataChange string) (string, error) {
	var (
		oldDataMap    map[string]any
		dataChangeMap map[string]any
		err           error
	)


	err = json.Unmarshal([]byte(component), &oldDataMap)
	if err != nil {
		return "", fmt.Errorf("unmarshal section component: %w", err)
	}

	err = json.Unmarshal([]byte(dataChange), &dataChangeMap)
	if err != nil {
		return "", fmt.Errorf("unmarshal next version oldDataMap: %w", err)
	}

	if themeStyle, exists := dataChangeMap[ThemeStylePropertyKey]; exists && themeStyle != nil {
		mutations, err := decodeMutations(ctx, themeStyle)
		if err != nil {
			return "", fmt.Errorf("decodeMutations: %w", err)
		}

		for _, mutation := range mutations {
			currentData := oldDataMap
			isRootMutation := true
			if currentData == nil {
				return "", fmt.Errorf("currentData is nil")
			}
			err := DoMutation(oldDataMap, currentData, &mutation, isRootMutation)
			if err != nil {
				return "", fmt.Errorf("do mutation: %w", err)
			}
		}
	}

	changedData, err := json.Marshal(oldDataMap)
	if err != nil {
		return "", fmt.Errorf("marshal section component: %w", err)
	}
	return string(changedData), nil
}

/*
mutateData check tag of mutations and update data if data have same tag as tagName
@params:
  - data: map[string]any
  - tagName: tag
  - dataChange: raw data change including conditions, the new value of one or multiple properties that need to be changed
  - mutations: that is converted from data change. We want to pass it to avoid that re-convert data change

@Return:
  - error
  - update value of data
*/
func mutateData(ctx context.Context, data map[string]any, tagName string, mutations []Mutation) error {
	var (
		children []any
		err      error
	)
	if data[Tag] == tagName {
		for _, mutation := range mutations {
			if data == nil {
				return fmt.Errorf("data is nil")
			}
			err := DoMutation(data, data, &mutation, true)
			if err != nil {
				return fmt.Errorf("do mutation: %w", err)
			}
		}
	}
	if c, exists := data[Children]; exists && c != nil {
		children, _ = c.([]any)
	}

	for i := range children {
		if child, ok := children[i].(map[string]any); ok && child != nil {
			err = mutateData(ctx, child, tagName, mutations)
			if err != nil {
				return fmt.Errorf("update child section component: %w", err)
			}
		}
	}

	return nil
}
/*
DoMutation : change key or value of data
@params:
  - componentData: section component data to get value (be used when mutations have valueFrom, valueJoin, params...)
  - currentData: data or child of data to process mutations
  - mutation: changes applied to current data ( have 4 type: ADD, REMOVE, UPDATE, CHANGE_TAG)
*/
func DoMutation(componentData, currentData map[string]any, mutation *Mutation, isRootMutation bool) error {
	var (
		err      error
		typeName string
		ok       bool
	)
	if !EvaluateCondition(componentData, mutation.Condition) {
		return nil
	}
	if isRootMutation {
		typeName = mutation.Type
	}
	if currentData == nil {
		return fmt.Errorf("currentData is nil")
	}
	fieldName := mutation.Name
	switch mutation.Action {
	case Add:
		currentData, ok = getTypeData(typeName, currentData, Add)
		if !ok {
			return fmt.Errorf("unable to parse currentData[%s] for %s action", typeName, Add)
		}
		currentData[fieldName] = map[string]any{}
		value, _ := mutation.value(componentData)
		if value != nil {
			currentData[fieldName] = value
		}

		err = DoChildMutations(componentData, currentData, mutation)
		if err != nil {
			return fmt.Errorf("do child mutations: %w", err)
		}

	case Remove:
		currentData, ok = getTypeData(typeName, currentData, Remove)
		if !ok {
			break
		}
		delete(currentData, mutation.Name)

	case Update:
		currentData, ok = getTypeData(typeName, currentData, Update)
		if !ok {
			break
		}

		err = DoChildMutations(componentData, currentData, mutation)
		if err != nil {
			return fmt.Errorf("do child mutations: %w", err)
		}

		if _, ok := currentData[fieldName]; ok {
			value, _ := mutation.value(componentData)
			if value != nil {
				currentData[fieldName] = value
			}
		}

		if len(mutation.NewName) != 0 {
			currentData[mutation.NewName] = currentData[fieldName]
			delete(currentData, fieldName)
		}

	case ChangeTag:
		currentData[Tag] = mutation.NewName

	default:
	}
	return nil
}

func getTypeData(typeName string, data map[string]any, action Action) (map[string]any, bool) {
	if typeName == "" {
		return data, true
	}

	v, ok := data[typeName]
	if !ok || v == nil {
		if action == Update || action == Remove {
			return nil, false
		}
		data[typeName] = make(map[string]any)
	}
	data, ok = data[typeName].(map[string]any)
	return data, ok
}

func DoChildMutations(data, currentData map[string]any, mutation *Mutation) error {
	fieldName := mutation.Name
	for _, childMutation := range mutation.Fields {
		if childData, ok := currentData[fieldName].(map[string]any); ok && childData != nil {
			isRootMutation := false
			err := DoMutation(data, childData, childMutation, isRootMutation)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeMutations(ctx context.Context, input any) ([]Mutation, error) {
	var mutations []Mutation
	mutationBytes, err := json.Marshal(input)
	if err != nil {
		return nil,  fmt.Errorf("json marshal mutations: %w", err)
	}
	err = json.Unmarshal(mutationBytes, &mutations)
	if err != nil {
		return nil, fmt.Errorf("unmarshal mutations: %w", err)
	}
	return mutations, nil
}
