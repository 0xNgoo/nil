//nolint:tagliatelle
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NilFoundation/nil/nil/cmd/nild/nildconfig"
	"github.com/NilFoundation/nil/nil/internal/config"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/keys"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/telemetry"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type nodeSpec struct {
	ID                   int    `yaml:"id"`
	Shards               []uint `yaml:"shards"`
	SplitShards          bool   `yaml:"splitShards"`
	BootstrapPeersIdx    []int  `yaml:"bootstrapPeersIdx"`
	DHTBootstrapPeersIdx []int  `yaml:"dhtBootstrapPeersIdx"`
	ArchiveNodeIndices   []int  `yaml:"archiveNodeIndices"`
}

type devnetSpec struct {
	NilServerName          string   `yaml:"nil_server_name"`
	NilCertEmail           string   `yaml:"nil_cert_email"`
	NildConfigDir          string   `yaml:"nild_config_dir"`
	NildCredentialsDir     string   `yaml:"nild_credentials_dir"`
	NildP2PBaseTCPPort     int      `yaml:"nild_p2p_base_tcp_port"`
	NilWipeOnUpdate        bool     `yaml:"nil_wipe_on_update"`
	NShards                uint32   `yaml:"nShards"`
	NilRPCHost             string   `yaml:"nil_rpc_host"`
	NilRPCPort             int      `yaml:"nil_rpc_port"`
	EnableRPCOnValidators  bool     `yaml:"nil_rpc_enable_on_validators"`
	ClickhouseHost         string   `yaml:"clickhouse_host"`
	ClickhousePort         int      `yaml:"clickhouse_port"`
	ClickhouseLogin        string   `yaml:"clickhouse_login"`
	ClickhouseDatabase     string   `yaml:"clickhouse_database"`
	CometaRPCHost          string   `yaml:"cometa_rpc_host"`
	CometaPort             int      `yaml:"cometa_port"`
	FaucetRPCHost          string   `yaml:"faucet_rpc_host"`
	FaucetPort             int      `yaml:"faucet_port"`
	NilLoadgenHost         string   `yaml:"nil_loadgen_host"`
	NilLoadgenPort         int      `yaml:"nil_loadgen_port"`
	NilUpdateRetryInterval int      `yaml:"nil_update_retry_interval_sec"`
	InstanceEnv            string   `yaml:"instance_env"`
	SignozJournaldLogs     []string `yaml:"signoz_journald_logs"`

	NilConfig        []nodeSpec `yaml:"nil_config"`
	NilArchiveConfig []nodeSpec `yaml:"nil_archive_config"`
	NilRPCConfig     []nodeSpec `yaml:"nil_rpc_config"`

	NilLoadGeneratorsEnable bool `yaml:"nil_load_generators_enable"`
}

type server struct {
	service         string
	name            string
	identity        string
	port            int
	rpcPort         int
	credsDir        string
	workDir         string
	nodeSpec        nodeSpec
	logClientEvents bool
	vkm             *keys.ValidatorKeysManager
}

type devnet struct {
	spec       *devnetSpec
	baseDir    string
	validators []server
	archivers  []server
	rpcNodes   []server
	zeroState  *execution.ZeroStateConfig
}

func DevnetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gen-configs",
		Short:        "Generate devnet configuration",
		Args:         cobra.ExactArgs(1),
		RunE:         genDevnet,
		SilenceUsage: true,
	}
	cmd.Flags().String("basedir", "/var/lib", "Base directory for devnet")
	cmd.Flags().String("only", "", "Write only specified service (e.g. nil-archive-1)")
	return cmd
}

func validatorKeysFile(credsDir string) string {
	return filepath.Join(credsDir, "validator-keys.yaml")
}

func (spec *devnetSpec) ensureValidatorKeys(srv *server) (*keys.ValidatorKeysManager, error) {
	vkm := keys.NewValidatorKeyManager(validatorKeysFile(srv.credsDir))
	if err := vkm.InitKey(); err != nil {
		return nil, err
	}
	return vkm, nil
}

