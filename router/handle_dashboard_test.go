package router

import (
	"testing"

	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestDashboard_Show(t *testing.T) {
	res := th.MastokGetRequestWithAuth(NewRouter(), "/dashboard")

	assert.Equal(t, res.Code, 200)
	assert.Contains(t, res.Body.String(), "campaigns")
	assert.Contains(t, res.Body.String(), "sessions")
}
