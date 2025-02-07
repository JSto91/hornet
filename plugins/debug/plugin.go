package debug

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/gohornet/hornet/pkg/model/storage"
	"github.com/gohornet/hornet/pkg/model/syncmanager"
	"github.com/gohornet/hornet/pkg/model/utxo"
	"github.com/gohornet/hornet/pkg/node"
	"github.com/gohornet/hornet/pkg/protocol/gossip"
	restapipkg "github.com/gohornet/hornet/pkg/restapi"
	"github.com/gohornet/hornet/pkg/tangle"
	"github.com/gohornet/hornet/plugins/restapi"
	"github.com/iotaledger/hive.go/configuration"
)

const (
	// RouteDebugSolidifier is the debug route to manually trigger the solidifier.
	// POST triggers the solidifier.
	RouteDebugSolidifier = "/solidifier"

	// RouteDebugOutputs is the debug route for getting all output IDs.
	// GET returns the outputIDs for all outputs.
	RouteDebugOutputs = "/outputs"

	// RouteDebugOutputsUnspent is the debug route for getting all unspent output IDs.
	// GET returns the outputIDs for all unspent outputs.
	RouteDebugOutputsUnspent = "/outputs/unspent"

	// RouteDebugOutputsSpent is the debug route for getting all spent output IDs.
	// GET returns the outputIDs for all spent outputs.
	RouteDebugOutputsSpent = "/outputs/spent"

	// RouteDebugMilestoneDiffs is the debug route for getting a milestone diff by it's milestoneIndex.
	// GET returns the utxo diff (new outputs & spents) for the milestone index.
	RouteDebugMilestoneDiffs = "/ms-diff/:" + restapipkg.ParameterMilestoneIndex

	// RouteDebugRequests is the debug route for getting all pending requests.
	// GET returns a list of all pending requests.
	RouteDebugRequests = "/requests"

	// RouteDebugMessageCone is the debug route for traversing a cone of a message.
	// it traverses the parents of a message until they reference an older milestone than the start message.
	// GET returns the path of this traversal and the "entry points".
	RouteDebugMessageCone = "/message-cones/:" + restapipkg.ParameterMessageID
)

func init() {
	Plugin = &node.Plugin{
		Status: node.StatusDisabled,
		Pluggable: node.Pluggable{
			Name:      "Debug",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Configure: configure,
		},
	}
}

var (
	Plugin *node.Plugin
	deps   dependencies
)

type dependencies struct {
	dig.In
	Storage           *storage.Storage
	SyncManager       *syncmanager.SyncManager
	Tangle            *tangle.Tangle
	RequestQueue      gossip.RequestQueue
	UTXOManager       *utxo.Manager
	NodeConfig        *configuration.Configuration `name:"nodeConfig"`
	NetworkID         uint64                       `name:"networkId"`
	RestPluginManager *restapi.RestPluginManager   `optional:"true"`
}

func configure() {
	// check if RestAPI plugin is disabled
	if Plugin.Node.IsSkipped(restapi.Plugin) {
		Plugin.LogPanic("RestAPI plugin needs to be enabled to use the Debug plugin")
	}

	routeGroup := deps.RestPluginManager.AddPlugin("debug/v1")

	routeGroup.POST(RouteDebugSolidifier, func(c echo.Context) error {
		deps.Tangle.TriggerSolidifier()

		return c.NoContent(http.StatusNoContent)
	})

	routeGroup.GET(RouteDebugOutputs, func(c echo.Context) error {
		resp, err := outputsIDs(c)
		if err != nil {
			return err
		}

		return restapipkg.JSONResponse(c, http.StatusOK, resp)
	})

	routeGroup.GET(RouteDebugOutputsUnspent, func(c echo.Context) error {
		resp, err := unspentOutputsIDs(c)
		if err != nil {
			return err
		}

		return restapipkg.JSONResponse(c, http.StatusOK, resp)
	})

	routeGroup.GET(RouteDebugOutputsSpent, func(c echo.Context) error {
		resp, err := spentOutputsIDs(c)
		if err != nil {
			return err
		}

		return restapipkg.JSONResponse(c, http.StatusOK, resp)
	})

	routeGroup.GET(RouteDebugMilestoneDiffs, func(c echo.Context) error {
		resp, err := milestoneDiff(c)
		if err != nil {
			return err
		}

		return restapipkg.JSONResponse(c, http.StatusOK, resp)
	})

	routeGroup.GET(RouteDebugRequests, func(c echo.Context) error {
		resp, err := requests(c)
		if err != nil {
			return err
		}

		return restapipkg.JSONResponse(c, http.StatusOK, resp)
	})

	routeGroup.GET(RouteDebugMessageCone, func(c echo.Context) error {
		resp, err := messageCone(c)
		if err != nil {
			return err
		}

		return restapipkg.JSONResponse(c, http.StatusOK, resp)
	})
}