func (devnet devnet) generateZeroState(nShards uint32, servers []server) (*execution.ZeroStateConfig, error) {
	validators := make([]config.ListValidators, nShards-1)
	for _, srv := range servers {
		key, err := srv.vkm.GetPublicKey()
		if err != nil {
			return nil, err
		}

		for _, id := range srv.nodeSpec.Shards {
			if id == 0 {
				continue
			}
			idx := id - 1
			validators[idx].List = append(validators[idx].List, config.ValidatorInfo{
				PublicKey: config.Pubkey(key),
			})
		}
	}
	return &execution.ZeroStateConfig{
		ConfigParams: execution.ConfigParams{
			Validators: config.ParamValidators{Validators: validators},
		},
	}, nil
}

func genDevnet(cmd *cobra.Command, args []string) error {
	baseDir, err := cmd.Flags().GetString("basedir")
	if err != nil {
		return fmt.Errorf("failed to get basedir flag: %w", err)
	}
	baseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for basedir: %w", err)
	}

	specFile := args[0]
	specYaml, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("can't read devnet spec %s: %w", specFile, err)
	}

	spec := &devnetSpec{}
	if err := yaml.Unmarshal(specYaml, spec); err != nil {
		return fmt.Errorf("can't parse devnet spec %s: %w", specFile, err)
	}

	validatorRPCBasePort := 0
	if spec.EnableRPCOnValidators {
		validatorRPCBasePort = spec.NilRPCPort + len(spec.NilRPCConfig)
	}
	validators, err := spec.makeServers(spec.NilConfig, spec.NildP2PBaseTCPPort, validatorRPCBasePort, "nil", baseDir, false)
	if err != nil {
		return fmt.Errorf("failed to setup validator nodes: %w", err)
	}

	devnet := devnet{spec: spec, baseDir: baseDir, validators: validators}

	archiveBaseP2P := spec.NildP2PBaseTCPPort + len(validators)

	if devnet.archivers, err = spec.makeServers(spec.NilArchiveConfig, archiveBaseP2P, 0, "nil-archive", baseDir, false); err != nil {
		return fmt.Errorf("failed to setup archive nodes: %w", err)
	}

	if devnet.rpcNodes, err = spec.makeServers(spec.NilRPCConfig, 0, spec.NilRPCPort, "nil-rpc", baseDir, true); err != nil {
		return fmt.Errorf("failed to setup rpc nodes: %w", err)
	}

	only, err := cmd.Flags().GetString("only")
	if err != nil {
		return fmt.Errorf("failed to get only flag: %w", err)
	}

	if devnet.zeroState, err = devnet.generateZeroState(spec.NShards, devnet.validators); err != nil {
		return err
	}
	if err := devnet.writeConfigs(devnet.validators, "validator", only); err != nil {
		return err
	}
	if err := devnet.writeConfigs(devnet.archivers, "archiver", only); err != nil {
		return err
	}
	if err := devnet.writeConfigs(devnet.rpcNodes, "RPC node", only); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func (spec *devnetSpec) EnsureValidatorKeys(srv server) (*keys.ValidatorKeysManager, error) {
	if err := os.MkdirAll(srv.credsDir, directoryPermissions); err != nil {
		return nil, err
	}

	return spec.ensureValidatorKeys(&srv)
}

func (devnet devnet) writeConfigs(servers []server, name string, only string) error {
	for i, server := range servers {
		if err := devnet.writeServerConfig(i, server, only); err != nil {
			return fmt.Errorf("failed to write %s config: %w", name, err)
		}
	}
	return nil
}

func (spec devnetSpec) makeServers(nodeSpecs []nodeSpec, basePort int, baseHTTPPort int, service string, baseDir string, logClientEvents bool) ([]server, error) {
	servers := make([]server, len(nodeSpecs))
	for i, nodeSpec := range nodeSpecs {
		servers[i].service = service
		servers[i].name = fmt.Sprintf("%s-%d", service, i)
		servers[i].nodeSpec = nodeSpec
		servers[i].logClientEvents = logClientEvents
		if basePort != 0 {
			servers[i].port = basePort + i
		}
		if baseHTTPPort != 0 {
			servers[i].rpcPort = baseHTTPPort + i
		}
		servers[i].credsDir = fmt.Sprintf("%s/%s-%d", spec.NildCredentialsDir, service, i)
		servers[i].workDir = fmt.Sprintf("%s/%s-%d", baseDir, service, i)
		var err error
		servers[i].identity, err = spec.EnsureIdentity(servers[i])
		if err != nil {
			return nil, err
		}
		servers[i].vkm, err = spec.EnsureValidatorKeys(servers[i])
		if err != nil {
			return nil, err
		}
	}
	return servers, nil
}

