/*
 * Package "controllers" provides Dots API controllers.
 */
package controllers

import (
	"github.com/nttdots/go-dots/dots_common"
	"github.com/nttdots/go-dots/dots_server/models"
	"github.com/nttdots/go-dots/libcoap"
)

/*
 * Controller API interface.
 * It provides function interfaces correspondent to CoAP methods.
 */
type ControllerInterface interface {
	HandleGet(Request, *models.Customer) (Response, error)
	HandlePost(Request, *models.Customer) (Response, error)
	HandleDelete(Request, *models.Customer) (Response, error)
	HandlePut(Request, *models.Customer) (Response, error)

	Get(interface{}, *models.Customer) (Response, error)
	Post(interface{}, *models.Customer) (Response, error)
	Delete(interface{}, *models.Customer) (Response, error)
	Put(interface{}, *models.Customer) (Response, error)
}

type ServiceMethod func(req Request, customer *models.Customer) (Response, error)

/*
 * Controller super class.
 */
type Controller struct {
}

func (c *Controller) HandleGet(req Request, customer *models.Customer) (Response, error) {
	return c.Get(req.Body, customer)
}

func (c *Controller) HandlePost(req Request, customer *models.Customer) (Response, error) {
	return c.Post(req.Body, customer)
}

func (c *Controller) HandleDelete(req Request, customer *models.Customer) (Response, error) {
	return c.Delete(req.Body, customer)
}

func (c *Controller) HandlePut(req Request, customer *models.Customer) (Response, error) {
	return c.Put(req.Body, customer)
}

/*
 * Handles CoAP Get requests.
 */
func (c *Controller) Get(interface{}, *models.Customer) (Response, error) {
	return Response{
		Code: dots_common.MethodNotAllowed,
		Type: dots_common.NonConfirmable,
	}, nil
}

/*
 * Handles CoAP Post requests.
 */
func (c *Controller) Post(interface{}, *models.Customer) (Response, error) {
	return Response{
		Code: dots_common.MethodNotAllowed,
		Type: dots_common.NonConfirmable,
	}, nil
}

/*
 * Handles CoAP Delete requests.
 */
func (c *Controller) Delete(interface{}, *models.Customer) (Response, error) {
	return Response{
		Code: dots_common.MethodNotAllowed,
		Type: dots_common.NonConfirmable,
	}, nil
}

/*
 * Handles CoAP Put requests.
 */
func (c *Controller) Put(interface{}, *models.Customer) (Response, error) {
	return Response{
		Code: dots_common.MethodNotAllowed,
		Type: dots_common.NonConfirmable,
	}, nil
}

/*
 * Regular API request
 */
type Request struct {
	Code    libcoap.Code
	Type    libcoap.Type
	Uri     []string
	Queries []string
	Body    interface{}
}

/*
 * Regular API response
 */
type Response struct {
	Code dots_common.Code
	Type dots_common.Type
	Body interface{}
}

/*
 * Error API response
 */
type Error struct {
	Code dots_common.Code
	Type dots_common.Type
	Body interface{}
}

func (e Error) Error() string {
	return e.Code.String()
}
