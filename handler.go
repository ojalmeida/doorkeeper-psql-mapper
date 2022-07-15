package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func (pm PsqlMapper) Handle(r *http.Request, w http.ResponseWriter) (err error) {

	var status int
	var msg string
	var data json.RawMessage

	defer func() {

		var out []byte

		if status != 200 {

			w.WriteHeader(status)

		}

		out, err = json.Marshal(response{
			Status: status,
			Msg:    msg,
			Data:   data,
		})

		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return

		}

		_, err = w.Write(out)

	}()

	status = http.StatusNotFound
	msg = "not found"

	for _, b := range pm.behaviors {

		if b.pathMapping.match(r.URL.RequestURI()) {

			params := map[string]string{}

			{
				err = r.ParseForm()
				if err != nil {
					return
				}

				for k := range r.Form {

					params[k] = r.Form.Get(k)

				}
			} // params population

			switch r.Method {

			case http.MethodGet:

				if hasID(b.pathMapping.path, r.URL.RequestURI()) {

					var result json.RawMessage

					pathSlice := strings.Split(r.URL.Path, "/")

					id := pathSlice[len(pathSlice)-1]

					err, result = getByID(r.Context(), pm.conn, b, id)

					if err != nil {

						pm.error.Println(err.Error())
						status = http.StatusInternalServerError
						msg = err.Error()

						return

					}

					status = http.StatusOK
					msg = ""
					data = result

				} else {

					var result json.RawMessage

					err, result = get(r.Context(), pm.conn, b, params)

					if err != nil {

						pm.error.Println(err.Error())
						status = http.StatusInternalServerError
						msg = err.Error()

						return

					}

					status = http.StatusOK
					msg = ""
					data = result

				}

			case http.MethodPost:

				var result json.RawMessage

				err, result = create(r.Context(), pm.conn, b, params)

				if err != nil {

					if errors.Is(err, ErrDuplicateKey) {

						status = http.StatusBadRequest
						msg = err.Error()
						return

					}

					pm.error.Println(err.Error())
					status = http.StatusInternalServerError
					msg = err.Error()

					return

				}

				status = http.StatusCreated
				msg = ""
				data = result

			case http.MethodPut:

				if !hasID(b.pathMapping.path, r.URL.RequestURI()) {

					status = http.StatusBadRequest
					msg = "missing item identifier"
					return

				}

				pathSlice := strings.Split(r.URL.Path, "/")

				id := pathSlice[len(pathSlice)-1]

				var result json.RawMessage

				err, result = updateByID(r.Context(), pm.conn, b, id, params)

				switch {

				case errors.Is(err, ErrDuplicateKey):

					status = http.StatusBadRequest
					msg = err.Error()
					return

				case errors.Is(err, ErrIdentifierUpdate):

					status = http.StatusBadRequest
					msg = err.Error()
					return
				}

				status = http.StatusOK
				msg = ""
				data = result

			case http.MethodDelete:

				if !hasID(b.pathMapping.path, r.URL.RequestURI()) {

					status = http.StatusBadRequest
					msg = "missing item identifier"
					return

				}

				pathSlice := strings.Split(r.URL.Path, "/")

				id := pathSlice[len(pathSlice)-1]

				var result json.RawMessage

				err = deleteByID(r.Context(), pm.conn, b, id)

				if err != nil {

					if errors.Is(err, ErrNotFound) {

						status = http.StatusNotFound
						msg = fmt.Sprintf("item not found")
						return

					}

					pm.error.Println(err.Error())
					status = http.StatusInternalServerError
					msg = err.Error()

					return

				}

				status = http.StatusNoContent
				msg = ""
				data = result

			case http.MethodOptions:

				w.Header().Set("Allow", "GET, POST, PUT, DELETE, OPTIONS")
				status = http.StatusOK

			}

			break

		}

	}

	return
}

func hasID(path string, uri string) bool {

	idRegexp := regexp.MustCompile(fmt.Sprintf(`%s/\w+$`, path))

	return idRegexp.MatchString(uri)

}
