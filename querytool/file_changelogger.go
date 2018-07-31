package querytool

import (
	"io/ioutil"
	"encoding/json"
	"crypto/sha256"
	"strings"
	"fmt"
)

// takes an argument to the path to the hash table json, and the file to be compared against
func HasFileChanged(hashtable, file string) (string, error) {
	json, err := JsonToMap(hashtable)
	if err != nil {
		return "",err
	}

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return "",err
	}

	h := sha256.New()
	h.Write([]byte(contents))
	b := h.Sum(nil)
	str := fmt.Sprintf("%x",b)

	if _, ok := json[file]; !ok {
		json[file] = str
		err := MapToJson(hashtable,json)

		if err != nil{
			return "", err
		}
		return strings.Join([]string{"Created new hash entry for ",file},""), nil
	} else if ok && json[file] != str {

		json[file] = str
		err := MapToJson(hashtable,json)

		if err != nil{
			return "",err
		}
		return strings.Join([]string{"Updated hash entry for ",file},""), nil
	} else {
		return "",nil
	}
}

// general function to read a json and unmarshall into a map
func JsonToMap(filepath string) (map[string]string, error) {

	contents, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	var jsonmap map[string]*json.RawMessage
	err = json.Unmarshal(contents, &jsonmap)
	if err != nil{
		return nil, err
	}

	configmap := make(map[string]string)

	for k,_ := range jsonmap {
		var str string
		err  = json.Unmarshal(*jsonmap[k],&str)

		if err != nil{
			return nil, err
		}

		configmap[k] = str
	}

	return configmap, nil
}

// general function to take a map and marshall to json then write
func MapToJson(filepath string, jsonstr map[string]string) error {
	b, err := json.Marshal(jsonstr)
	if err != nil {
		return err
	}

	ioutil.WriteFile(filepath, b, 0)
	return nil
}
