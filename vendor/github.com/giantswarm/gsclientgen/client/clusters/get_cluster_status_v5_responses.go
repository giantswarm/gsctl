// Code generated by go-swagger; DO NOT EDIT.

package clusters

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/giantswarm/gsclientgen/models"
)

// GetClusterStatusV5Reader is a Reader for the GetClusterStatusV5 structure.
type GetClusterStatusV5Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetClusterStatusV5Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetClusterStatusV5OK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 401:
		result := NewGetClusterStatusV5Unauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewGetClusterStatusV5Default(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetClusterStatusV5OK creates a GetClusterStatusV5OK with default headers values
func NewGetClusterStatusV5OK() *GetClusterStatusV5OK {
	return &GetClusterStatusV5OK{}
}

/*GetClusterStatusV5OK handles this case with default header values.

Cluster status
*/
type GetClusterStatusV5OK struct {
	Payload *models.V5GetClusterStatusResponse
}

func (o *GetClusterStatusV5OK) Error() string {
	return fmt.Sprintf("[GET /v5/clusters/{cluster_id}/status/][%d] getClusterStatusV5OK  %+v", 200, o.Payload)
}

func (o *GetClusterStatusV5OK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.V5GetClusterStatusResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetClusterStatusV5Unauthorized creates a GetClusterStatusV5Unauthorized with default headers values
func NewGetClusterStatusV5Unauthorized() *GetClusterStatusV5Unauthorized {
	return &GetClusterStatusV5Unauthorized{}
}

/*GetClusterStatusV5Unauthorized handles this case with default header values.

Permission denied
*/
type GetClusterStatusV5Unauthorized struct {
	Payload *models.V4GenericResponse
}

func (o *GetClusterStatusV5Unauthorized) Error() string {
	return fmt.Sprintf("[GET /v5/clusters/{cluster_id}/status/][%d] getClusterStatusV5Unauthorized  %+v", 401, o.Payload)
}

func (o *GetClusterStatusV5Unauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.V4GenericResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetClusterStatusV5Default creates a GetClusterStatusV5Default with default headers values
func NewGetClusterStatusV5Default(code int) *GetClusterStatusV5Default {
	return &GetClusterStatusV5Default{
		_statusCode: code,
	}
}

/*GetClusterStatusV5Default handles this case with default header values.

error
*/
type GetClusterStatusV5Default struct {
	_statusCode int

	Payload *models.V4GenericResponse
}

// Code gets the status code for the get cluster status v5 default response
func (o *GetClusterStatusV5Default) Code() int {
	return o._statusCode
}

func (o *GetClusterStatusV5Default) Error() string {
	return fmt.Sprintf("[GET /v5/clusters/{cluster_id}/status/][%d] getClusterStatusV5 default  %+v", o._statusCode, o.Payload)
}

func (o *GetClusterStatusV5Default) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.V4GenericResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
