package networkpolicy

import (
	"os"
	"path/filepath"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/viper"
)

// nodeConfig holds the parsed, policy-relevant configuration of an
// initialized node home. It never carries secret key material: only the
// public node ID derived from node_key.json is kept.
type nodeConfig struct {
	cmt      *cmtcfg.Config
	app      *serverconfig.Config
	selfID   string
	hasSelf  bool
	cmtPath  string
	appPath  string
	keyPath  string
	loadErrs []Violation
}

// loadHome reads config/config.toml, config/app.toml, and the public identity
// of config/node_key.json from an initialized node home. Parsing starts from
// the same default configuration the daemon itself applies, so omitted fields
// are judged at their effective runtime values. Failures are reported as
// violations without echoing file contents.
func loadHome(home string) *nodeConfig {
	configDir := filepath.Join(home, "config")
	nc := &nodeConfig{
		cmtPath: filepath.Join(configDir, "config.toml"),
		appPath: filepath.Join(configDir, "app.toml"),
		keyPath: filepath.Join(configDir, "node_key.json"),
	}

	if _, err := os.Stat(nc.cmtPath); err != nil {
		nc.loadErrs = append(nc.loadErrs, Violation{
			Check:   "config.toml",
			Message: "not an initialized node home: config/config.toml is missing; run init first",
		})
	} else if cfg, err := parseCometConfig(nc.cmtPath); err != nil {
		nc.loadErrs = append(nc.loadErrs, Violation{
			Check:   "config.toml",
			Message: "cannot parse CometBFT config.toml; check TOML syntax and supported value types",
		})
	} else {
		nc.cmt = cfg
	}

	if _, err := os.Stat(nc.appPath); err != nil {
		nc.loadErrs = append(nc.loadErrs, Violation{
			Check:   "app.toml",
			Message: "not an initialized node home: config/app.toml is missing; run init first",
		})
	} else if cfg, err := parseAppConfig(nc.appPath); err != nil {
		nc.loadErrs = append(nc.loadErrs, Violation{
			Check:   "app.toml",
			Message: "cannot parse Cosmos app.toml; check TOML syntax and supported value types",
		})
	} else {
		nc.app = cfg
	}

	// The node key is read only to derive the public node ID for self-conflict
	// detection. Its private material is never stored, printed, or returned.
	nodeKey, err := p2p.LoadNodeKey(nc.keyPath)
	if err != nil {
		nc.loadErrs = append(nc.loadErrs, Violation{
			Check:   "node_key.json",
			Message: "cannot establish node identity from config/node_key.json; run init first",
		})
	} else {
		nc.selfID = string(nodeKey.ID())
		nc.hasSelf = true
	}
	return nc
}

// parseCometConfig unmarshals config.toml onto the CometBFT defaults, exactly
// as the daemon startup path does.
func parseCometConfig(path string) (*cmtcfg.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := cmtcfg.DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// parseAppConfig unmarshals app.toml onto the SDK defaults via the SDK's own
// config parser.
func parseAppConfig(path string) (*serverconfig.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg, err := serverconfig.GetConfig(v)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
