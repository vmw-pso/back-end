package api

import "net/http"

func (api *API) handleHealthcheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		env := envelope{
			"status": "available",
			"system_info": map[string]string{
				"environment": api.cfg.Env,
				"version":     version,
			},
		}

		err := api.writeJSON(w, http.StatusOK, env, nil)
		if err != nil {
			api.serverErrorResponse(w, r, err)
		}
	}
}
