package winrm

import "github.com/masterzen/winrm/soap"

var soapXML string = "application/soap+xml"

type HttpPost func(*Client, *soap.SoapMessage) (string, error)

func Http_post(client *Client, request *soap.SoapMessage) (response string, err error) {
	transport := client.transport

	return transport.Post(client, request)
}
