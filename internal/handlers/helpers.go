package handlers

import (
	"net/http"
	reflect "reflect"

	"github.com/go-chi/chi/v5"
)

func parseURLParam(r *http.Request, req interface{}) {
	val := reflect.ValueOf(req).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		urlparamTag := fieldType.Tag.Get("urlparam")
		if urlparamTag == "" {
			continue
		}

		paramValue := chi.URLParam(r, urlparamTag)
		if paramValue != "" {
			if field.CanSet() {
				field.SetString(paramValue)
			}
		}
	}

}
