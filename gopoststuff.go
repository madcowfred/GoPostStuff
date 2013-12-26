package main

import (
    "flag"
    "fmt"
    //"github.com/madcowfred/yencode"
    "github.com/op/go-logging"
)

// Command ling flags
var verboseFlag = flag.Bool("v", false, "Show verbose debug information")
var configFlag = flag.String("c", "", "Use alternative config file")
var subjectFlag = flag.String("s", "", "Subject prefix to use")
var dirSubjectFlag = flag.Bool("d", false, "Use directory names as subject prefixes")

// Logger
var log = logging.MustGetLogger("gopoststuff")

func main() {
    // Parse command line flags
    flag.Parse()

    // fmt.Printf("Verbose: %v\n", *verboseFlag)
    // fmt.Printf("Config: %v\n", *configFlag)
    // fmt.Printf("Subject: %s\n", *subjectFlag)
    // fmt.Printf("DirSubject: %v\n", *dirSubjectFlag)
    // fmt.Printf("Remaining args: %v\n", args)

    // Set up logging
    var format = logging.MustStringFormatter("[%{level: -8s}] %{message}")
    logging.SetFormatter(format)
    if *verboseFlag {
        logging.SetLevel(logging.DEBUG, "gopoststuff")
    } else {
        logging.SetLevel(logging.INFO, "gopoststuff")
    }

    log.Debug("test debug")
    log.Info("test info")
}
