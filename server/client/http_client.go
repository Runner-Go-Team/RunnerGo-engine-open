package client

import (
	"RunnerGo-engine/config"
	"RunnerGo-engine/model"
	"crypto/tls"
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
	"time"
)

func HTTPRequest(method, url string, body *model.Body, query *model.Query, header *model.Header, auth *model.Auth, httpApiSetup *model.HttpApiSetup) (resp *fasthttp.Response, req *fasthttp.Request, requestTime uint64, sendBytes float64, err error, str string, startTime, endTime time.Time) {

	client := fastClient(httpApiSetup.ReadTimeOut, httpApiSetup.WriteTimeOut)
	req = fasthttp.AcquireRequest()

	// set methon
	req.Header.SetMethod(method)

	// set header
	header.SetHeader(req)

	urls := strings.Split(url, "//")
	if !strings.EqualFold(urls[0], model.HTTP) && !strings.EqualFold(urls[0], model.HTTPS) {
		url = model.HTTP + "//" + url

	}

	urlQuery := req.URI().QueryArgs()

	if query.Parameter != nil {
		for _, v := range query.Parameter {
			if v.IsChecked != model.Open {
				continue
			}
			if !strings.Contains(url, v.Key) {
				by := v.ValueToByte()
				urlQuery.AddBytesV(v.Key, by)
				url = url + fmt.Sprintf("&%s=%s", v.Key, string(v.ValueToByte()))
			}
		}
	}
	// set url
	req.SetRequestURI(url)
	// set body
	str = body.SetBody(req)

	// set auth
	auth.SetAuth(req)
	resp = fasthttp.AcquireResponse()

	startTime = time.Now()
	// 发送请求
	if httpApiSetup.IsRedirects == 0 {
		maxRedirectsCount := 3
		if httpApiSetup.RedirectsNum != maxRedirectsCount {
			maxRedirectsCount = httpApiSetup.RedirectsNum
		}
		err = client.DoRedirects(req, resp, maxRedirectsCount)
	} else {
		err = client.Do(req, resp)
	}
	endTime = time.Now()
	requestTime = uint64(time.Since(startTime))
	sendBytes = float64(req.Header.ContentLength()) / 1024
	if sendBytes <= 0 {
		sendBytes = float64(len(req.Body())) / 1024
	}
	return
}

// 获取fasthttp客户端
func fastClient(readTimeOut, writeTimeOut int64) (fc *fasthttp.Client) {
	fc = &fasthttp.Client{
		Name:                     config.Conf.Http.Name,
		NoDefaultUserAgentHeader: config.Conf.Http.NoDefaultUserAgentHeader,
		TLSConfig:                &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:          config.Conf.Http.MaxConnPerHost,
		MaxIdleConnDuration:      config.Conf.Http.MaxIdleConnDuration * time.Millisecond,
		MaxConnWaitTimeout:       config.Conf.Http.MaxConnWaitTimeout * time.Millisecond,
	}
	if writeTimeOut != 0 {
		fc.WriteTimeout = time.Duration(writeTimeOut) * time.Millisecond
	}

	if readTimeOut != 0 {
		fc.ReadTimeout = time.Duration(readTimeOut) * time.Millisecond
	}

	return fc
}
