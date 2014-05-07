package main

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"
	//"fmt"
	"code.google.com/p/gcfg"
	"github.com/op/go-logging"
)

const (
	GPS_VERSION = "0.2.0-git"
)

// Command ling flags
var verboseFlag = flag.Bool("v", false, "Show verbose debug information")
var configFlag = flag.String("c", "", "Use alternative config file")
var groupFlag = flag.String("g", "", "Newsgroup(s) to post to - separate multiple with a comma \",\"")
var subjectFlag = flag.String("s", "", "Subject to use")
var dirSubjectFlag = flag.Bool("d", false, "Use directory names as subjects")

// Logger
var log = logging.MustGetLogger("gopoststuff")

// Config
var Config struct {
	Global struct {
		From         string
		DefaultGroup string
		SubjectPrefix string
		ArticleSize  int64
		ChunkSize    int64
	}

	Server map[string]*struct {
		Address     string
		Port        int
		Username    string
		Password    string
		Connections int
		TLS         bool
		InsecureSSL bool
	}
}

func main() {
	// Parse command line flags
	flag.Parse()

	// Set up logging
	var format = logging.MustStringFormatter(" %{level: -8s}  %{message}")
	logging.SetFormatter(format)
	if *verboseFlag {
		logging.SetLevel(logging.DEBUG, "gopoststuff")
	} else {
		logging.SetLevel(logging.INFO, "gopoststuff")
	}

	log.Info("gopoststuff starting...")

	// Make sure -d or -s was specified
	if len(*subjectFlag) == 0 && !*dirSubjectFlag {
		log.Fatal("Need to specify -d or -s option, try gopoststuff --help")
	}

	// Check arguments
	if len(flag.Args()) == 0 {
		log.Fatal("No filenames provided")
	}

	// Check that all supplied arguments exist
	for _, arg := range flag.Args() {
		st, err := os.Stat(arg)
		if err != nil {
			log.Fatalf("stat %s: %s", arg, err)
		}

		// If -d was specified, make sure that it's a directory
		if *dirSubjectFlag && !st.IsDir() {
			log.Fatalf("-d option used but not a directory: %s", arg)
		}
	}

	// Load config file
	var cfgFile string
	if len(*configFlag) > 0 {
		cfgFile = *configFlag
	} else {
		// Default to user homedir for config file
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		cfgFile = filepath.Join(u.HomeDir, ".gopoststuff.conf")
	}

	log.Debug("Reading config from %s", cfgFile)

	err := gcfg.ReadFileInto(&Config, cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	// Fix default values
	if (Config.Global.ChunkSize == 0) {
		Config.Global.ChunkSize = 10240
	}

	// Start the magical spawner
	Spawner(flag.Args())
}
