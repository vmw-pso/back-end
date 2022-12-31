package api

import "net/http"

func (api *API) errorLog(r *http.Request, err error) {
	api.logger.PrintError(err, map[string]string{
		"request-method": r.Method,
		"request_urL":    r.URL.String(),
	})
}

func (api *API) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}
	err := api.writeJSON(w, status, env, nil)
	if err != nil {
		api.errorLog(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (api *API) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	api.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (api *API) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	api.errorResponse(w, r, http.StatusConflict, message)
}

func (api *API) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	api.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (api *API) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	api.errorResponse(w, r, http.StatusNotFound, message)
}

func (api *API) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	api.errorLog(r, err)

	message := "the server encounter a problem and could not process the request"
	api.errorResponse(w, r, http.StatusInternalServerError, message)
}
