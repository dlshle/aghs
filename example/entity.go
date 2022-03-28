package example

import "encoding/json"

type Student struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"`
}

func UnmarshalStudent(data []byte) (student Student, err error) {
	err = json.Unmarshal(data, &student)
	return
}
