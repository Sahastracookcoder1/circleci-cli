package cmd

import (
	"context"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/CircleCI-Public/circleci-cli/filetree"
	"github.com/CircleCI-Public/circleci-cli/proxy"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

const defaultConfigPath = ".circleci/config.yml"

// Path to the config.yml file to operate on.
// Used to for compatibility with `circleci config validate --path`
var configPath string

func newConfigCommand() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Operate on build config files",
	}

	collapseCommand := &cobra.Command{
		Use:   "collapse PATH",
		Short: "Collapse your CircleCI configuration to a single file",
		RunE:  collapseConfig,
		Args:  cobra.MaximumNArgs(1),
	}

	validateCommand := &cobra.Command{
		Use:     "validate PATH (use \"-\" for STDIN)",
		Aliases: []string{"check"},
		Short:   "Check that the config file is well formed.",
		RunE:    validateConfig,
		Args:    cobra.MaximumNArgs(1),
	}
	validateCommand.PersistentFlags().StringVarP(&configPath, "config", "c", ".circleci/config.yml", "path to config file")
	err := validateCommand.PersistentFlags().MarkHidden("config")
	if err != nil {
		panic(err)
	}

	expandCommand := &cobra.Command{
		Use:   "expand PATH (use \"-\" for STDIN)",
		Short: "Expand the config.",
		RunE:  expandConfig,
		Args:  cobra.ExactArgs(1),
	}

	migrateCommand := &cobra.Command{
		Use:                "migrate",
		Short:              "migrate a pre-release 2.0 config to the official release version",
		RunE:               migrateConfig,
		Hidden:             true,
		DisableFlagParsing: true,
	}
	// These flags are for documentation and not actually parsed
	migrateCommand.PersistentFlags().StringP("config", "c", ".circleci/config.yml", "path to config file")
	migrateCommand.PersistentFlags().BoolP("in-place", "i", false, "whether to update file in place.  If false, emits to stdout")

	configCmd.AddCommand(collapseCommand)
	configCmd.AddCommand(validateCommand)
	configCmd.AddCommand(expandCommand)
	configCmd.AddCommand(migrateCommand)

	return configCmd
}

// The PATH arg is actually optional, in order to support compatibility with the --path flag.
func validateConfig(cmd *cobra.Command, args []string) error {
	path := defaultConfigPath
	// First, set the path to configPath set by --path flag for compatibility
	if configPath != "" {
		path = configPath
	}

	// Then, if an arg is passed in, choose that instead
	if len(args) == 1 {
		path = args[0]
	}

	ctx := context.Background()
	response, err := api.ConfigQuery(ctx, Logger, path)

	if err != nil {
		return err
	}

	if !response.Valid {
		return response.ToError()
	}

	Logger.Infof("Config file at %s is valid", path)
	return nil
}

func expandConfig(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	response, err := api.ConfigQuery(ctx, Logger, args[0])

	if err != nil {
		return err
	}

	if !response.Valid {
		return response.ToError()
	}

	Logger.Info(response.OutputYaml)
	return nil
}

func collapseConfig(cmd *cobra.Command, args []string) error {
	tree, err := filetree.NewTree(args[0])
	if err != nil {
		return errors.Wrap(err, "An error occurred trying to build the tree")
	}

	y, err := yaml.Marshal(&tree)
	if err != nil {
		return errors.Wrap(err, "Failed trying to marshal the tree to YAML ")
	}
	Logger.Infof("%s\n", string(y))
	return nil
}

func migrateConfig(cmd *cobra.Command, args []string) error {
	return proxy.Exec([]string{"config", "migrate"}, args)
}