package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"time"

	"github.com/go-gorp/gorp"
	"github.com/gorilla/mux"
	"github.com/ovh/cds/engine/api/bootstrap"
	"github.com/ovh/cds/engine/api/test"
)

func newTestAPIWithIzanamiToken(t *testing.T, token string, bootstrapFunc ...test.Bootstrapf) (*API, *gorp.DbMap, *Router) {
	bootstrapFunc = append(bootstrapFunc, bootstrap.InitiliazeDB)
	db, cache := test.SetupPG(t, bootstrapFunc...)
	router := newRouter(mux.NewRouter(), "/"+test.GetTestName(t))
	var cancel context.CancelFunc
	router.Background, cancel = context.WithCancel(context.Background())
	api := &API{
		StartupTime:         time.Now(),
		Router:              router,
		DBConnectionFactory: test.DBConnectionFactory,
		Config:              Configuration{},
		Cache:               cache,
	}
	api.Config.Features.Izanami.Token = token
	api.InitRouter()
	t.Cleanup(func() {
		cancel()
	})
	return api, db, router
}

func TestFeatureClean(t *testing.T) {
	api, _, router := newTestAPIWithIzanamiToken(t, "mytoken", bootstrap.InitiliazeDB)

	vars := map[string]string{}
	uri := router.GetRoute("POST", api.cleanFeatureHandler, vars)
	req, err := http.NewRequest("POST", uri, nil)
	test.NoError(t, err)
	req.Header.Set("X-Izanami-Token", "666")

	w := httptest.NewRecorder()
	router.Mux.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	w = httptest.NewRecorder()
	req.Header.Set("X-Izanami-Token", "mytoken")
	router.Mux.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)
}
