package api

import (
	"errors"
	"net/http"

	"github.com/vmw-pso/back-end/internal/data"
	"github.com/vmw-pso/back-end/internal/validator"
)

func (api *API) handleCreateResource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID             int64    `json:"id"`
			Name           string   `json:"name"`
			Email          string   `json:"email"`
			JobTitle       string   `json:"jobTitle"`
			Manager        string   `json:"manager"`
			Workgroup      string   `json:"workgroup"`
			Clearance      string   `json:"clearance"`
			Specialties    []string `json:"specialties"`
			Certifications []string `json:"certifications"`
		}

		err := api.readJSON(w, r, &input)
		if err != nil {
			api.badRequestResponse(w, r, err)
			return
		}

		resource := data.Resource{
			ID:             input.ID,
			Name:           input.Name,
			Email:          input.Email,
			JobTitle:       input.JobTitle,
			Manager:        input.Manager,
			Workgroup:      input.Workgroup,
			Clearance:      input.Clearance,
			Specialties:    input.Specialties,
			Certifications: input.Certifications,
			Active:         true,
		}

		v := validator.New()

		if data.ValidateResource(v, resource); !v.Valid() {
			api.failedValidationResponse(w, r, v.Errors)
			return
		}

		err = api.models.Resources.Insert(&resource)
		if err != nil {
			api.serverErrorResponse(w, r, err)
			return
		}

		err = api.writeJSON(w, http.StatusCreated, envelope{"resource": resource}, nil)
		if err != nil {
			api.serverErrorResponse(w, r, err)
		}
	}
}

func (api *API) handleUpdateResource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := api.readIDParam(r)
		if err != nil || id < 1 {
			api.notFoundResponse(w, r)
			return
		}

		resource, err := api.models.Resources.Get(id)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNotFound):
				api.notFoundResponse(w, r)
			default:
				api.serverErrorResponse(w, r, err)
			}
			return
		}

		var input struct {
			Name           *string  `json:"name"`
			Email          *string  `json:"email"`
			JobTitle       *string  `json:"jobTitle"`
			Manager        *string  `json:"manager"`
			Workgroup      *string  `json:"workgroup"`
			Specialties    []string `json:"specialties"`
			Certifications []string `json:"certifications"`
			Active         *bool    `json:"active"`
		}

		err = api.readJSON(w, r, &input)
		if err != nil {
			api.badRequestResponse(w, r, err)
			return
		}

		if input.Name != nil {
			resource.Name = *input.Name
		}

		if input.Email != nil {
			resource.Email = *input.Email
		}

		if input.JobTitle != nil {
			resource.JobTitle = *input.JobTitle
		}

		if input.Manager != nil {
			resource.Manager = *input.Manager
		}

		if input.Workgroup != nil {
			resource.Workgroup = *input.Workgroup
		}

		if input.Specialties != nil {
			resource.Specialties = input.Specialties
		}

		if input.Certifications != nil {
			resource.Certifications = input.Certifications
		}

		if input.Active != nil {
			resource.Active = *input.Active
		}

		v := validator.New()

		if data.ValidateResource(v, *resource); !v.Valid() {
			api.failedValidationResponse(w, r, v.Errors)
			return
		}

		err = api.models.Resources.Update(resource)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				api.editConflictResponse(w, r)
			default:
				api.serverErrorResponse(w, r, err)
			}
			return
		}

		err = api.writeJSON(w, http.StatusOK, envelope{"resource": resource}, nil)
		if err != nil {
			api.serverErrorResponse(w, r, err)
		}
	}
}

func (api *API) handleListResources() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name           string
			Workgroups     []string
			Clearance      string
			Specialties    []string
			Certifications []string
			Manager        string
			Active         bool
			data.Filters
		}

		v := validator.New()

		qs := r.URL.Query()

		input.Name = api.readString(qs, "name", "")
		input.Workgroups = api.readCSV(qs, "workgroups", []string{})
		input.Clearance = api.readString(qs, "clearance", "")
		input.Specialties = api.readCSV(qs, "specialties", []string{})
		input.Certifications = api.readCSV(qs, "certifications", []string{})
		input.Manager = api.readString(qs, "manager", "")
		input.Active = api.readBool(qs, "active", true, v)
		input.Filters.Page = api.readInt(qs, "page", 1, v)
		input.Filters.PageSize = api.readInt(qs, "pageSize", 20, v)
		input.Filters.Sort = api.readString(qs, "sort", "id")
		input.Filters.SortSafelist = []string{"id", "name", "-id", "-name"}

		if data.ValidateFilters(v, input.Filters); !v.Valid() {
			api.failedValidationResponse(w, r, v.Errors)
			return
		}

		resources, metadata, err := api.models.Resources.GetAll(input.Name, input.Workgroups, input.Clearance,
			input.Specialties, input.Certifications, input.Manager, input.Active, input.Filters)
		if err != nil {
			api.serverErrorResponse(w, r, err)
			return
		}

		err = api.writeJSON(w, http.StatusOK, envelope{"resources": resources, "metadata": metadata}, nil)
		if err != nil {
			api.serverErrorResponse(w, r, err)
		}
	}
}
