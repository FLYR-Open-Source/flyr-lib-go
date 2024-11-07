package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetHttpTransport(t *testing.T) {
	client := http.Client{}
	client = SetHttpTransport(client)

	require.NotNil(t, client.Transport)
}

func TestNewHttpClient(t *testing.T) {
	client := NewHttpClient()
	require.NotNil(t, client.Transport)
}
