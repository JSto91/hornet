package framework

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/gohornet/hornet/core/app"
	"github.com/gohornet/hornet/core/gossip"
	"github.com/gohornet/hornet/core/p2p"
	"github.com/gohornet/hornet/core/protocfg"
	"github.com/gohornet/hornet/core/snapshot"
	"github.com/gohornet/hornet/pkg/utils"
	"github.com/gohornet/hornet/plugins/autopeering"
	"github.com/gohornet/hornet/plugins/dashboard"
	"github.com/gohornet/hornet/plugins/inx"
	"github.com/gohornet/hornet/plugins/profiling"
	"github.com/gohornet/hornet/plugins/receipt"
	"github.com/gohornet/hornet/plugins/restapi"
	"github.com/gohornet/inx-coordinator/core/coordinator"
	"github.com/gohornet/inx-coordinator/plugin/migrator"
	iotago "github.com/iotaledger/iota.go/v3"
)

const (
	// The default REST API port of every node.
	RestAPIPort = 14265

	GenesisAddressPublicKeyHex = "f7868ab6bb55800b77b8b74191ad8285a9bf428ace579d541fda47661803ff44"
	GenesisAddressHex          = "6920b176f613ec7be59e68fc68f597eb3393af80f74c7c3db78198147d5f1f92"
	GenesisAddressBech32       = "atoi1qp5jpvtk7cf7c7l9ne50c684jl4n8ya0srm5clpak7qes9ratu0eysflmsz"

	autopeeringMaxTries = 50

	containerNodeImage           = "hornet:dev"
	coordinatorImage             = "iotaledger/inx-coordinator:latest"
	indexerImage                 = "iotaledger/inx-indexer:latest"
	containerWhiteFlagMockServer = "wfmock:latest"

	containerNameTester  = "/tester"
	containerNameReplica = "replica_"
	containerNameINX     = "inx_"

	logsDir = "/tmp/logs/"

	assetsDir = "/assets"

	dockerLogsPrefixLen    = 8
	exitStatusSuccessful   = 0
	containerNameEntryNode = "entry_node"
)

var (
	disabledPluginsPeer      = []string{}
	disabledPluginsEntryNode = []string{}
	// The seed on which the total supply resides on per default.
	GenesisSeed    ed25519.PrivateKey
	GenesisAddress iotago.Ed25519Address
)

func init() {
	prvkey, err := utils.ParseEd25519PrivateKeyFromString("256a818b2aac458941f7274985a410e57fb750f3a3a67969ece5bd9ae7eef5b2f7868ab6bb55800b77b8b74191ad8285a9bf428ace579d541fda47661803ff44")
	if err != nil {
		panic(err)
	}
	GenesisSeed = prvkey
	GenesisAddress = iotago.Ed25519AddressFromPubKey(GenesisSeed.Public().(ed25519.PublicKey))
}

// DefaultConfig returns the default NodeConfig.
func DefaultConfig() *NodeConfig {
	cfg := &NodeConfig{
		Name: "",
		Envs: []string{"LOGGER_LEVEL=debug"},
		Binds: []string{
			fmt.Sprintf("hornet-testing-assets:%s:rw", assetsDir),
		},
		Network:     DefaultNetworkConfig(),
		Snapshot:    DefaultSnapshotConfig(),
		Protocol:    DefaultProtocolConfig(),
		RestAPI:     DefaultRestAPIConfig(),
		INX:         DefaultINXConfig(),
		Plugins:     DefaultPluginConfig(),
		Profiling:   DefaultProfilingConfig(),
		Dashboard:   DefaultDashboardConfig(),
		Receipts:    DefaultNodeReceiptValidatorConfig(),
		Autopeering: DefaultAutopeeringConfig(),
		INXCoo:      DefaultINXCoordinatorConfig(),
	}
	cfg.ExposedPorts = nat.PortSet{
		nat.Port(fmt.Sprintf("%s/tcp", strings.Split(cfg.RestAPI.BindAddress, ":")[1])): {},
		"6060/tcp": {},
		"8081/tcp": {},
	}
	return cfg
}

