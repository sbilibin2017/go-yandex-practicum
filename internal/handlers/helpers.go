package handlers

import (
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
)

// parseURLParam извлекает параметры из URL запроса и записывает их в соответствующие
// поля структуры, используя тег `urlparam` для сопоставления.
//
// Параметры:
//   - r: *http.Request — HTTP-запрос, из которого извлекаются параметры URL.
//   - req: interface{} — указатель на структуру, поля которой будут заполнены.
//
// Поведение:
//
//	Функция перебирает все поля структуры, на которую указывает req.
//	Если поле имеет тег `urlparam:"<имя_параметра>"`, то из URL запроса извлекается параметр с таким именем.
//	Если значение параметра не пустое, оно записывается в соответствующее поле структуры (если поле может быть установлено).
//
// Ограничения:
//   - Поддерживаются только поля типа string.
//   - req должен быть указателем на структуру, иначе произойдет паника.
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
