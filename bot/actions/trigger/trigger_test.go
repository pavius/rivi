package trigger

import (
	"net/http"
	"testing"

	"github.com/bivas/rivi/bot/mock"
	"github.com/bivas/rivi/util/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"
)

func TestSerialization(t *testing.T) {
	input := map[string]interface{}{
		"endpoint": "http://example.com",
		"method":   "GET",
	}

	var f factory
	result := f.BuildAction(input)
	assert.NotNil(t, result, "should create action")
	s, ok := result.(*action)
	assert.True(t, ok, "should be of this package")
	assert.Equal(t, "GET", s.rule.Method, "method")
	assert.Equal(t, "http://example.com", s.rule.Endpoint, "endpoint")
}

func TestTriggerDefaults(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	meta := &mock.MockEventData{
		Number: 1,
		Title:  "title1",
		State:  "tested",
		Owner:  "test",
		Repo:   "repo1",
		Origin: "tester",
	}
	rule := &rule{
		Endpoint: "http://example.com/trigger",
	}
	rule.Defaults()
	httpmock.RegisterResponder(
		"POST",
		"http://example.com/trigger",
		func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "trigger", req.Header.Get("X-RiviBot-Event"), "missing correct event")
			assert.Equal(t, "RiviBot-Agent/1.0", req.UserAgent(), "user agent")
			return httpmock.NewStringResponse(200, ""), nil
		})
	action := &action{rule: rule, client: http.DefaultClient, logger: log.Get("trigger.test")}
	action.Apply(&mock.MockConfiguration{}, meta)
	assert.Nil(t, action.err, "error when sending trigger")
}

func TestTriggerGet(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	meta := &mock.MockEventData{
		Number: 1,
		Title:  "title1",
		State:  "tested",
		Owner:  "test",
		Repo:   "repo1",
		Origin: "tester",
	}
	rule := &rule{
		Endpoint: "http://example.com/trigger",
		Method:   "GET",
	}
	rule.Defaults()
	httpmock.RegisterResponder(
		"GET",
		"http://example.com/trigger",
		func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "trigger", req.Header.Get("X-RiviBot-Event"), "missing correct event")
			assert.Equal(t, "RiviBot-Agent/1.0", req.UserAgent(), "user agent")
			return httpmock.NewStringResponse(200, ""), nil
		})
	action := &action{rule: rule, client: http.DefaultClient, logger: log.Get("trigger.test")}
	action.Apply(&mock.MockConfiguration{}, meta)
	assert.Nil(t, action.err, "error when sending trigger")
}

func TestTriggerHeaders(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	meta := &mock.MockEventData{
		Number: 1,
		Title:  "title1",
		State:  "tested",
		Owner:  "test",
		Repo:   "repo1",
		Origin: "tester",
	}
	headers := make(map[string]string)
	headers["not-allowed"] = "fail"
	headers["x-allowed"] = "allowed"
	headers["x-rivibot-fake"] = "fail"
	rule := &rule{
		Endpoint: "http://example.com/trigger",
		Headers:  headers,
	}
	rule.Defaults()
	httpmock.RegisterResponder(
		"POST",
		"http://example.com/trigger",
		func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "trigger", req.Header.Get("X-RiviBot-Event"), "missing correct event")
			assert.Equal(t, "RiviBot-Agent/1.0", req.UserAgent(), "user agent")
			assert.Equal(t, "allowed", req.Header.Get("x-allowed"), "user added header")
			assert.Empty(t, req.Header.Get("not-allowed"), "not allowed header")
			assert.Empty(t, req.Header.Get("x-rivibot-fake"), "not allowed x-rivi header")
			return httpmock.NewStringResponse(200, ""), nil
		})
	action := &action{rule: rule, client: http.DefaultClient, logger: log.Get("trigger.test")}
	action.Apply(&mock.MockConfiguration{}, meta)
	assert.Nil(t, action.err, "error when sending trigger")
}