// WhiteFlagMockServerConfig defines the config for a white-flag mock server instance.
type WhiteFlagMockServerConfig struct {
	// The name for this white-flag mock server.
	Name string
	// environment variables.
	Envs []string
	// Binds for the container.
	Binds []string
}

// DefaultWhiteFlagMockServerConfig returns the default WhiteFlagMockServerConfig.
func DefaultWhiteFlagMockServerConfig(name string, configFileName string) *WhiteFlagMockServerConfig {
	return &WhiteFlagMockServerConfig{
		Name: name,
		Envs: []string{
			fmt.Sprintf("WHITE_FLAG_MOCK_CONFIG=%s/%s", assetsDir, configFileName),
		},
		Binds: []string{
			fmt.Sprintf("hornet-testing-assets:%s:rw", assetsDir),
		},
	}
}

//TODO: remove when we rename the node config to app in HORNET
// AppPluginConfig defines plugin specific configuration.
type AppPluginConfig struct {
	// Holds explicitly enabled plugins.
	Enabled []string
	// Holds explicitly disabled plugins.
	Disabled []string
}

// CLIFlags returns the config as CLI flags.
func (pluginConfig *AppPluginConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", "app.enablePlugins", strings.Join(pluginConfig.Enabled, ",")),
		fmt.Sprintf("--%s=%s", "app.disablePlugins", strings.Join(pluginConfig.Disabled, ",")),
	}
}

// DefaultAppPluginConfig returns the default plugin config.
func DefaultAppPluginConfig() AppPluginConfig {
	return AppPluginConfig{
		Enabled:  []string{},
		Disabled: []string{},
	}
}

type INXCoordinatorConfig struct {
	// Whether to let the node run as the coordinator.
	RunAsCoo bool
	// The name of this node.
	Name string
	// Environment variables.
	Envs []string
	// Binds for the container.
	Binds []string
	// INXAddress is the INX address of the node.
	INXAddress string
	// Coordinator config.
	Coordinator CoordinatorConfig
	// Plugin config.
	Plugins AppPluginConfig
	// Migrator config.
	Migrator MigratorConfig
	// Receipt validator config.
	Validator ReceiptValidatorConfig
}

// DefaultINXCoordinatorConfig returns the default INX coordinator config.
func DefaultINXCoordinatorConfig() *INXCoordinatorConfig {
	return &INXCoordinatorConfig{
		RunAsCoo: false,
		Name:     "",
		Envs:     []string{"LOGGER_LEVEL=debug"},
		Binds: []string{
			fmt.Sprintf("hornet-testing-assets:%s:rw", assetsDir),
		},
		Coordinator: DefaultCoordinatorConfig(),
		Plugins:     DefaultAppPluginConfig(),
		Migrator:    DefaultMigratorConfig(),
		Validator:   DefaultReceiptValidatorConfig(),
	}
}

// WithMigration adjusts the config to activate the migrator plugin.
func (cfg *INXCoordinatorConfig) WithMigration() {
	cfg.Migrator.Bootstrap = true
	cfg.Plugins.Enabled = append(cfg.Plugins.Enabled, "Migrator")
}

// CLIFlags returns the config as CLI flags.
func (cfg *INXCoordinatorConfig) CLIFlags() []string {
	var cliFlags []string
	cliFlags = append(cliFlags, fmt.Sprintf("--inx.address=%s", cfg.INXAddress))
	cliFlags = append(cliFlags, cfg.Coordinator.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Plugins.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Migrator.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Validator.CLIFlags()...)
	return cliFlags
}

// INXIndexerConfig defines the config of an INX-Indexer.
type INXIndexerConfig struct {
	// The name of this node.
	Name string
	// Environment variables.
	Envs []string
	// Binds for the container.
	Binds       []string
	INXAddress  string
	BindAddress string
}

