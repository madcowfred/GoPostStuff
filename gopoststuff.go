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

// Command ling flags
var verboseFlag = flag.Bool("v", false, "Show verbose debug information")
var configFlag = flag.String("c", "", "Use alternative config file")
var subjectFlag = flag.String("s", "", "Subject prefix to use")
var dirSubjectFlag = flag.Bool("d", false, "Use directory names as subject prefixes")

// Logger
var log = logging.MustGetLogger("gopoststuff")

// Config
var Config struct {
	Global struct {
		From         string
		DefaultGroup string
		ArticleSize  int32
	}

	Server map[string]*struct {
		Address     string
		Port        uint16
		Username    string
		Password    string
		Connections uint8
		TLS         bool
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

	// fmt.Printf("Verbose: %v\n", *verboseFlag)
	// fmt.Printf("Config: %v\n", *configFlag)
	// fmt.Printf("Subject: %s\n", *subjectFlag)
	// fmt.Printf("DirSubject: %v\n", *dirSubjectFlag)
	// fmt.Printf("Remaining args: %v\n", flag.Args())

	// Make sure -d or -s was specified
	if len(*subjectFlag) == 0 && !*dirSubjectFlag {
		log.Fatal("Need to specify -d or -s option, try --help!")
	}

	// Check arguments
	if len(flag.Args()) == 0 {
		log.Fatal("no filenames provided")
	}

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

	err := gcfg.ReadFileInto(&Config, cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	// Find some files

	// mc := NewMmapCache()
	// md, err := mc.MapFile("gopoststuff.go", 1)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Info("%s", md.data[0:4])

	// md.Decrement()
	// err = mc.CloseFile("gopoststuff.go")
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
