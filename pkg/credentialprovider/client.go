package credentialprovider

/*
#cgo LDFLAGS: -lcpasswordsdk
#include "client.h"
*/
import "C"

import (
	"fmt"
)

type Client struct {
	Request  C.ObjectHandle
	Response C.ObjectHandle
}

type Properties struct {
	RequestName         string
	SafeName            string
	AppID               string
	ObjectName          string            // Aka AccountName
	RequestedAttributes []string          // List of "PassProps.*" to fetch
	Attributes          map[string]string // holds the values of the requested attributes
}

func NewProperties(reqname string, safe string, appid string, objname string, attrs []string) Properties {
	props := Properties{
		RequestName: reqname,
		SafeName:    safe,
		AppID:       appid,
		ObjectName:  objname,
	}
	for i := 0; i < len(attrs); i++ {
		props.RequestedAttributes = append(props.RequestedAttributes, attrs[i])
	}
	props.Attributes = make(map[string]string)
	return props
}
func (cl *Client) FetchProperties(props Properties) (Properties, error) {
	for _, attr := range props.RequestedAttributes {
		val, err := cl.FetchProperty(props, attr)
		if err == nil {
			props.Attributes[attr] = val
		}
	}
	return props, nil
}

func (cl *Client) FetchProperty(props Properties, attr string) (string, error) {
	err := cl.CreateRequest(props.RequestName)
	if err != nil {
		return "", err
	}
	err = cl.SetAttribute("AppDescs.AppID", props.AppID)
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf("Safe=%s;Object=%s", props.SafeName, props.ObjectName)
	err = cl.SetAttribute("Query", query)
	if err != nil {
		return "", err
	}
	err = cl.SetAttribute("FailRequestOnPasswordChange", "true")
	if err != nil {
		return "", err
	}
	err = cl.GetPassword()
	if err != nil {
		return "", err
	}
	return cl.GetAttribute(attr)
}

func (cl *Client) CreateRequest(req string) error {
	request := C.CString(req)
	cl.Request = C.PSDK_CreateRequest(request)
	check := C.ErrorCheck(cl.Request)
	if check == nil {
		return nil
	}
	msg := C.GoString(check)
	err := fmt.Errorf("%s", msg)
	return err
}
func (cl *Client) FreeRequest() {
	C.ReleaseHandle(cl.Request)
}
func (cl *Client) FreeResponse() {
	C.ReleaseHandle(cl.Response)
}
func (cl *Client) SetAttribute(name string, value string) error {
	n := C.CString(name)
	v := C.CString(value)
	result := C.PSDK_SetAttribute(cl.Request, n, v)
	if result == C.PSDK_Get_RC_ERROR() {
		return fmt.Errorf("failed to set attribute")
	}
	return nil
}
func (cl *Client) GetPassword() error {
	result := C.GetPassword(cl.Request, &cl.Response)
	if result == nil {
		return nil
	}
	msg := C.GoString(result)
	err := fmt.Errorf("%s", msg)
	return err
}
func (cl *Client) GetAttribute(name string) (string, error) {
	n := C.CString(name)
	rawval := C.GetAttribute(cl.Response, n)
	val := C.GoString(rawval)
	if C.ContainsError(rawval) == 0 {
		return "", fmt.Errorf("%s", val)
	}
	return val, nil
}