func DefaultINXIndexerConfig() *INXIndexerConfig {
	return &INXIndexerConfig{
		Name: "",
		Envs: []string{"LOGGER_LEVEL=debug"},
		Binds: []string{
			fmt.Sprintf("hornet-testing-assets:%s:rw", assetsDir),
		},
	}
}

func (cfg *INXIndexerConfig) CLIFlags() []string {
	var cliFlags []string
	cliFlags = append(cliFlags, fmt.Sprintf("--inx.address=%s", cfg.INXAddress))
	cliFlags = append(cliFlags, fmt.Sprintf("--indexer.bindAddress=%s", cfg.BindAddress))
	return cliFlags
}

// NodeConfig defines the config of a HORNET node.
type NodeConfig struct {
	// The name of this node.
	Name string
	// Environment variables.
	Envs []string
	// Binds for the container.
	Binds []string
	// Exposed ports of this container.
	ExposedPorts nat.PortSet
	// Network config.
	Network NetworkConfig
	// Web API config.
	RestAPI RestAPIConfig
	// INX config.
	INX INXConfig
	// Snapshot config.
	Snapshot SnapshotConfig
	// Protocol config.
	Protocol ProtocolConfig
	// Plugin config.
	Plugins PluginConfig
	// Profiling config.
	Profiling ProfilingConfig
	// Dashboard config.
	Dashboard DashboardConfig
	// Receipts config
	Receipts ReceiptsConfig
	// Autopeering config.
	Autopeering AutopeeringConfig
	// INXCoo inx-coordinator config.
	INXCoo *INXCoordinatorConfig
}

// AsCoo adjusts the config to make it usable as the Coordinator's config.
func (cfg *NodeConfig) AsCoo() {
	cfg.Plugins.Enabled = append(cfg.Plugins.Enabled, "INX")
	cfg.INXCoo.RunAsCoo = true
	cfg.INXCoo.Envs = append(cfg.INXCoo.Envs, fmt.Sprintf("COO_PRV_KEYS=%s", strings.Join(cfg.INXCoo.Coordinator.PrivateKeys, ",")))
}

// WithReceipts adjusts the config to activate the receipts plugin.
func (cfg *NodeConfig) WithReceipts() {
	cfg.Plugins.Enabled = append(cfg.Plugins.Enabled, "Receipts")
	cfg.INXCoo.WithMigration()
}

// CLIFlags returns the config as CLI flags.
func (cfg *NodeConfig) CLIFlags() []string {
	var cliFlags []string
	cliFlags = append(cliFlags, cfg.Network.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Snapshot.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Protocol.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.RestAPI.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.INX.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Plugins.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Profiling.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Dashboard.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Receipts.CLIFlags()...)
	cliFlags = append(cliFlags, cfg.Autopeering.CLIFlags()...)
	return cliFlags
}

// NetworkConfig defines the network specific configuration.
type NetworkConfig struct {
	// the private key used to derive the node identity.
	IdentityPrivKey string
	// the bind addresses of this node.
	BindMultiAddresses []string
	// the path to the p2p database.
	DatabasePath string
	// the high watermark to use within the connection manager.
	ConnMngHighWatermark int
	// the low watermark to use within the connection manager.
	ConnMngLowWatermark int
	// the static peers this node should retain a connection to.
	Peers []string
	// aliases of the static peers.
	PeerAliases []string
	// time to wait before trying to reconnect to a disconnected peer.
	ReconnectInterval time.Duration
	// the maximum amount of unknown peers a gossip protocol connection is established to
	GossipUnknownPeersLimit int
}

