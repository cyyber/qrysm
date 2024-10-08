package apimiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/api/gateway/apimiddleware"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
)

func TestSetVoluntaryExitEpoch(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		endpoint := &apimiddleware.Endpoint{
			PostRequest: &SetVoluntaryExitRequestJson{},
		}
		epoch := "300"

		var body bytes.Buffer
		request := httptest.NewRequest("POST", fmt.Sprintf("http://foo.example?epoch=%s", epoch), &body)

		runDefault, errJson := setVoluntaryExitEpoch(endpoint, nil, request)
		require.Equal(t, true, errJson == nil)
		assert.Equal(t, apimiddleware.RunDefault(true), runDefault)

		var b SetVoluntaryExitRequestJson
		err := json.NewDecoder(request.Body).Decode(&b)
		require.NoError(t, err)
		require.Equal(t, epoch, b.Epoch)
	})
	t.Run("invalid query returns error", func(t *testing.T) {
		endpoint := &apimiddleware.Endpoint{
			PostRequest: &SetVoluntaryExitRequestJson{},
		}
		epoch := "/12"
		var body bytes.Buffer
		request := httptest.NewRequest("POST", fmt.Sprintf("http://foo.example?epoch=%s", epoch), &body)

		runDefault, errJson := setVoluntaryExitEpoch(endpoint, nil, request)
		assert.NotNil(t, errJson)
		assert.Equal(t, apimiddleware.RunDefault(false), runDefault)
		err := errors.New(errJson.Msg())
		assert.ErrorContains(t, "invalid epoch", err)
	})
}
