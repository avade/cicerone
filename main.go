package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/onsi/cicerone/commands"
	"github.com/onsi/say"
)

type Command interface {
	Usage() string
	Description() string
	Command(outputDir string, args ...string) error
}

var outputDir string
var comms []Command

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	comms = []Command{
		&commands.FezzikTasks{},
		&commands.AnalyzeCFPushes{},
		&commands.SlurpBosh{},
		&commands.AnalyzeConvergenceForMissingCells{},
		&commands.FezzikLRPs{},
		&commands.AnalyzeCreateContainer{},

		//one-offs
		// &commands.SlurpDisappearingCells{},
		// &commands.AnalyzeDisappearingCells{},
	}

	flag.StringVar(&outputDir, "output-dir", ".", "Output Directory to store plots")
	flag.Parse()
}

func main() {
	if len(flag.Args()) == 0 {
		PrintUsageAndExit()
	}

	args := flag.Args()

	for _, command := range comms {
		commandName := strings.Split(command.Usage(), " ")[0]
		if commandName == args[0] {
			err := command.Command(outputDir, args[1:]...)

			if err != nil {
				fmt.Printf("Command %s failed", commandName)
				fmt.Println(err.Error())
				os.Exit(1)
			}

			os.Exit(0)
		}
	}

	PrintUsageAndExit()
}

func PrintUsageAndExit() {
	fmt.Println("cicerone COMMAND ...")
	fmt.Println("--------------------")
	fmt.Println("Available commands:")
	for _, command := range comms {
		say.Println(1, say.Green(command.Usage()))
		say.Println(2, command.Description())

	}
	os.Exit(1)
}
