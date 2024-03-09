package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	db "github.com/toml5566/go_pos_backend/internal/database"
	"github.com/toml5566/go_pos_backend/utils"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := utils.Config{
		TokenSecretKey:      utils.RandString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
