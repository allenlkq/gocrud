package entities

import "reflect"

type Structure struct {
	Name       string    `json:"name"`
	Definition map[string]string    `json:"definition"`
}

var StructureRegistry = make(map[string]reflect.Type)


