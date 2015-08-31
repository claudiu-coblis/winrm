package winrm

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/masterzen/winrm/soap"
)

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

type Transporter interface {
	Post(client *Client, request *soap.SoapMessage) (string, error)
}

func NewTransport(endpoint *Endpoint) (*Transport, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: endpoint.Insecure,
		},
	}

	if endpoint.CACert != nil && len(*endpoint.CACert) > 0 {
		certPool, err := readCACerts(endpoint.CACert)
		if err != nil {
			return nil, err
		}

		transport.TLSClientConfig.RootCAs = certPool
	}

	if isSetCertAndPrivateKey(endpoint.Cert, endpoint.Key) {
		certPool, err := tls.X509KeyPair(*endpoint.Cert, *endpoint.Key)
		if err != nil {
			return nil, fmt.Errorf("Error parsing keypair: %s", err)
		}

		transport.TLSClientConfig.Certificates = []tls.Certificate{certPool}
	}

	return &Transport{transport}, nil
}

type Transport struct {
	transport *http.Transport
}

func (s *Transport) Post(client *Client, request *soap.SoapMessage) (response string, err error) {
	httpClient := &http.Client{Transport: s.transport}

	req, err := http.NewRequest("POST", client.url, strings.NewReader(request.String()))
	if err != nil {
		err = fmt.Errorf("impossible to create http request %s", err)
		return
	}

	req.Header.Set("Content-Type", soapXML+";charset=UTF-8")

	ok := false

	transport := *s.transport
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