// CLIFlags returns the config as CLI flags.
func (netConfig *NetworkConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", p2p.CfgP2PIdentityPrivKey, netConfig.IdentityPrivKey),
		fmt.Sprintf("--%s=%s", p2p.CfgP2PBindMultiAddresses, strings.Join(netConfig.BindMultiAddresses, ",")),
		fmt.Sprintf("--%s=%s", p2p.CfgP2PDatabasePath, netConfig.DatabasePath),
		fmt.Sprintf("--%s=%d", p2p.CfgP2PConnMngHighWatermark, netConfig.ConnMngHighWatermark),
		fmt.Sprintf("--%s=%d", p2p.CfgP2PConnMngLowWatermark, netConfig.ConnMngLowWatermark),
		fmt.Sprintf("--%s=%s", p2p.CfgP2PPeers, strings.Join(netConfig.Peers, ",")),
		fmt.Sprintf("--%s=%s", p2p.CfgP2PPeerAliases, strings.Join(netConfig.PeerAliases, ",")),
		fmt.Sprintf("--%s=%s", p2p.CfgP2PReconnectInterval, netConfig.ReconnectInterval),
		fmt.Sprintf("--%s=%d", gossip.CfgP2PGossipUnknownPeersLimit, netConfig.GossipUnknownPeersLimit),
	}
}

// DefaultNetworkConfig returns the default network config.
func DefaultNetworkConfig() NetworkConfig {
	return NetworkConfig{
		IdentityPrivKey:         "",
		BindMultiAddresses:      []string{"/ip4/0.0.0.0/tcp/15600"},
		DatabasePath:            "p2pstore",
		ConnMngHighWatermark:    8,
		ConnMngLowWatermark:     4,
		Peers:                   []string{},
		PeerAliases:             []string{},
		ReconnectInterval:       1 * time.Second,
		GossipUnknownPeersLimit: 4,
	}
}

// AutopeeringConfig defines the autopeering specific configuration.
type AutopeeringConfig struct {
	// The ist of autopeering entry nodes to use.
	EntryNodes []string
	// BindAddr bind address for autopeering.
	BindAddr string
	// Whether the node should act as an autopeering entry node.
	RunAsEntryNode bool
	// The max number of inbound autopeers.
	InboundPeers int
	// The max the number of outbound autopeers.
	OutboundPeers int
	// The lifetime of the private and public local salt.
	SaltLifetime time.Duration
}

// CLIFlags returns the config as CLI flags.
func (autoConfig *AutopeeringConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", autopeering.CfgNetAutopeeringEntryNodes, strings.Join(autoConfig.EntryNodes, ",")),
		fmt.Sprintf("--%s=%s", autopeering.CfgNetAutopeeringBindAddr, autoConfig.BindAddr),
		fmt.Sprintf("--%s=%v", autopeering.CfgNetAutopeeringRunAsEntryNode, autoConfig.RunAsEntryNode),
		fmt.Sprintf("--%s=%d", autopeering.CfgNetAutopeeringInboundPeers, autoConfig.InboundPeers),
		fmt.Sprintf("--%s=%d", autopeering.CfgNetAutopeeringOutboundPeers, autoConfig.OutboundPeers),
		fmt.Sprintf("--%s=%s", autopeering.CfgNetAutopeeringSaltLifetime, autoConfig.SaltLifetime),
	}
}

// DefaultAutopeeringConfig returns the default autopeering config.
func DefaultAutopeeringConfig() AutopeeringConfig {
	return AutopeeringConfig{
		EntryNodes:     nil,
		BindAddr:       "0.0.0.0:14626",
		RunAsEntryNode: false,
		InboundPeers:   2,
		OutboundPeers:  2,
		SaltLifetime:   30 * time.Minute,
	}
}

// RestAPIConfig defines the REST API specific configuration.
type RestAPIConfig struct {
	// The bind address for the REST API.
	BindAddress string
	// Public REST API routes.
	PublicRoutes []string
	// Protected REST API routes.
	ProtectedRoutes []string
	// Whether the node does proof-of-work for submitted messages.
	PoWEnabled bool
}

// CLIFlags returns the config as CLI flags.
func (restAPIConfig *RestAPIConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", restapi.CfgRestAPIBindAddress, restAPIConfig.BindAddress),
		fmt.Sprintf("--%s=%s", restapi.CfgRestAPIPublicRoutes, strings.Join(restAPIConfig.PublicRoutes, ",")),
		fmt.Sprintf("--%s=%s", restapi.CfgRestAPIProtectedRoutes, strings.Join(restAPIConfig.ProtectedRoutes, ",")),
		fmt.Sprintf("--%s=%v", restapi.CfgRestAPIPoWEnabled, restAPIConfig.PoWEnabled),
	}
}

