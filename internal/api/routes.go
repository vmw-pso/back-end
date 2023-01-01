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

	router.HandlerFunc(http.MethodGet, "/v1/projects", api.handleListProjects())
	router.HandlerFunc(http.MethodPost, "/v1/projects", api.handleCreateProject())
	router.HandlerFunc(http.MethodGet, "/v1/projects/:id", api.handleShowProject())
	router.HandlerFunc(http.MethodPatch, "/v1/projects/:id", api.handleUpdateProject())

	return api.recoverPanic(api.enableCORS(router))
}
