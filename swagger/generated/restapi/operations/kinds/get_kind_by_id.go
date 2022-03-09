// Code generated by go-swagger; DO NOT EDIT.

package kinds

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetKindByIDHandlerFunc turns a function with the right signature into a get kind by ID handler
type GetKindByIDHandlerFunc func(GetKindByIDParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetKindByIDHandlerFunc) Handle(params GetKindByIDParams) middleware.Responder {
	return fn(params)
}

// GetKindByIDHandler interface for that can handle valid get kind by ID params
type GetKindByIDHandler interface {
	Handle(GetKindByIDParams) middleware.Responder
}

// NewGetKindByID creates a new http.Handler for the get kind by ID operation
func NewGetKindByID(ctx *middleware.Context, handler GetKindByIDHandler) *GetKindByID {
	return &GetKindByID{Context: ctx, Handler: handler}
}

/* GetKindByID swagger:route GET /equipment/kinds/{kindId} Kinds getKindById

Get information about the kind of equipment by id.

*/
type GetKindByID struct {
	Context *middleware.Context
	Handler GetKindByIDHandler
}

func (o *GetKindByID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetKindByIDParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
