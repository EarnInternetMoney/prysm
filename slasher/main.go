package main

import (
	"fmt"
	"os"
	"runtime"

	joonix "github.com/joonix/log"
	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/prysmaticlabs/prysm/shared/debug"
	"github.com/prysmaticlabs/prysm/shared/featureconfig"
	"github.com/prysmaticlabs/prysm/shared/logutil"
	"github.com/prysmaticlabs/prysm/shared/version"
	"github.com/prysmaticlabs/prysm/slasher/flags"
	"github.com/prysmaticlabs/prysm/slasher/node"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"
)

var log = logrus.WithField("prefix", "main")

func startSlasher(ctx *cli.Context) error {
	featureconfig.ConfigureSlasher(ctx)
	verbosity := ctx.String(cmd.VerbosityFlag.Name)
	level, err := logrus.ParseLevel(verbosity)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)
	slasher, err := node.NewSlasherNode(ctx)
	if err != nil {
		return err
	}
	slasher.Start()
	return nil
}

var appFlags = []cli.Flag{
	cmd.VerbosityFlag,
	cmd.DataDirFlag,
	cmd.EnableTracingFlag,
	cmd.TracingProcessNameFlag,
	cmd.TracingEndpointFlag,
	cmd.TraceSampleFractionFlag,
	flags.MonitoringPortFlag,
	cmd.LogFileName,
	cmd.LogFormat,
	cmd.ClearDB,
	cmd.ForceClearDB,
	cmd.ConfigFileFlag,
	debug.PProfFlag,
	debug.PProfAddrFlag,
	debug.PProfPortFlag,
	debug.MemProfileRateFlag,
	debug.CPUProfileFlag,
	debug.TraceFlag,
	flags.RPCPort,
	flags.KeyFlag,
	flags.RebuildSpanMapsFlag,
	flags.BeaconCertFlag,
	flags.BeaconRPCProviderFlag,
}

func init() {
	appFlags = cmd.WrapFlags(append(appFlags, featureconfig.SlasherFlags...))
}

func main() {
	app := cli.App{}
	app.Name = "hash slinging slasher"
	app.Usage = `launches an Ethereum Serenity slasher server that interacts with a beacon chain.`
	app.Version = version.GetVersion()
	app.Flags = appFlags
	app.Action = startSlasher
	app.Before = func(ctx *cli.Context) error {
		// Load any flags from file, if specified.
		if ctx.IsSet(cmd.ConfigFileFlag.Name) {
			if err := altsrc.InitInputSourceWithContext(
				appFlags,
				altsrc.NewYamlSourceFromFlagFunc(
					cmd.ConfigFileFlag.Name))(ctx); err != nil {
				return err
			}
		}

		format := ctx.String(cmd.LogFormat.Name)
		switch format {
		case "text":
			formatter := new(prefixed.TextFormatter)
			formatter.TimestampFormat = "2006-01-02 15:04:05"
			formatter.FullTimestamp = true
			// If persistent log files are written - we disable the log messages coloring because
			// the colors are ANSI codes and seen as Gibberish in the log files.
			formatter.DisableColors = ctx.String(cmd.LogFileName.Name) != ""
			logrus.SetFormatter(formatter)
			break
		case "fluentd":
			logrus.SetFormatter(joonix.NewFormatter())
			break
		case "json":
			logrus.SetFormatter(&logrus.JSONFormatter{})
			break
		default:
			return fmt.Errorf("unknown log format %s", format)
		}

		logFileName := ctx.String(cmd.LogFileName.Name)
		if logFileName != "" {
			if err := logutil.ConfigurePersistentLogging(logFileName); err != nil {
				log.WithError(err).Error("Failed to configuring logging to disk.")
			}
		}

		runtime.GOMAXPROCS(runtime.NumCPU())
		return debug.Setup(ctx)
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit(ctx)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
