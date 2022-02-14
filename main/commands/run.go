package commands

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/vmessocket/vmessocket/common/cmdarg"
	"github.com/vmessocket/vmessocket/common/platform"
	"github.com/vmessocket/vmessocket/core"
	_ "github.com/vmessocket/vmessocket/main/all"
	"github.com/vmessocket/vmessocket/main/commands/base"
)

var CmdRun = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} run [-c config.json] [-d dir]",
	Short:       "run V2Ray with config",
	Long: `
Run V2Ray with config.

{{.Exec}} will also use the config directory specified by environment 
variable "v2ray.location.confdir". If no config found, it tries 
to load config from one of below:

	1. The default "config.json" in the current directory
	2. The config file from ENV "v2ray.location.config"
	3. The stdin if all failed above

Arguments:

	-c, -config <file>
		Config file for V2Ray. Multiple assign is accepted.

	-d, -confdir <dir>
		A directory with config files. Multiple assign is accepted.

	-r
		Load confdir recursively.

	-format <format>
		Format of config input. (default "auto")

Examples:

	{{.Exec}} {{.LongName}} -c config.json
	{{.Exec}} {{.LongName}} -d path/to/dir

Use "{{.Exec}} help format-loader" for more information about format.
	`,
	Run: executeRun,
}

var (
	configFiles cmdarg.Arg
	configDir   string
	version     = flag.Bool("version", false, "Show current version of vmessocket.")
	test        = flag.Bool("test", false, "Test config file only, without launching vmessocket server.")
	format      = flag.String("format", "json", "Format of input file.")

	_ = func() error {
		flag.Var(&configFiles, "config", "Config file for vmessocket. Multiple assign is accepted (only json). Latter ones overrides the former ones.")
		flag.Var(&configFiles, "c", "Short alias of -config")
		flag.StringVar(&configDir, "confdir", "", "A dir with multiple json config")

		return nil
	}()
)

func dirExists(file string) bool {
	if file == "" {
		return false
	}
	info, err := os.Stat(file)
	return err == nil && info.IsDir()
}

func executeRun(cmd *base.Command, args []string) {
	setConfigFlags(cmd)
	cmd.Flag.Parse(args)
	printVersion()
	configFiles = getConfigFilePath()
	server, err := startV2Ray()
	if err != nil {
		base.Fatalf("Failed to start: %s", err)
	}

	if err := server.Start(); err != nil {
		base.Fatalf("Failed to start: %s", err)
	}
	defer server.Close()

	runtime.GC()

	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
		<-osSignals
	}
}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func getConfigFilePath() cmdarg.Arg {
	if dirExists(configDir) {
		log.Println("Using confdir from arg:", configDir)
		readConfDir(configDir)
	} else if envConfDir := platform.GetConfDirPath(); dirExists(envConfDir) {
		log.Println("Using confdir from env:", envConfDir)
		readConfDir(envConfDir)
	}
	if len(configFiles) > 0 {
		return configFiles
	}
	if workingDir, err := os.Getwd(); err == nil {
		configFile := filepath.Join(workingDir, "config.json")
		if fileExists(configFile) {
			log.Println("Using default config: ", configFile)
			return cmdarg.Arg{configFile}
		}
	}
	if configFile := platform.GetConfigurationPath(); fileExists(configFile) {
		log.Println("Using config from env: ", configFile)
		return cmdarg.Arg{configFile}
	}
	log.Println("Using config from STDIN")
	return cmdarg.Arg{"stdin:"}
}

func GetConfigFormat() string {
	switch strings.ToLower(*format) {
	case "pb", "protobuf":
		return "protobuf"
	default:
		return "json"
	}
}

func printVersion() {
	version := core.VersionStatement()
	for _, s := range version {
		fmt.Println(s)
	}
}

func readConfDir(dirPath string) {
	confs, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatalln(err)
	}
	for _, f := range confs {
		if strings.HasSuffix(f.Name(), ".json") {
			configFiles.Set(path.Join(dirPath, f.Name()))
		}
	}
}

func startVmessocket() (core.Server, error) {
	configFiles := getConfigFilePath()
	config, err := core.LoadConfig(GetConfigFormat(), configFiles[0], configFiles)
	if err != nil {
		return nil, newError("failed to read config files: [", configFiles.String(), "]").Base(err)
	}
	server, err := core.New(config)
	if err != nil {
		return nil, newError("failed to create server").Base(err)
	}
	return server, nil
}
