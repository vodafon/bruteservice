package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type Analysis struct {
	Status                 int       `json:"status"`
	RequestURL             string    `json:"request_url"`
	RequestMethod          string    `json:"request_method"`
	ResponseHeaderKey      string    `json:"response_header_key"`
	ResponseHeaderValue    string    `json:"response_header_value"`
	Response               string    `json:"response"`
	ResponseHeaderKeyValue [2]string `json:"response_header_key_value"`
	RequestHeaderKey       string    `json:"request_header_key"`
	RequestHeaderValue     string    `json:"request_header_value"`
	RequestHeaderKeyValue  [2]string `json:"request_header_key_value"`

	requestURLReg             *regexp.Regexp
	responseHeaderKeyReg      *regexp.Regexp
	responseHeaderValueReg    *regexp.Regexp
	responseReg               *regexp.Regexp
	responseHeaderKeyValueReg [2]*regexp.Regexp
	requestHeaderKeyReg       *regexp.Regexp
	requestHeaderValueReg     *regexp.Regexp
	requestHeaderKeyValueReg  [2]*regexp.Regexp
}

func (an *Analysis) Compile() {
	an.requestURLReg = compileIfPresent(an.RequestURL)
	an.responseHeaderKeyReg = compileIfPresent(an.ResponseHeaderKey)
	an.responseHeaderValueReg = compileIfPresent(an.ResponseHeaderValue)
	an.responseReg = compileIfPresent(an.Response)
	an.responseHeaderKeyValueReg[0] = compileIfPresent(an.ResponseHeaderKeyValue[0])
	an.responseHeaderKeyValueReg[1] = compileIfPresent(an.ResponseHeaderKeyValue[1])
	an.requestHeaderKeyReg = compileIfPresent(an.RequestHeaderKey)
	an.requestHeaderValueReg = compileIfPresent(an.RequestHeaderValue)
	an.requestHeaderKeyValueReg[0] = compileIfPresent(an.RequestHeaderKeyValue[0])
	an.requestHeaderKeyValueReg[1] = compileIfPresent(an.RequestHeaderKeyValue[1])
}

func compileIfPresent(str string) *regexp.Regexp {
	if str == "" {
		return nil
	}

	return regexp.MustCompile(str)
}

func (an *Analysis) Analyze(resp *http.Response) (bool, error) {
	var err error
	match := an.analyzeResponseStatus(resp)
	if !match {
		return false, nil
	}
	match, err = an.analyzeResponse(resp)
	if err != nil || !match {
		return false, err
	}
	match = an.analyzeResponseHeaders(resp)
	if !match {
		return false, nil
	}
	return true, nil
}

func (an Analysis) analyzeResponseStatus(resp *http.Response) bool {
	if an.Status == 0 {
		return true
	}
	return resp.StatusCode == an.Status
}

func (an Analysis) analyzeResponseHeaders(resp *http.Response) bool {
	if an.responseHeaderKeyReg == nil && an.responseHeaderValueReg == nil &&
		an.responseHeaderKeyValueReg[0] == nil && an.responseHeaderKeyValueReg[1] == nil {
		return true
	}
	matches := make(map[string]bool)
	for k, vs := range resp.Header {
		v := strings.Join(vs, ", ")
		matches["key1"] = matches["key1"] || nilOrMatchHeader(an.responseHeaderKeyReg, k, k+":"+v)
		matches["val1"] = matches["val1"] || nilOrMatchHeader(an.responseHeaderValueReg, v, k+":"+v)
		matches["key_val"] = matches["key_val"] ||
			(nilOrMatchHeader(an.responseHeaderKeyValueReg[0], k, k+":"+v) && nilOrMatchHeader(an.responseHeaderKeyValueReg[1], v, k+":"+v))
	}
	for _, v := range matches {
		if !v {
			return false
		}
	}
	return true
}

func nilOrMatchHeader(re *regexp.Regexp, val, header string) bool {
	if re == nil {
		return true
	}
	rematch := re.FindString(strings.ToLower(val))
	return len(rematch) != 0
}

func (an Analysis) analyzeResponse(resp *http.Response) (bool, error) {
	if an.responseReg == nil {
		return true, nil
	}
	bodyB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	bodyL := bytes.ToLower(bodyB)
	resp.Body = ioutil.NopCloser(bytes.NewReader(bodyB))

	rematch := an.responseReg.Find(bodyL)
	if len(rematch) == 0 {
		return false, nil
	}
	return true, nil
}
