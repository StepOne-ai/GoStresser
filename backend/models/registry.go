package models

var registeredModels []any

func Register(model any) {
    registeredModels = append(registeredModels, model)
}

func GetRegisteredModels() []any {
    return registeredModels
}