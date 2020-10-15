package deejdsp

// Pretty much ripped from origonal deej repo with changes to suit the displays

import (
	"fmt"
	"io/ioutil"

	"github.com/jax-b/deej/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// DSPCanonicalConfig config of DSP
type DSPCanonicalConfig struct {
	DisplayMapping         map[int]string
	StartupDelay           int
	logger                 *zap.SugaredLogger
	CommandDelay           int
	BWThreshold            int
	IconFinderDotComAPIKey string
}

type marshalledConfig struct {
	DisplayMapping         map[int]interface{} `yaml:"display_mapping"`
	StartupDelay           int                 `yaml:"startup_delay"`
	CommandDelay           int                 `yaml:"command_delay"`
	BWThreshold            int                 `yaml:"BlackWhite_Threshold"`
	IconFinderDotComAPIKey string              `yaml:"IconFinderDotComAPIKey"`
}

const configFilepath = "config.yaml"

// NewDSPConfig creates the config object
func NewDSPConfig(logger *zap.SugaredLogger) (*DSPCanonicalConfig, error) {
	logger = logger.Named("config")

	cc := &DSPCanonicalConfig{
		logger: logger,
	}

	return cc, nil
}

// Load reads a config file from disk and tries to parse it
func (cc *DSPCanonicalConfig) Load() error {
	cc.logger.Debugw("Loading config", "path", configFilepath)

	// make sure it exists
	if !util.FileExists(configFilepath) {
		cc.logger.Warnw("Config file not found", "path", configFilepath)
		return fmt.Errorf("config file doesn't exist: %s", configFilepath)
	}

	// open->read->close the file
	configBytes, err := ioutil.ReadFile(configFilepath)
	if err != nil {
		cc.logger.Warnw("Failed to read config file", "error", err)
		return fmt.Errorf("read config file: %w", err)
	}

	// unmarshall it into the yaml-aware struct
	mc := &marshalledConfig{}
	if err := yaml.Unmarshal(configBytes, mc); err != nil {
		cc.logger.Warnw("Failed to unmarhsal config into struct", "error", err)

		return fmt.Errorf("unmarshall yaml config: %w", err)
	}

	// canonize it
	if err := cc.populateFromMarshalled(mc); err != nil {
		cc.logger.Warnw("Failed to populate config fields from marshalled struct", "error", err)
		return fmt.Errorf("populate config fields: %w", err)
	}

	cc.logger.Info("Loaded config successfully")
	cc.logger.Infow("Config values",
		"DisplayMapping", cc.DisplayMapping, "StartupDelay", cc.StartupDelay, "CommandDelay", cc.CommandDelay, "BWThreshold", cc.BWThreshold)

	return nil
}

func (cc *DSPCanonicalConfig) populateFromMarshalled(mc *marshalledConfig) error {

	// start by loading the slider mapping because it's the only failable part for now
	if mc.DisplayMapping == nil {
		cc.logger.Warnw("Missing key in config, using default value",
			"key", "display_mapping",
			"value", map[int]string{0: ""})

		cc.DisplayMapping = map[int]string{0: ""}
	} else {
		cc.DisplayMapping = make(map[int]string)
		// this is where we need to parse out each value (which is an interface{} at this point),
		// and type-assert it into either a string or a list of strings
		for key, value := range mc.DisplayMapping {
			switch typedValue := value.(type) {
			case string:
				if typedValue == "" {

				} else {
					cc.DisplayMapping[key] = typedValue
				}

			// silently ignore nil values and treat as no targets
			case nil:
				cc.DisplayMapping[key] = ""
			default:
				cc.logger.Warnw("Invalid value for slider mapping key",
					"key", key,
					"value", typedValue,
					"valueType", fmt.Sprintf("%t", typedValue))

				return fmt.Errorf("invalid slider mapping for slider %d: got type %t, need string or []string", key, typedValue)
			}
		}
	}

	if mc.StartupDelay <= 0 {
		cc.logger.Warnw("Missing key in config, using default value",
			"key", "startup_delay",
			"value", mc.StartupDelay)
		cc.StartupDelay = 50
	} else {
		cc.StartupDelay = mc.StartupDelay
	}

	if mc.CommandDelay <= 0 {
		cc.logger.Warnw("Missing key in config, using default value",
			"key", "command_delay",
			"value", mc.CommandDelay)
		cc.CommandDelay = 0
	} else {
		cc.CommandDelay = mc.CommandDelay
	}

	if mc.BWThreshold <= 0 {
		cc.logger.Warnw("Missing key in config, using default value",
			"key", "BlackWhite_Threshold",
			"value", mc.BWThreshold)
		cc.BWThreshold = 200
	} else {
		cc.BWThreshold = mc.BWThreshold
	}

	if mc.IconFinderDotComAPIKey == "" {
		cc.logger.Warnw("Missing key in config, cannot grab icons",
			"key", "iconFinderDotComAPIKey")
		cc.IconFinderDotComAPIKey = ""
	} else {
		cc.IconFinderDotComAPIKey = mc.IconFinderDotComAPIKey
	}

	return nil
}
