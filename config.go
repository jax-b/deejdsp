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
	DisplayMapping *displayMap
	logger         *zap.SugaredLogger
}

type marshalledConfig struct {
	DisplayMapping map[int]interface{} `yaml:"display_mapping"`
}

var defaultDisplayMapping = func() *displayMap {
	emptyMap := newDisplayMap()
	emptyMap.set(0, []string{""})
	return emptyMap
}()

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
		"displayMapping", cc.DisplayMapping)

	return nil
}

func (cc *DSPCanonicalConfig) populateFromMarshalled(mc *marshalledConfig) error {

	// start by loading the slider mapping because it's the only failable part for now
	if mc.DisplayMapping == nil {
		cc.logger.Warnw("Missing key in config, using default value",
			"key", "slider_mapping",
			"value", defaultDisplayMapping)

		cc.DisplayMapping = defaultDisplayMapping
	} else {

		displayMapping := newDisplayMap()

		// this is where we need to parse out each value (which is an interface{} at this point),
		// and type-assert it into either a string or a list of strings
		for key, value := range mc.DisplayMapping {
			switch typedValue := value.(type) {
			case string:
				if typedValue == "" {
					displayMapping.set(key, []string{})
				} else {
					displayMapping.set(key, []string{typedValue})
				}

			// silently ignore nil values and treat as no targets
			case nil:
				displayMapping.set(key, []string{})

			// we can't directly type-assert to a []string, so we must check each item. yup, that sucks
			case []interface{}:
				sliderItems := []string{}

				for _, listItem := range typedValue {

					// silently ignore nil values
					if listItem == nil {
						continue
					}

					listItemStr, ok := listItem.(string)
					if !ok {
						cc.logger.Warnw("Non-string value in slider mapping list",
							"key", key,
							"value", listItem,
							"valueType", fmt.Sprintf("%t", listItem))

						return fmt.Errorf("invalid slider mapping for slider %d: got type %t, need string or []string", key, typedValue)
					}

					// ignore empty strings
					if listItemStr != "" {
						sliderItems = append(sliderItems, listItemStr)
					}
				}

				displayMapping.set(key, sliderItems)
			default:
				cc.logger.Warnw("Invalid value for slider mapping key",
					"key", key,
					"value", typedValue,
					"valueType", fmt.Sprintf("%t", typedValue))

				return fmt.Errorf("invalid slider mapping for slider %d: got type %t, need string or []string", key, typedValue)
			}
		}

		cc.DisplayMapping = displayMapping
	}
	return nil
}
