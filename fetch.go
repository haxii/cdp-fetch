package cdp_fetch

import (
	"encoding/json"
	"github.com/go-rod/rod"
	"net/http"
	"strings"
)

type Fetch struct {
	page *rod.Page
}

func NewFetch(page *rod.Page) *Fetch {
	return &Fetch{page: page}
}

// Request request def for fetch https://developer.mozilla.org/en-US/docs/Web/API/Request/Request
type Request struct {
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body,omitempty"`
	// Mode The mode you want to use for the request, e.g., cors, no-cors, same-origin, or navigate. The default is cors.
	Mode string `json:"mode,omitempty"`
	// Credentials The request credentials you want to use for the request: omit, same-origin, or include. The default is same-origin.
	Credentials string `json:"credentials,omitempty"`
	// Cache The cache mode you want to use for the request.
	Cache string `json:"cache,omitempty"`
	// Redirect The redirect mode to use: follow, error, or manual. The default is follow.
	Redirect string `json:"redirect,omitempty"`
	// Referrer A string specifying no-referrer, client, or a URL. The default is about:client.
	Referrer string `json:"referrer,omitempty"`
	// Integrity Contains the subresource integrity value of the request (e.g., sha256-BpfBw7ivV8q2jLiT13fxDYAe2tJllusRSZ273h2nFSE=).
	Integrity string `json:"integrity,omitempty"`
}

func (req *Request) Marshall() ([]byte, error) {
	return json.Marshal(req)
}

// Response response def for fetch https://developer.mozilla.org/en-US/docs/Web/API/Response
type Response struct {
	// Type The type of the response (e.g., basic, cors).
	Type    string      `json:"type"`
	URL     string      `json:"url"`
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
	Body    string      `json:"body"`
}

func (api *Fetch) Fetch(url string, req Request) (*Response, error) {
	jsSrc, err := makeFetchJS(url, req)
	if err != nil {
		return nil, err
	}
	evalResult, evalErr := api.page.Evaluate(rod.Eval(jsSrc).ByPromise())
	if evalErr != nil {
		return nil, evalErr
	}
	resp := &Response{}
	err = evalResult.Value.Unmarshal(resp)
	return resp, nil
}

func makeFetchJS(url string, req Request) (string, error) {
	reqBytes, err := req.Marshall()
	if err != nil {
		return "", err
	}
	src := strings.Builder{}
	src.WriteString(`async _ =>{`) // async func begin
	{
		// make request
		src.WriteString(`const resp = await fetch("`)
		src.WriteString(url)
		src.WriteString(`",`)
		src.Write(reqBytes)
		src.WriteString(");")
		// parse response and return
		src.WriteString("headers={}; for (let e of resp.headers.entries()){headers[e[0]]=e.slice(1)};")
		src.WriteString("const body = await resp.text();")
		src.WriteString("return {type:resp.type,url:resp.url,status:resp.status,headers,body};")
	}
	src.WriteString(`}`) // async func ends
	return src.String(), nil
}
