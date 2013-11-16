package controllers

import (
    // custom packages
    "github.com/cliohq/clio/core"
    "github.com/cliohq/clio/helpers/test"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    // current app resources
    "cartua/app/routes"

    // stdlib packages
    "testing"
    "net/http"
    "net/http/httptest"
)


var Server *httptest.Server
var Clio test.Response

func TestControllers (t *testing.T) {

    // setup
    Server = httptest.NewServer(http.HandlerFunc(core.Handler))
    Clio = test.NewResponse(Server)
    routes.Register()

    RegisterFailHandler (Fail)
    RunSpecs (t, "Controllers")

    // teardown
    Server.Close()
}