// DefaultRestAPIConfig returns the default REST API config.
func DefaultRestAPIConfig() RestAPIConfig {
	return RestAPIConfig{
		BindAddress: "0.0.0.0:14265",
		PublicRoutes: []string{
			"/health",
			"/api/v2/*",
			"/api/plugins/*",
		},
		ProtectedRoutes: []string{},
		PoWEnabled:      true,
	}
}

// INXConfig defines the INX specific configuration.
type INXConfig struct {
	// The bind address for the INX.
	BindAddress string
}

// CLIFlags returns the config as CLI flags.
func (inxConfig *INXConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", inx.CfgINXBindAddress, inxConfig.BindAddress),
	}
}

// DefaultINXConfig returns the default INX config.
func DefaultINXConfig() INXConfig {
	return INXConfig{
		BindAddress: "0.0.0.0:9029",
	}
}

// PluginConfig defines plugin specific configuration.
type PluginConfig struct {
	// Holds explicitly enabled plugins.
	Enabled []string
	// Holds explicitly disabled plugins.
	Disabled []string
}

func (pluginConfig *PluginConfig) ContainsINX() bool {
	for _, p := range pluginConfig.Enabled {
		if strings.EqualFold(p, "INX") {
			return true
		}
	}
	return false
}

// CLIFlags returns the config as CLI flags.
func (pluginConfig *PluginConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", app.CfgNodeEnablePlugins, strings.Join(pluginConfig.Enabled, ",")),
		fmt.Sprintf("--%s=%s", app.CfgNodeDisablePlugins, strings.Join(pluginConfig.Disabled, ",")),
	}
}

// DefaultPluginConfig returns the default plugin config.
func DefaultPluginConfig() PluginConfig {
	disabled := make([]string, len(disabledPluginsPeer))
	copy(disabled, disabledPluginsPeer)
	return PluginConfig{
		Enabled:  []string{},
		Disabled: disabled,
	}
}

// SnapshotConfig defines snapshot specific configuration.
type SnapshotConfig struct {
	// The path to the full snapshot file.
	FullSnapshotFilePath string
	// the path to the delta snapshot file.
	DeltaSnapshotFilePath string
}

// CLIFlags returns the config as CLI flags.
func (snapshotConfig *SnapshotConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", snapshot.CfgSnapshotsFullPath, snapshotConfig.FullSnapshotFilePath),
		fmt.Sprintf("--%s=%s", snapshot.CfgSnapshotsDeltaPath, snapshotConfig.DeltaSnapshotFilePath),
	}
}

// DefaultSnapshotConfig returns the default snapshot config.
func DefaultSnapshotConfig() SnapshotConfig {
	return SnapshotConfig{
		FullSnapshotFilePath:  "/assets/full_snapshot.bin",
		DeltaSnapshotFilePath: "/assets/delta_snapshot.bin",
	}
}

// CoordinatorConfig defines coordinator specific configuration.
type CoordinatorConfig struct {
	// The coo private keys.
	PrivateKeys []string
	// Whether to run the coordinator in bootstrap node.
	Bootstrap bool
	// The interval in which to issue new milestones.
	IssuanceInterval time.Duration
}

// CLIFlags returns the config as CLI flags.
func (cooConfig *CoordinatorConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--cooBootstrap=%v", cooConfig.Bootstrap),
		fmt.Sprintf("--%s=%s", coordinator.CfgCoordinatorInterval, cooConfig.IssuanceInterval),
	}
}