func (srv server) NetworkKeysFile() string {
	return srv.credsDir + "/network-keys.yaml"
}

const (
	directoryPermissions = 0o755
	filePermissions      = 0o644
)

func (spec *devnetSpec) EnsureIdentity(srv server) (string, error) {
	if err := os.MkdirAll(srv.credsDir, directoryPermissions); err != nil {
		return "", err
	}
	privKey, err := network.LoadOrGenerateKeys(srv.NetworkKeysFile())
	if err != nil {
		return "", fmt.Errorf("failed to load or generate keys: %w", err)
	}
	_, _, identity, err := network.SerializeKeys(privKey)
	return identity.String(), err
}

func (devnet *devnet) writeServerConfig(instanceId int, srv server, only string) error {
	if only != "" && srv.name != only {
		return nil
	}

	cfg := nildconfig.Config{}
	cfg.Config = &nilservice.Config{
		Network:   &network.Config{},
		Telemetry: &telemetry.Config{},
	}
	cfg.RPCPort = srv.rpcPort
	cfg.Network.TcpPort = srv.port
	cfg.ZeroState = devnet.zeroState

	var err error
	cfg.NetworkKeysPath, err = filepath.Abs(srv.NetworkKeysFile())
	if err != nil {
		return fmt.Errorf("failed to get absolute path for network keys: %w", err)
	}

	cfg.ValidatorKeysPath, err = filepath.Abs(validatorKeysFile(srv.credsDir))
	if err != nil {
		return fmt.Errorf("failed to get absolute path for validator keys: %w", err)
	}

	spec := devnet.spec
	cfg.NShards = spec.NShards
	inst := srv.nodeSpec
	cfg.MyShards = inst.Shards
	cfg.SplitShards = inst.SplitShards
	cfg.BootstrapPeers = devnet.getPeers(devnet.validators, inst.BootstrapPeersIdx)
	cfg.AdminSocketPath = srv.workDir + "/admin_socket"
	cfg.LogClientRpcEvents = srv.logClientEvents
	cfg.DB = db.NewDefaultBadgerDBOptions()
	cfg.DB.Path = srv.workDir + "/database"
	cfg.DB.AllowDrop = spec.NilWipeOnUpdate
	cfg.Network.DHTEnabled = true
	cfg.Network.DHTBootstrapPeers = devnet.getPeers(devnet.validators, inst.DHTBootstrapPeersIdx)

	if len(inst.ArchiveNodeIndices) > 0 {
		cfg.RpcNode = &nilservice.RpcNodeConfig{
			ArchiveNodeList: devnet.getPeers(devnet.archivers, inst.ArchiveNodeIndices),
		}
	}

	cfg.Telemetry.ExportMetrics = true

	serialized, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.MkdirAll(srv.workDir, directoryPermissions)
	if err != nil {
		return err
	}

	configDir := fmt.Sprintf("%s/%s-%d", spec.NildConfigDir, srv.service, instanceId)
	err = os.MkdirAll(configDir, directoryPermissions)
	if err != nil {
		return err
	}
	return os.WriteFile(configDir+"/nild.yaml", serialized, filePermissions)
}

func identityToAddress(port int, identity string) string {
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, identity)
}

func (devnet *devnet) getPeers(servers []server, indices []int) network.AddrInfoSlice {
	peers := make(network.AddrInfoSlice, len(indices))
	for i, idx := range indices {
		srv := servers[idx]
		address := identityToAddress(srv.port, srv.identity)
		peers[i].Set(address) //nolint:errcheck
	}
	return peers
}
