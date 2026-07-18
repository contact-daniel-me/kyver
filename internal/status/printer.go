package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/indexer"
	"github.com/contact-daniel-me/kyver/internal/util"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const KyverVersion = "v0.1.0"

type Environment struct {
	OS           string `json:"os"`
	GoVersion    string `json:"goVersion"`
	KyverVersion string `json:"kyverVersion"`
}

type Report struct {
	RepositoryName  string             `json:"repositoryName"`
	RepositoryRoot  string             `json:"repositoryRoot"`
	Git             GitInfo            `json:"git"`
	IndexData       *indexer.IndexData `json:"-"`
	IndexStatus     string             `json:"indexStatus"`
	IndexVersion    string             `json:"indexVersion"`
	IndexGenerated  string             `json:"indexGenerated"`
	Stats           *Statistics        `json:"stats"`
	Environment     Environment        `json:"environment"`
	Recommendations []string           `json:"recommendations"`
}

func GenerateReport(rootDir string, data *indexer.IndexData, gitInfo *GitInfo, stats *Statistics, recs []string) *Report {
	repoName := filepath.Base(rootDir)
	if data != nil {
		repoName = data.Repository
	}

	indexStatus := "Available"
	indexGen := ""
	if data == nil {
		indexStatus = "Missing"
	} else {
		indexGen = data.IndexedAt.Format("2006-01-02 15:04")
	}

	return &Report{
		RepositoryName: repoName,
		RepositoryRoot: rootDir,
		Git:            *gitInfo,
		IndexData:      data,
		IndexStatus:    indexStatus,
		IndexVersion:   "1.0",
		IndexGenerated: indexGen,
		Stats:          stats,
		Environment: Environment{
			OS:           strings.Title(runtime.GOOS),
			GoVersion:    runtime.Version(),
			KyverVersion: KyverVersion,
		},
		Recommendations: recs,
	}
}

func PrintHumanReadable(r *Report) {
	fmt.Println("Kyver Repository Status")
	fmt.Println()

	fmt.Println("Repository:")
	fmt.Printf("  Name            : %s\n", r.RepositoryName)
	fmt.Printf("  Root            : %s\n\n", r.RepositoryRoot)

	fmt.Println("Git:")
	fmt.Printf("  Branch          : %s\n", r.Git.Branch)
	fmt.Printf("  Commit          : %s\n", r.Git.Commit)
	fmt.Printf("  Working Tree    : %s\n\n", r.Git.WorkingTree)

	fmt.Println("Index:")
	fmt.Printf("  Status          : %s\n", r.IndexStatus)
	if r.IndexStatus == "Available" {
		fmt.Printf("  Version         : %s\n", r.IndexVersion)
		fmt.Printf("  Generated       : %s\n", r.IndexGenerated)
	}
	fmt.Println()

	if r.Stats != nil {
		p := message.NewPrinter(language.English)
		fmt.Println("Statistics:")
		p.Printf("  Total Files     : %d\n", r.Stats.TotalFiles)
		p.Printf("  Directories     : %d\n", r.Stats.Directories)
		fmt.Printf("  Repository Size : %s\n", util.FormatSize(r.Stats.RepositorySize))
		p.Printf("  Total LOC       : %d\n\n", r.Stats.TotalLOC)

		if len(r.Stats.Languages) > 0 {
			fmt.Println("Languages")
			fmt.Println()
			for _, lang := range r.Stats.Languages {
				p.Printf("  %-14s %4d files\n", lang.Language, lang.Count)
			}
			fmt.Println()
		}

		if len(r.Stats.LargestFiles) > 0 {
			fmt.Println("Largest Files")
			fmt.Println()
			for _, f := range r.Stats.LargestFiles {
				fmt.Printf("  %s\n", f)
			}
			fmt.Println()
		}

		if len(r.Stats.RecentFiles) > 0 {
			fmt.Println("Recent Files")
			fmt.Println()
			for _, f := range r.Stats.RecentFiles {
				fmt.Printf("  %s\n", f)
			}
			fmt.Println()
		}
	}

	fmt.Println("Environment")
	fmt.Println()
	fmt.Printf("  OS              %s\n", r.Environment.OS)
	fmt.Printf("  Go Version      %s\n", r.Environment.GoVersion)
	fmt.Printf("  Kyver Version   %s\n\n", r.Environment.KyverVersion)

	if len(r.Recommendations) > 0 {
		fmt.Println("Recommendations")
		fmt.Println()
		for _, rec := range r.Recommendations {
			fmt.Printf("• %s\n", rec)
		}
		fmt.Println()
	} else {
		fmt.Println("✓ Repository healthy.")
		fmt.Println()
	}
}

func PrintJSON(r *Report) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r)
}

func PrintSummary(r *Report) {
	sizeStr := "0 B"
	if r.Stats != nil {
		sizeStr = util.FormatSize(r.Stats.RepositorySize)
	}
	fmt.Printf("%s - %s @ %s (%d files, %s) - Index %s\n",
		r.RepositoryName, r.Git.Branch, r.Git.Commit, r.Stats.TotalFiles, sizeStr, r.IndexStatus)
}
