package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"chay/version"
)

func main() {
	component_data, err := ioutil.ReadFile("cmd/component.json") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	component_string := string(component_data)

	json_convert_data, err := ioutil.ReadFile("cmd/json_convert.json") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	json_convert_string := string(json_convert_data)

	expected, err := version.UpgradeSectionComponent(context.Background(), component_string, json_convert_string)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("cmd/output.json", []byte(expected), 0644)
}
