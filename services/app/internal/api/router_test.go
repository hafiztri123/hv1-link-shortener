package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerAndRegisterRoutes(t *testing.T) {
	server := NewServer(nil, nil, nil, nil, nil)
	router := server.RegisterRoutes()

	assert.NotNil(t, server, "New server should not be nil")
	assert.NotNil(t, router, "Register routes should not be nil")

}
