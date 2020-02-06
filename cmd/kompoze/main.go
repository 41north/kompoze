package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sync"

	k "github.com/41north/kompoze/internal"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	name        = "kompoze"
	description = "Render Docker Compose / Stack files with the power of go templates"
)

var (
	// Makefile fills this variable
	VERSION string
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.ErrorLevel)
}

func main() {
	app := &cli.App{
		Name:                   name,
		Usage:                  description,
		Version:                VERSION,
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "base-path",
				Aliases: []string{"b"},
				Usage:   "allows to set the base path to resolve relative paths",
			},
			&cli.BoolFlag{
				Name:    "no-overwrite",
				Aliases: []string{"n"},
				Usage:   "do not overwrite destination file if it already exists",
			},
			&cli.BoolFlag{
				Name:    "stdout",
				Aliases: []string{"s"},
				Usage:   "forces output to be written to stdout",
			},
			&cli.StringFlag{
				Name:    "delims",
				Aliases: []string{"D"},
				Usage:   "template tag delimiters",
				Value:   "{{:}}",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "prints debugging messages",
			},
		},
		Action: func(c *cli.Context) error {
			return executeRenderAction(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error found! Aborting execution: %s", err)
	}
}

func executeRenderAction(c *cli.Context) error {
	var (
		basePath    string
		definitions []string
		tplDelims   []string

		rawDelims   = c.String("delims")
		forceStdOut = c.Bool("stdout")
		noOverwrite = c.Bool("no-overwrite")
		debug       = c.Bool("debug")

		wg  sync.WaitGroup
		err error
	)

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	// Check base path
	if !c.IsSet("base-path") {
		if basePath, err = os.Getwd(); err != nil {
			log.Fatalf("Invalid base path: %s", err)
		}
	} else {
		if basePath, err = filepath.Abs(c.String("base-path")); err != nil {
			log.Fatalf("Invalid base path: %s", err)
		}
	}

	// Check if we are providing definitions files or resort to default one
	if c.Args().Present() {
		definitions = c.Args().Slice()
	} else {
		definitions = []string{"definition.toml"}
	}

	// Validate delims
	r, _ := regexp.Compile("^([[:ascii:]]+):([[:ascii:]]+)$")
	delims := r.FindStringSubmatch(rawDelims)
	if len(delims) != 3 {
		log.Fatalf("Bad delimiters argument: %s. Expected \"left:right\"", rawDelims)
	}
	tplDelims = []string{delims[1], delims[2]}

	// Go render
	for _, definition := range definitions {
		wg.Add(1)
		go func() {
			k.Render(definition, basePath, tplDelims, forceStdOut, noOverwrite)
			wg.Done()
		}()
	}
	wg.Wait()

	return nil
}
