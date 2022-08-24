package highestearningpool

import (
	"fmt"
	"net/http"
)

// verify interface compliance
var _ MyHTTPClientSrv = (*MockHTTPClientImpl)(nil)

type MockHTTPClientImpl struct {
	DoFn func(req *http.Request) (*http.Response, error)
}

func (c *MockHTTPClientImpl) Do(req *http.Request) (*http.Response, error) {
	if c != nil && c.DoFn != nil {
		return c.DoFn(req)
	}

	return nil, fmt.Errorf("test http client error")
}
