package util

import (
	"encoding/json"
	"fmt"
)

func PrintJson(obj interface{}) {
	// Marshal the data with indentation
	prettyJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the indented JSON
	fmt.Println(string(prettyJSON))

}
