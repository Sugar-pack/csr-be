package docs

import (
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

const SwaggerErrRef = "#/definitions/SwaggerError"

// AddErrorCodeToAllEndpoints adds the error code documentation to all swagger endpoints.
func AddErrorToSecuredEndpoints(code int, response spec.Response, spec *loads.Document) {
	for path, pathItem := range spec.Spec().Paths.Paths {
		operations := GetSecuredOperation(&pathItem.PathItemProps)
		for _, operation := range operations {
			operation.Responses.StatusCodeResponses[code] = response
		}
		spec.Spec().Paths.Paths[path] = pathItem
	}
}

func GetSecuredOperation(pathItemProps *spec.PathItemProps) []*spec.Operation {
	var operations []*spec.Operation
	if pathItemProps.Get != nil && pathItemProps.Get.Security != nil {
		operations = append(operations, pathItemProps.Get)
	}
	if pathItemProps.Put != nil && pathItemProps.Put.Security != nil {
		operations = append(operations, pathItemProps.Put)
	}
	if pathItemProps.Post != nil && pathItemProps.Post.Security != nil {
		operations = append(operations, pathItemProps.Post)
	}
	if pathItemProps.Delete != nil && pathItemProps.Delete.Security != nil {
		operations = append(operations, pathItemProps.Delete)
	}
	if pathItemProps.Options != nil && pathItemProps.Options.Security != nil {
		operations = append(operations, pathItemProps.Options)
	}
	if pathItemProps.Head != nil && pathItemProps.Head.Security != nil {
		operations = append(operations, pathItemProps.Head)
	}
	if pathItemProps.Patch != nil && pathItemProps.Patch.Security != nil {
		operations = append(operations, pathItemProps.Patch)
	}
	return operations
}

// UnauthorizedError returns the error code and response for unauthorized error.
func UnauthorizedError() (int, spec.Response) {
	return http.StatusUnauthorized, spec.Response{
		// Keep in mind that this response will be used by frontend team.
		// So, if you change it, you should notify them.
		ResponseProps: spec.ResponseProps{
			Description: "Unauthorized",
			Schema:      SwaggerErrorReference(),
		},
	}
}

func ForbiddenError() (int, spec.Response) {
	return http.StatusForbidden, spec.Response{
		// Keep in mind that this response will be used by frontend team.
		// So, if you change it, you should notify them.
		ResponseProps: spec.ResponseProps{
			Description: "Forbidden",
			Schema:      SwaggerErrorReference(),
		},
	}
}

func SwaggerErrorReference() *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Ref: spec.MustCreateRef(SwaggerErrRef),
		},
	}
}
