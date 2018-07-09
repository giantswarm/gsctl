// Code generated by go-swagger; DO NOT EDIT.

package organizations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/giantswarm/gsclientgen/models"
)

// GetOrganizationsReader is a Reader for the GetOrganizations structure.
type GetOrganizationsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetOrganizationsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetOrganizationsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewGetOrganizationsDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetOrganizationsOK creates a GetOrganizationsOK with default headers values
func NewGetOrganizationsOK() *GetOrganizationsOK {
	return &GetOrganizationsOK{}
}

/*GetOrganizationsOK handles this case with default header values.

Success
*/
type GetOrganizationsOK struct {
	Payload []*models.V4OrganizationListItem
}

func (o *GetOrganizationsOK) Error() string {
	return fmt.Sprintf("[GET /v4/organizations/][%d] getOrganizationsOK  %+v", 200, o.Payload)
}

func (o *GetOrganizationsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetOrganizationsDefault creates a GetOrganizationsDefault with default headers values
func NewGetOrganizationsDefault(code int) *GetOrganizationsDefault {
	return &GetOrganizationsDefault{
		_statusCode: code,
	}
}

/*GetOrganizationsDefault handles this case with default header values.

Error
*/
type GetOrganizationsDefault struct {
	_statusCode int

	Payload *models.V4GenericResponse
}

// Code gets the status code for the get organizations default response
func (o *GetOrganizationsDefault) Code() int {
	return o._statusCode
}

func (o *GetOrganizationsDefault) Error() string {
	return fmt.Sprintf("[GET /v4/organizations/][%d] getOrganizations default  %+v", o._statusCode, o.Payload)
}

func (o *GetOrganizationsDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.V4GenericResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
