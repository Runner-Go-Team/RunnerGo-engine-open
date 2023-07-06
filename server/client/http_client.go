package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/valyala/fasthttp"
	"strings"
	"sync"
	"time"
)

var (
	KeepAliveClient *fasthttp.Client
	once            sync.Once
)

func HTTPRequest(method, url string, body *model.Body, query *model.Query, header *model.Header, cookie *model.Cookie, auth *model.Auth, httpApiSetup *model.HttpApiSetup) (resp *fasthttp.Response, req *fasthttp.Request, requestTime uint64, sendBytes float64, err error, str string, startTime, endTime time.Time) {
	var client *fasthttp.Client
	req = fasthttp.AcquireRequest()
	if httpApiSetup.KeepAlive {
		newKeepAlive(httpApiSetup, auth)
		client = KeepAliveClient
		req.Header.Set("Connection", "keep-alive")
	} else {
		client = fastClient(httpApiSetup, auth)
	}

	// set method
	req.Header.SetMethod(method)
	// set header
	header.SetHeader(req)
	cookie.SetCookie(req)
	urls := strings.Split(url, "//")
	if !strings.EqualFold(urls[0], constant.HTTP) && !strings.EqualFold(urls[0], constant.HTTPS) {
		url = constant.HTTP + "//" + url

	}

	urlQuery := req.URI().QueryArgs()

	if query.Parameter != nil {
		for _, v := range query.Parameter {
			if v.IsChecked != constant.Open {
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
	req.SetTimeout(30 * time.Second)
	// 发送请求
	if httpApiSetup.IsRedirects == 0 {
		err = client.DoRedirects(req, resp, httpApiSetup.RedirectsNum)
	} else {
		err = client.Do(req, resp)
	}
	//err = client.Do(req, resp)
	endTime = time.Now()
	requestTime = uint64(time.Since(startTime))
	sendBytes = float64(req.Header.ContentLength()) / 1024
	if sendBytes <= 0 {
		sendBytes = float64(len(req.Body())) / 1024
	}
	return
}

// 获取fasthttp客户端
func fastClient(httpApiSetup *model.HttpApiSetup, auth *model.Auth) (fc *fasthttp.Client) {
	tr := &tls.Config{InsecureSkipVerify: true}
	if auth != nil || auth.Bidirectional != nil {
		switch auth.Type {
		case constant.Bidirectional:
			tr.InsecureSkipVerify = false
			if auth.Bidirectional.CaCert != "" {
				if strings.HasPrefix(auth.Bidirectional.CaCert, "https://") || strings.HasPrefix(auth.Bidirectional.CaCert, "http://") {
					client := &fasthttp.Client{}
					loadReq := fasthttp.AcquireRequest()
					defer loadReq.ConnectionClose()
					// set url
					loadReq.Header.SetMethod("GET")
					loadReq.SetRequestURI(auth.Bidirectional.CaCert)
					loadResp := fasthttp.AcquireResponse()
					defer loadResp.ConnectionClose()
					if err := client.Do(loadReq, loadResp); err != nil {
						log.Logger.Error(fmt.Sprintf("机器ip:%s, 下载crt文件失败：", middlewares.LocalIp), err)
					}
					if loadResp != nil && loadResp.Body() != nil {
						caCertPool := x509.NewCertPool()
						if caCertPool != nil {
							caCertPool.AppendCertsFromPEM(loadResp.Body())
							tr.ClientCAs = caCertPool
						}
					}
				}
			}
		case constant.Unidirectional:
			tr.InsecureSkipVerify = false
		}
	}
	fc = &fasthttp.Client{
		TLSConfig: tr,
	}
	if httpApiSetup.ClientName != "" {
		fc.Name = httpApiSetup.ClientName
	}
	if httpApiSetup.UserAgent {
		fc.NoDefaultUserAgentHeader = false
	}
	if httpApiSetup.MaxIdleConnDuration != 0 {
		fc.MaxIdleConnDuration = time.Duration(httpApiSetup.MaxIdleConnDuration) * time.Second
	} else {
		fc.MaxIdleConnDuration = time.Duration(0) * time.Second
	}
	if httpApiSetup.MaxConnPerHost != 0 {
		fc.MaxConnsPerHost = httpApiSetup.MaxConnPerHost
	}

	fc.MaxConnWaitTimeout = time.Duration(httpApiSetup.MaxConnWaitTimeout) * time.Second
	fc.WriteTimeout = time.Duration(httpApiSetup.WriteTimeOut) * time.Millisecond
	fc.ReadTimeout = time.Duration(httpApiSetup.ReadTimeOut) * time.Millisecond

	return fc
}

func newKeepAlive(httpApiSetup *model.HttpApiSetup, auth *model.Auth) {
	once.Do(func() {
		tr := &tls.Config{InsecureSkipVerify: true}
		if auth != nil && auth.Bidirectional != nil {
			switch auth.Type {
			case constant.Bidirectional:
				tr.InsecureSkipVerify = false
				if auth.Bidirectional.CaCert != "" {
					if strings.HasPrefix(auth.Bidirectional.CaCert, "https://") || strings.HasPrefix(auth.Bidirectional.CaCert, "http://") {
						client := &fasthttp.Client{}
						loadReq := fasthttp.AcquireRequest()
						defer loadReq.ConnectionClose()
						// set url
						loadReq.Header.SetMethod("GET")
						loadReq.SetRequestURI(auth.Bidirectional.CaCert)
						loadResp := fasthttp.AcquireResponse()
						defer loadResp.ConnectionClose()
						if err := client.Do(loadReq, loadResp); err != nil {
							log.Logger.Error(fmt.Sprintf("机器ip:%s, 下载crt文件失败：", middlewares.LocalIp), err)
						}
						if loadResp != nil && loadResp.Body() != nil {
							caCertPool := x509.NewCertPool()
							if caCertPool != nil {
								caCertPool.AppendCertsFromPEM(loadResp.Body())
								tr.ClientCAs = caCertPool
							}
						}
					}
				}
			case constant.Unidirectional:
				tr.InsecureSkipVerify = false
			}
		}
		KeepAliveClient = &fasthttp.Client{
			TLSConfig: tr,
		}
		if httpApiSetup.ClientName != "" {
			KeepAliveClient.Name = httpApiSetup.ClientName
		}
		if httpApiSetup.UserAgent {
			KeepAliveClient.NoDefaultUserAgentHeader = false
		}
		KeepAliveClient.MaxIdleConnDuration = time.Duration(httpApiSetup.MaxIdleConnDuration) * time.Second
		if httpApiSetup.MaxConnPerHost != 0 {
			KeepAliveClient.MaxConnsPerHost = httpApiSetup.MaxConnPerHost
		}

		KeepAliveClient.WriteTimeout = time.Duration(httpApiSetup.WriteTimeOut) * time.Millisecond
		KeepAliveClient.ReadTimeout = time.Duration(httpApiSetup.ReadTimeOut) * time.Millisecond
		KeepAliveClient.MaxConnWaitTimeout = time.Duration(httpApiSetup.MaxConnWaitTimeout) * time.Second
	})
}