// DefaultCoordinatorConfig returns the default coordinator config.
func DefaultCoordinatorConfig() CoordinatorConfig {
	return CoordinatorConfig{
		Bootstrap: true,
		PrivateKeys: []string{"651941eddb3e68cb1f6ef4ef5b04625dcf5c70de1fdc4b1c9eadb2c219c074e0ed3c3f1a319ff4e909cf2771d79fece0ac9bd9fd2ee49ea6c0885c9cb3b1248c",
			"0e324c6ff069f31890d496e9004636fd73d8e8b5bea08ec58a4178ca85462325f6752f5f46a53364e2ee9c4d662d762a81efd51010282a75cd6bd03f28ef349c"},
		IssuanceInterval: 10 * time.Second,
	}
}

// ReceiptsConfig defines the receipt validator plugin specific configuration.
type ReceiptsConfig struct {
	// Whether receipt backups are enabled.
	BackupEnabled bool
	// The path to the receipts folder.
	BackupFolder string
	// Whether the receipts plugin should validate receipts
	Validate bool
	// Whether to ignore soft errors or not.
	IgnoreSoftErrors bool
	// The validator config
	Validator ReceiptValidatorConfig
}

func (receiptsConfig *ReceiptsConfig) CLIFlags() []string {
	flags := []string{
		fmt.Sprintf("--%s=%v", receipt.CfgReceiptsBackupEnabled, receiptsConfig.BackupEnabled),
		fmt.Sprintf("--%s=%s", receipt.CfgReceiptsBackupPath, receiptsConfig.BackupFolder),
		fmt.Sprintf("--%s=%v", receipt.CfgReceiptsValidatorValidate, receiptsConfig.Validate),
		fmt.Sprintf("--%s=%v", receipt.CfgReceiptsValidatorIgnoreSoftErrors, receiptsConfig.IgnoreSoftErrors),
	}
	flags = append(flags, receiptsConfig.Validator.CLIFlags()...)
	return flags
}

type ReceiptValidatorConfig struct {
	// The API to query.
	APIAddress string
	// The API timeout.
	APITimeout time.Duration
	// Legacy Coordinator address
	CoordinatorAddress string
	// The merkle tree depth.
	CoordinatorMerkleTreeDepth int
}

func DefaultNodeReceiptValidatorConfig() ReceiptsConfig {
	return ReceiptsConfig{
		BackupEnabled:    false,
		BackupFolder:     "receipts",
		Validate:         false,
		IgnoreSoftErrors: false,
		Validator:        DefaultReceiptValidatorConfig(),
	}
}

func (validatorConfig *ReceiptValidatorConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", receipt.CfgReceiptsValidatorAPIAddress, validatorConfig.APIAddress),
		fmt.Sprintf("--%s=%s", receipt.CfgReceiptsValidatorAPITimeout, validatorConfig.APITimeout),
		fmt.Sprintf("--%s=%s", receipt.CfgReceiptsValidatorCoordinatorAddress, validatorConfig.CoordinatorAddress),
		fmt.Sprintf("--%s=%d", receipt.CfgReceiptsValidatorCoordinatorMerkleTreeDepth, validatorConfig.CoordinatorMerkleTreeDepth),
	}
}

// DefaultReceiptValidatorConfig returns the default receipt validator config.
func DefaultReceiptValidatorConfig() ReceiptValidatorConfig {
	return ReceiptValidatorConfig{
		APIAddress:                 "http://localhost:14265",
		APITimeout:                 5 * time.Second,
		CoordinatorAddress:         "JFQ999DVN9CBBQX9DSAIQRAFRALIHJMYOXAQSTCJLGA9DLOKIWHJIFQKMCQ9QHWW9RXQMDBVUIQNIY9GZ",
		CoordinatorMerkleTreeDepth: 18,
	}
}

// MigratorConfig defines migrator plugin specific configuration.
type MigratorConfig struct {
	// The max amount of entries to include in a receipt.
	MaxEntries int
	// Whether to run the migrator plugin in bootstrap mode.
	Bootstrap bool
	// The index of the first legacy milestone to migrate.
	StartIndex int
	// The state file path.
	StateFilePath string
}

