package settings_test

import (
	"net/http"
	"testing"

	"github.com/jrapoport/gothic/hosts/rest"
	"github.com/jrapoport/gothic/hosts/rest/admin/settings"
	"github.com/jrapoport/gothic/test/thttp"
	"github.com/jrapoport/gothic/test/tsrv"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testResponse(t *testing.T, s *rest.Host) string {
	test := s.API.Settings()
	b, err := json.Marshal(test)
	require.NoError(t, err)
	return string(b)
}

func TestSettingsServer_Settings(t *testing.T) {
	t.Parallel()
	srv, web, _ := tsrv.RESTHost(t, []rest.RegisterServer{
		settings.RegisterServer,
	}, false)
	j := srv.Config().JWT
	// not admin
	tok := thttp.UserToken(t, j, false, false)
	_, err := thttp.DoAuthRequest(t, web, http.MethodGet, settings.Endpoint, tok, nil, nil)
	assert.NoError(t, err)
	// admin
	tok = thttp.UserToken(t, j, false, true)
	res, err := thttp.DoAuthRequest(t, web, http.MethodGet, settings.Endpoint, tok, nil, nil)
	assert.NoError(t, err)
	assert.JSONEq(t, testResponse(t, srv), res)
}
