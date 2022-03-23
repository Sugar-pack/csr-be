// Code generated by go-swagger; DO NOT EDIT.

package kinds

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

// NewPatchKindParams creates a new PatchKindParams object
//
// There are no default values defined in the spec.
func NewPatchKindParams() PatchKindParams {

	return PatchKindParams{}
}

// PatchKindParams contains all the bound params for the patch kind operation
// typically these are obtained from a http.Request
//
// swagger:parameters PatchKind
type PatchKindParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*kind id
	  Required: true
	  In: path
	*/
	KindID string
	/*
	  Required: true
	  In: body
	*/
	PatchTask *models.PatchTask
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPatchKindParams() beforehand.
func (o *PatchKindParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rKindID, rhkKindID, _ := route.Params.GetOK("kindId")
	if err := o.bindKindID(rKindID, rhkKindID, route.Formats); err != nil {
		res = append(res, err)
	}

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.PatchTask
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("patchTask", "body", ""))
			} else {
				res = append(res, errors.NewParseError("patchTask", "body", "", err))
			}
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			ctx := validate.WithOperationRequest(context.Background())
			if err := body.ContextValidate(ctx, route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.PatchTask = &body
			}
		}
	} else {
		res = append(res, errors.Required("patchTask", "body", ""))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindKindID binds and validates parameter KindID from path.
func (o *PatchKindParams) bindKindID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.KindID = raw

	return nil
}
