package winrm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/masterzen/winrm/soap"
)

var soapXML string = "application/soap+xml"

type HttpPost func(*Client, *soap.SoapMessage) (string, error)

func body(response *http.Response) (content string, err error) {
	contentType := response.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, soapXML) {
		var body []byte
		body, err = ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			err = fmt.Errorf("error while reading request body %s", err)
			return
		}

		content = string(body)
		return
	} else {
		err = fmt.Errorf("invalid content-type: %s", contentType)
		return
	}
	return
}

func Http_post(client *Client, request *soap.SoapMessage) (response string, err error) {
	transport := client.transport

	return transport.Post(client.url, client.username, client.password, request)
}
