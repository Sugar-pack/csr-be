// Code generated by go-swagger; DO NOT EDIT.

package equipment

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

// NewFindEquipmentParams creates a new FindEquipmentParams object
//
// There are no default values defined in the spec.
func NewFindEquipmentParams() FindEquipmentParams {

	return FindEquipmentParams{}
}

// FindEquipmentParams contains all the bound params for the find equipment operation
// typically these are obtained from a http.Request
//
// swagger:parameters FindEquipment
type FindEquipmentParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Filtered list of an equipment
	  In: body
	*/
	FindEquipment *models.Equipment
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewFindEquipmentParams() beforehand.
func (o *FindEquipmentParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.Equipment
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			res = append(res, errors.NewParseError("findEquipment", "body", "", err))
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
				o.FindEquipment = &body
			}
		}
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
