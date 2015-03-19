package commands

import (
	"fmt"
	"path/filepath"
	"regexp"

	"code.google.com/p/plotinum/plot"

	"strings"

	"github.com/onsi/cicerone/converters"
	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/cicerone/viz"
	"github.com/onsi/say"
)

func init() {
	appGuidRegExp = regexp.MustCompile(`([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)
}

type AnalyzeCreateContainer struct{}

func (f *AnalyzeCreateContainer) Usage() string {
	return "analyze-create-container GARDEN-LOG-FILE"
}

func (f *AnalyzeCreateContainer) Description() string {
	return `
Takes a single of garden log file.

Analyze-create-container then generates timeline plots and histograms
for the durations of key events.

e.g. analyze-create-container ~/garden.stdout.log
`
}

func (f *AnalyzeCreateContainer) Command(outputDir string, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("Expected a garden log file")
	}

	entriesByHandle, err := loadGardenLogFiles(args[0])
	if err != nil {
		return err
	}

	timelineDescription := TimelineDescription{
		{"Creating", MatchMessage(`garden-linux\.pool\..*\.creating`)},
		{"AcquiredPR", MatchMessage(`garden-linux\.pool\..*\.acquired-pool-resources`)},
		{"rootfs-created", MatchMessage(`garden-linux\.pool\..*\.create-rootfs\.command\.succeeded`)},
		{"create-sh-finished", MatchMessage(`garden-linux\.pool\..*\.create-script\.command\.succeeded`)},
		{"log-chain-created", MatchMessage(`garden-linux\..*\.filter\.log-chain-created`)},
		{"log-chain-conntrack-set-up", MatchMessage(`garden-linux\..*\.filter\.log-chain-conntrack-set-up`)},
		{"log-chain-setup-finished", MatchMessage(`garden-linux\..*\.filter\.log-chain-setup-finished`)},
		{"filter-setup", MatchMessage(`garden-linux\.pool\..*\.setup-filter\.finished`)},
		{"Created", MatchMessage(`garden-linux\.pool\..*\.created`)},
		{"Started", MatchMessage(`garden-linux\.pool\..*\.start\.started`)},
	}

	timelines, err := entriesByHandle.ConstructTimelines(timelineDescription)
	if err != nil {
		return err
	}
	completeTimelines := timelines.CompleteTimelines()
	say.Println(0, say.Red("Complete Timelines: %d/%d (%.2f%%)\n",
		len(completeTimelines),
		len(timelines),
		float64(len(completeTimelines))/float64(len(timelines))*100.0))

	plotCreateContainerTimelinesAndHistograms(completeTimelines, outputDir, "container-creates")

	return nil
}

func loadGardenLogFiles(file string) (*GroupedEntries, error) {
	groups := NewGroupedEntries()
	entries, err := converters.EntriesFromLagerFile(file)
	fmt.Printf("entries in log file: %d", len(entries))
	if err != nil {
		return nil, err
	}

	uniqueHandles := []string{}
	for _, entry := range entries {
		message := entry.LogEntry.Message
		if strings.Contains(message, "garden-linux.pool") && strings.Contains(message, "creating") {
			handle := getContainerHandle(message)
			if !contains(uniqueHandles, handle) {
				fmt.Println("************* Adding handle to unique set")
				uniqueHandles = append(uniqueHandles, handle)
			}
		}
	}

	for _, handle := range uniqueHandles {
		entriesForHandle := []Entry{}
		for _, entry := range entries {
			message := entry.LogEntry.Message
			if strings.Contains(message, handle) {
				fmt.Println("************* Adding entry for handle")
				entriesForHandle = append(entriesForHandle, entry)
			}
		}
		groups.AppendEntries(handle, entriesForHandle)
	}

	return groups, nil
}

//{"timestamp":"1426587234.415438175","source":"garden-linux","message":"garden-linux.pool.j21gjdqagvq.creating","log_level":1,"data":{"session":"2.1"}}
func getContainerHandle(message string) string {
	return strings.Split(message, ".")[2]
}

func contains(list []string, element string) bool {
	for _, e := range list {
		if e == element {
			return true
		}
	}
	return false
}

func plotCreateContainerTimelinesAndHistograms(timelines Timelines, outputDir string, prefix string) {
	timelines.SortByStartTime()

	histograms := viz.NewEntryPairsHistogramBoard(timelines)
	histograms.Save(3.0*float64(len(timelines.Description())), 6.0, filepath.Join(outputDir, prefix+"-histograms.png"))

	correlationBoard, _ := viz.NewCorrelationBoard(timelines)
	correlationBoard.Save(24.0, 24.0, filepath.Join(outputDir, prefix+"-correlation.png"))

	timelineBoard := &viz.Board{}
	p, _ := plot.New()
	p.Title.Text = "Timelines"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 20.0, filepath.Join(outputDir, prefix+"-timelines.png"))
}
