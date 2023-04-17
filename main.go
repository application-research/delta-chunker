// It creates a new Echo instance, adds some middleware, creates a new WhyPFS node, creates a new GatewayHandler, and then
// adds a route to the Echo instance
package main

import (
	"delta-chunker/cmd"
	"delta-chunker/config"
	_ "embed"
	"fmt"
	_ "net/http"
	"os"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
)

var (
	log = logging.Logger("chunker")
)

var Commit string
var Version string

// content
// content_split table

// It initializes the config, gets all the commands, and runs the app.
func main() {

	// get the config
	cfg := config.InitConfig()

	cfg.Common.Commit = Commit
	cfg.Common.Version = Version

	// get all the commands
	var commands []*cli.Command

	// cli
	commands = append(commands, cmd.CarCmd()...)
	commands = append(commands, cmd.CarChunkRunnerCmd(&cfg)...)
	commands = append(commands, cmd.HexCmd()...)

	app := &cli.App{
		Commands:    commands,
		Name:        "delta chunker / car generator",
		Description: "A file/directory chunker and car generator for Delta",
		Version:     fmt.Sprintf("%s+git.%s\n", cfg.Common.Version, cfg.Common.Commit),
		Usage:       "delta-chunker [command] [arguments]",
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