// CLIFlags returns the config as CLI flags.
func (migConfig *MigratorConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%v", migrator.CfgMigratorBootstrap, migConfig.Bootstrap),
		fmt.Sprintf("--%s=%v", migrator.CfgMigratorReceiptMaxEntries, migConfig.MaxEntries),
		fmt.Sprintf("--%s=%d", migrator.CfgMigratorStartIndex, migConfig.StartIndex),
		fmt.Sprintf("--%s=%s", migrator.CfgMigratorStateFilePath, migConfig.StateFilePath),
	}
}

// DefaultMigratorConfig returns the default migrator plugin config.
func DefaultMigratorConfig() MigratorConfig {
	return MigratorConfig{
		Bootstrap:     false,
		MaxEntries:    iotago.MaxMigratedFundsEntryCount,
		StartIndex:    1,
		StateFilePath: "migrator.state",
	}
}

// ProtocolConfig defines protocol specific configuration.
type ProtocolConfig struct {
	// The minimum PoW score needed.
	MinPoWScore float64
	// The coo public key ranges.
	PublicKeyRanges []protocfg.ConfigPublicKeyRange
	// The network ID on which this node operates on.
	NetworkIDName string
	// The HRP which should be used for Bech32 addresses.
	Bech32HRP iotago.NetworkPrefix
}

// CLIFlags returns the config as CLI flags.
func (protoConfig *ProtocolConfig) CLIFlags() []string {

	keyRangesJSON, err := json.Marshal(protoConfig.PublicKeyRanges)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal COO public key ranges: %s", err))
	}

	return []string{
		fmt.Sprintf("--%s=%0.0f", protocfg.CfgProtocolMinPoWScore, protoConfig.MinPoWScore),
		fmt.Sprintf("--%s=%s", protocfg.CfgProtocolPublicKeyRangesJSON, string(keyRangesJSON)),
		fmt.Sprintf("--%s=%s", protocfg.CfgProtocolNetworkIDName, protoConfig.NetworkIDName),
		fmt.Sprintf("--%s=%s", protocfg.CfgProtocolBech32HRP, protoConfig.Bech32HRP),
	}
}

// DefaultProtocolConfig returns the default protocol config.
func DefaultProtocolConfig() ProtocolConfig {
	return ProtocolConfig{
		MinPoWScore: 100,
		PublicKeyRanges: []protocfg.ConfigPublicKeyRange{
			{
				Key:        "ed3c3f1a319ff4e909cf2771d79fece0ac9bd9fd2ee49ea6c0885c9cb3b1248c",
				StartIndex: 0,
				EndIndex:   0,
			},
			{
				Key:        "f6752f5f46a53364e2ee9c4d662d762a81efd51010282a75cd6bd03f28ef349c",
				StartIndex: 0,
				EndIndex:   0,
			},
		},
		NetworkIDName: "alphanet1",
		Bech32HRP:     iotago.PrefixTestnet,
	}
}

// ProfilingConfig defines the profiling specific configuration.
type ProfilingConfig struct {
	// The bind address of the pprof server.
	BindAddress string
}

// CLIFlags returns the config as CLI flags.
func (profilingConfig *ProfilingConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", profiling.CfgProfilingBindAddress, profilingConfig.BindAddress),
	}
}

// DefaultProfilingConfig returns the default profiling config.
func DefaultProfilingConfig() ProfilingConfig {
	return ProfilingConfig{
		BindAddress: "0.0.0.0:6060",
	}
}

// DashboardConfig holds the dashboard specific configuration.
type DashboardConfig struct {
	// The bind address of the dashboard
	BindAddress string
}

// CLIFlags returns the config as CLI flags.
func (dashboardConfig *DashboardConfig) CLIFlags() []string {
	return []string{
		fmt.Sprintf("--%s=%s", dashboard.CfgDashboardBindAddress, dashboardConfig.BindAddress),
	}
}

// DefaultDashboardConfig returns the default profiling config.
func DefaultDashboardConfig() DashboardConfig {
	return DashboardConfig{
		BindAddress: "0.0.0.0:8081",
	}
}
