package healthcheck

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen, color.Bold)
	red    = color.New(color.FgRed, color.Bold)
	yellow = color.New(color.FgYellow, color.Bold)
	cyan   = color.New(color.FgCyan, color.Bold)
	dim    = color.New(color.Faint)
	bold   = color.New(color.Bold)
	white  = color.New(color.FgWhite)
)

var spinner = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type ReceiverResult struct {
	Name    string
	Checks  []PermissionCheck
	Err     error
	Skipped bool
}

func RenderResultsStream(pending []string, skipped []string, resultCh <-chan ReceiverResult) bool {
	fmt.Println()
	cyan.Print("  Receiver Health Checks")
	fmt.Println()
	dim.Println("  " + strings.Repeat("═", 50))

	var healthy, partial, down int
	remaining := make(map[string]bool, len(pending))
	for _, name := range pending {
		remaining[name] = true
	}

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	spinIdx := 0
	done := 0
	total := len(pending)

	showSpinner := func() {
		var names []string
		for _, name := range pending {
			if remaining[name] {
				names = append(names, name)
			}
		}
		label := strings.Join(names, ", ")
		if len(label) > 40 {
			label = label[:37] + "..."
		}
		fmt.Printf("\r\033[2K")
		dim.Printf("  %s Checking: %s", spinner[spinIdx%len(spinner)], label)
		spinIdx++
	}

	showSpinner()

	for done < total {
		select {
		case r := <-resultCh:
			fmt.Printf("\r\033[2K")
			delete(remaining, r.Name)

			fmt.Println()
			if r.Err != nil {
				renderDown(r)
				down++
			} else {
				allOk := true
				for _, c := range r.Checks {
					if !c.Granted {
						allOk = false
						break
					}
				}
				if allOk {
					renderHealthy(r)
					healthy++
				} else {
					renderPartial(r)
					partial++
				}
			}

			done++
			if done < total {
				showSpinner()
			}

		case <-ticker.C:
			if done < total {
				showSpinner()
			}
		}
	}

	fmt.Printf("\r\033[2K")

	renderSkipped(skipped)
	renderSummary(healthy, partial, down, len(skipped))

	return partial > 0 || down > 0
}

func RenderResults(results []ReceiverResult) {
	var healthy, partial, down int
	var monitored []ReceiverResult
	var skippedNames []string

	for _, r := range results {
		if r.Skipped {
			skippedNames = append(skippedNames, r.Name)
			continue
		}
		monitored = append(monitored, r)
	}

	fmt.Println()
	cyan.Print("  Receiver Health Checks")
	fmt.Println()
	dim.Println("  " + strings.Repeat("═", 50))

	for _, r := range monitored {
		fmt.Println()
		if r.Err != nil {
			renderDown(r)
			down++
		} else {
			allOk := true
			for _, c := range r.Checks {
				if !c.Granted {
					allOk = false
					break
				}
			}
			if allOk {
				renderHealthy(r)
				healthy++
			} else {
				renderPartial(r)
				partial++
			}
		}
	}

	renderSkipped(skippedNames)
	renderSummary(healthy, partial, down, len(skippedNames))
}

func renderHealthy(r ReceiverResult) {
	green.Print("  ● ")
	bold.Print(r.Name)
	padStatus(r.Name, "HEALTHY", green)
	renderChecks(r.Checks)
}

func renderPartial(r ReceiverResult) {
	yellow.Print("  ● ")
	bold.Print(r.Name)
	padStatus(r.Name, "PARTIAL", yellow)
	renderChecks(r.Checks)
}

func renderDown(r ReceiverResult) {
	red.Print("  ● ")
	bold.Print(r.Name)
	padStatus(r.Name, "DOWN", red)
	dim.Printf("    └─ %s\n", trimPrefix(r.Err.Error()))
}

func renderChecks(checks []PermissionCheck) {
	for i, c := range checks {
		connector := "├─"
		if i == len(checks)-1 {
			connector = "└─"
		}
		if c.Granted {
			dim.Printf("    %s ", connector)
			green.Print("✓ ")
			white.Printf("%s\n", c.Name)
		} else {
			dim.Printf("    %s ", connector)
			red.Print("✗ ")
			white.Print(c.Name)
			if c.Err != nil {
				dim.Printf(" — %s", trimPrefix(c.Err.Error()))
			}
			fmt.Println()
		}
	}
}

func renderSkipped(names []string) {
	if len(names) == 0 {
		return
	}

	sort.Strings(names)

	fmt.Println()
	dim.Println("  ── Health check not implemented " + strings.Repeat("─", 19))

	line := "  "
	for i, name := range names {
		addition := name
		if i < len(names)-1 {
			addition += " · "
		}
		if len(line)+len(addition) > 60 {
			dim.Println(line)
			line = "  "
		}
		line += addition
	}
	if len(line) > 2 {
		dim.Println(line)
	}
}

func renderSummary(healthy, partial, down, skipped int) {
	fmt.Println()
	dim.Print("  Summary: ")
	green.Printf("%d healthy", healthy)
	dim.Print(" · ")
	yellow.Printf("%d partial", partial)
	dim.Print(" · ")
	red.Printf("%d down", down)
	dim.Print(" · ")
	dim.Printf("%d skipped", skipped)
	fmt.Println()
	fmt.Println()
}

func padStatus(name, status string, c *color.Color) {
	padding := 48 - len(name) - len(status)
	if padding < 2 {
		padding = 2
	}
	fmt.Print(strings.Repeat(" ", padding))
	c.Println(status)
}

func trimPrefix(s string) string {
	prefixes := []string{"postgresql: ", "mongodb: ", "mysql: "}
	for _, p := range prefixes {
		s = strings.TrimPrefix(s, p)
	}
	return s
}
