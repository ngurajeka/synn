package main

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/ngurajeka/synn/command"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig()
}

var ui cli.Ui

func main() {
	stdout := viper.GetStringSlice("cli.stdout")
	if len(stdout) == 0 {
		stdout = []string{"stdout"}
	}
	stderr := viper.GetStringSlice("cli.stderr")
	if len(stderr) == 0 {
		stderr = []string{"stderr"}
	}
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = stdout
	zapConfig.ErrorOutputPaths = stderr
	logger, _ := zapConfig.Build()
	defer logger.Sync()

	ui = &cli.BasicUi{Writer: os.Stdout}

	commands := &cli.CLI{
		Args: os.Args[1:],
		Commands: map[string]cli.CommandFactory{
			"generate": func() (cli.Command, error) {
				return command.NewGenerateCmd(ui, logger), nil
			},
		},
		HelpFunc: cli.BasicHelpFunc("synn"),
		Version:  "1.0.0",
	}

	exitCode, err := commands.Run()
	if err != nil {
		logger.Error("error executing accent",
			zap.Strings("args", os.Args),
			zap.Error(err),
		)
		os.Exit(1)
	}

	os.Exit(exitCode)
}
