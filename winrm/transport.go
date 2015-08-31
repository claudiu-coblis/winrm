package winrm

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/masterzen/winrm/soap"
)

var mutex = &sync.Mutex{}

type Transporter interface {
	Post(url, username, password string, request *soap.SoapMessage) (response string, err error)
}

// newTransport will create a new HTTP Transport, with options specified within the endpoint configuration
func newTransport(endpoint *Endpoint) (*Transport, error) {
	mutex.Lock()
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: endpoint.Insecure,
		},
	}

	if endpoint.CACert != nil && len(*endpoint.CACert) > 0 {
		certPool, err := readCACerts(endpoint.CACert)
		if err != nil {
			mutex.Unlock()
			return nil, err
		}

		transport.TLSClientConfig.RootCAs = certPool
	}

	if isSetCertAndPrivateKey(endpoint.Cert, endpoint.Key) {
		certPool, err := tls.X509KeyPair(*endpoint.Cert, *endpoint.Key)
		if err != nil {
			mutex.Unlock()
			return nil, fmt.Errorf("Error parsing keypair: %s", err)
		}

		transport.TLSClientConfig.Certificates = []tls.Certificate{certPool}
	}

	mutex.Unlock()

	return &Transport{transport}, nil
}

type Transport struct {
	Transport *http.Transport
}

func (s *Transport) Post(url, username, password string, request *soap.SoapMessage) (response string, err error) {
	httpClient := &http.Client{Transport: s.Transport}

	req, err := http.NewRequest("POST", url, strings.NewReader(request.String()))
	if err != nil {
		err = fmt.Errorf("impossible to create http request %s", err)
		return
	}

	req.Header.Set("Content-Type", soapXML+";charset=UTF-8")

	ok := false

	transport := *s.Transport
	if transport.TLSClientConfig.Certificates != nil {
		req.Header.Add("Authorization", "http://schemas.dmtf.org/wbem/wsman/1/wsman/secprofile/https/mutual")
		ok = true
	} else if username != "" && password != "" {
		req.SetBasicAuth(username, password)
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
