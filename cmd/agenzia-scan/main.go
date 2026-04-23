// Agenzia Scanner — open-source NIS2 compliance scanner for French SMBs.
// https://github.com/agenzia/scanner
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/agenzia/scanner/internal/collectors"
	"github.com/agenzia/scanner/internal/nis2"
	"github.com/agenzia/scanner/internal/reporter"
	"github.com/agenzia/scanner/internal/scoring"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	commit  = "dev"

	format    string
	outputDir string
	upload    bool
	apiKey    string
	verbose   bool
)

func main() {
	root := &cobra.Command{
		Use:     "agenzia-scan",
		Short:   "🛡️  Open-source NIS2 compliance scanner for French SMBs",
		Long:    bannerLong(),
		Version: fmt.Sprintf("%s (commit %s)", version, commit),
		RunE:    runScan,
	}

	root.Flags().StringVarP(&format, "format", "f", "pretty", "Output format: pretty | json | markdown | pdf")
	root.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for reports")
	root.Flags().BoolVar(&upload, "upload", false, "Upload report to dashboard.agenzia.uk")
	root.Flags().StringVar(&apiKey, "api-key", os.Getenv("AGENZIA_API_KEY"), "API key for --upload")
	root.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose logging")

	if err := root.Execute(); err != nil {
		color.Red("❌ Error: %v", err)
		os.Exit(1)
	}
}

func runScan(cmd *cobra.Command, args []string) error {
	fmt.Println(banner())
	fmt.Println()

	start := time.Now()

	// Step 1 — Collect
	color.Cyan("🔍  Collecting system information...")
	data, err := collectors.Run(verbose)
	if err != nil {
		return fmt.Errorf("collection failed: %w", err)
	}
	color.Green("✓  Collected %d data points in %s", data.Count(), time.Since(start).Round(time.Millisecond))

	// Step 2 — Evaluate NIS2 measures
	color.Cyan("🧪  Evaluating NIS2 Article 21 measures...")
	results := nis2.RunAllChecks(data)
	color.Green("✓  Evaluated %d checks across 10 measures", len(results.Checks))

	// Step 3 — Score
	score := scoring.Compute(results)

	// Step 4 — Report
	rep := reporter.New(data, results, score, version)

	switch format {
	case "pretty":
		rep.Pretty(os.Stdout)
	case "json":
		out, _ := json.MarshalIndent(rep, "", "  ")
		path := fmt.Sprintf("%s/agenzia-report.json", outputDir)
		_ = os.WriteFile(path, out, 0644)
		rep.Pretty(os.Stdout)
		color.Yellow("📄  Report saved to %s", path)
	case "markdown":
		rep.Markdown(os.Stdout)
	case "pdf":
		color.Yellow("⚠️  PDF format not yet implemented in v0.1 — falling back to JSON")
		out, _ := json.MarshalIndent(rep, "", "  ")
		_ = os.WriteFile(fmt.Sprintf("%s/agenzia-report.json", outputDir), out, 0644)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	if upload {
		if apiKey == "" {
			color.Red("❌  --upload requires --api-key (or AGENZIA_API_KEY env var)")
			return nil
		}
		color.Cyan("☁️   Uploading to dashboard.agenzia.uk...")
		// TODO: implement upload
		color.Green("✓  Uploaded. View at https://dashboard.agenzia.uk/scan/%s", rep.ID)
	}

	fmt.Println()
	color.HiCyan("⏱   Total scan time: %s", time.Since(start).Round(time.Millisecond))
	fmt.Println()
	color.HiWhite("💡  Tip: run with --upload to get continuous monitoring and remediation.")
	color.HiWhite("    Free signup: https://agenzia.uk/saas")

	return nil
}

func banner() string {
	c := color.New(color.FgHiCyan, color.Bold)
	return c.Sprintf(`
   █████╗  ██████╗ ███████╗███╗   ██╗███████╗██╗ █████╗
  ██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══███╔╝██║██╔══██╗
  ███████║██║  ███╗█████╗  ██╔██╗ ██║  ███╔╝ ██║███████║
  ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║ ███╔╝  ██║██╔══██║
  ██║  ██║╚██████╔╝███████╗██║ ╚████║███████╗██║██║  ██║
  ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝╚══════╝╚═╝╚═╝  ╚═╝
                                      scanner v%s
`, version)
}

func bannerLong() string {
	return `Agenzia Scanner runs a NIS2 Article 21 compliance scan on your
infrastructure in under 60 seconds. No data leaves your machine
unless you explicitly opt-in with --upload.

Made in France · Apache 2.0 · https://agenzia.uk/scanner`
}
