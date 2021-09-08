package runner

import (
	"bytes"
	"context"
	"io/ioutil"
	nethttp "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/slok/goresilience"
	"github.com/slok/goresilience/retry"
)

type http struct {
	base

	url         *url.URL
	method      string
	contentType string
	body        string
	retry       goresilience.Runner
}

func (r *http) Run() {
	r.Lock()
	defer r.Unlock()

	// Reset previous state
	r.err = nil

	//

	start := time.Now()
	logger := r.log.WithField("prefix", r.ctx.Name()).WithField("id", GenerateID())
	logger.Infof("%s %s", strings.ToUpper(r.method), r.url)
	if r.ctx.LogsFile() != nil {
		defer r.ctx.LogsFile().Sync()
	}

	err := r.retry.Run(context.Background(), func(ctx context.Context) error {
		request, err := nethttp.NewRequest(strings.ToUpper(r.method), r.url.String(), bytes.NewBufferString(r.body))
		if err != nil {
			return err
		}
		request.Close = true
		request.Header.Set("Content-Type", r.contentType)

		resp, err := nethttp.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		switch {
		case resp.StatusCode < 400:
			return nil
		case resp.StatusCode >= 400:
			body, err := ioutil.ReadAll(resp.Body)
			logger.WithField("code", resp.StatusCode).WithField("status", resp.Status).Error(string(body))
			return err
		}
		return nil
	})
	if err != nil {
		r.err = err
		logger.WithField("elapsed_time", time.Since(start)).WithField("ignored", r.ignoreError).Error(err)
		return
	}

	logger.WithField("elapsed_time", time.Since(start)).Info("finished")
}

func init() {
	supported := map[string]bool{
		nethttp.MethodGet:    true,
		nethttp.MethodHead:   true,
		nethttp.MethodPost:   true,
		nethttp.MethodPut:    true,
		nethttp.MethodPatch:  true,
		nethttp.MethodDelete: true,
	}

	Register("http", func(ctx Context, payload map[string]interface{}) (Runner, error) {
		if _, ok := payload["http"]; !ok {
			return nil, errors.New("taskfile: http: missing command http value")
		}

		rawurl, ok := payload["http"].(string)
		if !ok {
			return nil, errors.New("taskfile: http: http field must be a string")
		}

		var err error
		requester := &http{
			base: base{
				ctx: ctx,
			},
			method: nethttp.MethodGet,
		}

		requester.url, err = url.Parse(rawurl)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse URL")
		}

		//
		// Request
		if v, ok := payload["method"]; ok {
			method, ok := v.(string)
			if !ok {
				return nil, errors.New("taskfile: http: method must be a string")
			}

			if !supported[strings.ToUpper(method)] {
				return nil, errors.Errorf("taskfile: http: unsupported method '%s'", method)
			}

			requester.method = method
		}

		if v, ok := payload["content_type"]; ok {
			ct, ok := v.(string)
			if !ok {
				return nil, errors.New("taskfile: http: content_type must be a string")
			}

			requester.contentType = ct
		}

		if v, ok := payload["body"]; ok {
			body, ok := v.(string)
			if !ok {
				return nil, errors.New("taskfile: http: body must be a string")
			}

			requester.body = body
		}

		//
		// Retry
		config := retry.Config{}
		if v, ok := payload["retry"]; ok {
			times, ok := v.(int)
			if !ok {
				return nil, errors.New("taskfile: http: retry must be an integer")
			}
			config.Times = int(times)
		}
		if v, ok := payload["retry_interval"]; ok {
			duration, ok := v.(string)
			if !ok {
				return nil, errors.New("taskfile: http: retry_interval must be a string")
			}

			config.WaitBase, err = time.ParseDuration(duration)
			if err != nil {
				return nil, errors.Wrap(err, "taskfile: http: retry_interval")
			}
		}
		requester.retry = retry.New(config)

		//
		// IgnoreError
		if v, ok := payload["ignore_error"]; ok {
			b, ok := v.(bool)
			if !ok {
				return nil, errors.New("taskfile: http: ignore_error field must be a boolean")
			}

			requester.ignoreError = b
		}

		return requester, nil
	})
}
