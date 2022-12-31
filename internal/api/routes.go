package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (api *API) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", api.handleHealthcheck())

	router.HandlerFunc(http.MethodGet, "/v1/resources", api.handleListResources())
	router.HandlerFunc(http.MethodPost, "/v1/resources", api.handleCreateResource())
	router.HandlerFunc(http.MethodPatch, "/v1/resources/:id", api.handleUpdateResource())

	return api.recoverPanic(api.enableCORS(router))
}
