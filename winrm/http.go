package winrm

import (
	"fmt"
	"io/ioutil"
	"strings"

	ghttp "launchpad.net/gwacl/fork/http"
	"net/http"

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

func gbody(response *ghttp.Response) (content string, err error) {
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
	if client.transport != nil {
		return Http_Post(client, request)
	} else if client.gtransport != nil {
		return GHttp_Post(client, request)
	}
	return "", fmt.Errorf("Transport error!")
}

func Http_Post(client *Client, request *soap.SoapMessage) (response string, err error) {
	httpClient := &http.Client{Transport: client.transport}

	req, err := http.NewRequest("POST", client.url, strings.NewReader(request.String()))
	if err != nil {
		err = fmt.Errorf("impossible to create http request %s", err)
		return
	}
	req.Header.Set("Content-Type", soapXML+";charset=UTF-8")

	ok := false

	transport := client.transport
	if transport.TLSClientConfig.Certificates != nil {
		req.Header.Add("Authorization", "http://schemas.dmtf.org/wbem/wsman/1/wsman/secprofile/https/mutual")
		ok = true
	} else if client.username != "" && client.password != "" {
		req.SetBasicAuth(client.username, client.password)
		ok = true
	}

	if ok == false {
		err = fmt.Errorf("Invalid transport type")
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("unknown error %s", err)
		return
	}

	if resp.StatusCode == 200 {
		response, err = body(resp)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("http error: %d - %s", resp.StatusCode, body)
	}

	return
}

func GHttp_Post(client *Client, request *soap.SoapMessage) (response string, err error) {
	httpClient := &ghttp.Client{Transport: client.gtransport}

	req, err := ghttp.NewRequest("POST", client.url, strings.NewReader(request.String()))
	if err != nil {
		err = fmt.Errorf("impossible to create http request %s", err)
		return
	}
	req.Header.Set("Content-Type", soapXML+";charset=UTF-8")

	ok := false

	transport := client.gtransport
	if transport.TLSClientConfig.Certificates != nil {
		req.Header.Add("Authorization", "http://schemas.dmtf.org/wbem/wsman/1/wsman/secprofile/https/mutual")
		ok = true
	} else if client.username != "" && client.password != "" {
		req.SetBasicAuth(client.username, client.password)
		ok = true
	}

	if ok == false {
		err = fmt.Errorf("Invalid transport type")
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("unknown error %s", err)
		return
	}

	if resp.StatusCode == 200 {
		response, err = gbody(resp)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("http error: %d - %s", resp.StatusCode, body)
	}

	return
}
