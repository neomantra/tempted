package temporaltui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	//"github.com/hashicorp/nomad/api"
	"strings"
	"time"

	"github.com/neomantra/tempted/internal/tui/components/page"
	"github.com/neomantra/tempted/internal/tui/components/viewport"
	"github.com/neomantra/tempted/internal/tui/constants"
	"github.com/neomantra/tempted/internal/tui/keymap"
	"github.com/neomantra/tempted/internal/tui/style"
)

type Page int8

const (
	Unset Page = iota
	WorkflowsPage
	WorkflowDetailsPage
	WorkflowTermPage
)

func GetAllPageConfigs(width, height int, copySavePath bool) map[Page]page.Config {
	return map[Page]page.Config{
		WorkflowsPage: {
			Width: width, Height: height,
			FilterPrefix: "Jobs", LoadingString: WorkflowsPage.LoadingString(),
			CopySavePath: copySavePath, SelectionEnabled: true, WrapText: false, RequestInput: false,
			ViewportConditionalStyle: constants.JobsViewportConditionalStyle,
		},
		WorkflowDetailsPage: {
			Width: width, Height: height,
			LoadingString: WorkflowDetailsPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
	}
}

func (p Page) DoesLoad() bool {
	noLoadPages := []Page{}
	for _, noLoadPage := range noLoadPages {
		if noLoadPage == p {
			return false
		}
	}
	return true
}

func (p Page) DoesReload() bool {
	noReloadPages := []Page{}
	for _, noReloadPage := range noReloadPages {
		if noReloadPage == p {
			return false
		}
	}
	return true
}

func (p Page) doesUpdate() bool {
	noUpdatePages := []Page{
		// LoglinePage,     // doesn't load
		// ExecPage,        // doesn't reload
		// LogsPage,        // currently makes scrolling impossible - solve in https://github.com/robinovitch61/wander/issues/1
		// JobSpecPage,     // would require changes to make scrolling possible
		// AllocSpecPage,   // would require changes to make scrolling possible
		// JobEventsPage,   // constant connection, streams data
		// JobEventPage,    // doesn't load
		// AllocEventsPage, // constant connection, streams data
		// AllocEventPage,  // doesn't load
		// AllEventsPage,   // constant connection, streams data
		// AllEventPage,    // doesn't load
	}
	for _, noUpdatePage := range noUpdatePages {
		if noUpdatePage == p {
			return false
		}
	}
	return true
}

func (p Page) String() string {
	switch p {
	case Unset:
		return "undefined"
	case WorkflowsPage:
		return "workflows"
	case WorkflowDetailsPage:
		return "workflow details"
	case WorkflowTermPage:
		return "workflow termination"
	}
	return "unknown"
}

func (p Page) LoadingString() string {
	return fmt.Sprintf("Loading %s...", p.String())
}

func (p Page) Forward() Page {
	switch p {
	case WorkflowsPage:
		return WorkflowDetailsPage
	}
	return p
}

func (p Page) Backward() Page {
	switch p {
	case WorkflowDetailsPage:
		return WorkflowDetailsPage
	}
	return p
}

func (p Page) GetFilterPrefix(workflowID string) string {
	switch p {
	case WorkflowsPage:
		return "Workflow"
	case WorkflowDetailsPage:
		return fmt.Sprintf("Workflow Details for %s", style.Bold.Render(workflowID))
	case WorkflowTermPage:
		return fmt.Sprintf("Workflow Termination for %s", style.Bold.Render(workflowID))
	default:
		panic("page not found")
	}
}

type PageLoadedMsg struct {
	Page        Page
	TableHeader []string
	AllPageRows []page.Row
}

type UpdatePageDataMsg struct {
	ID   int
	Page Page
}

func UpdatePageDataWithDelay(id int, p Page, d time.Duration) tea.Cmd {
	if p.doesUpdate() && d > 0 {
		return tea.Tick(d, func(t time.Time) tea.Msg { return UpdatePageDataMsg{id, p} })
	}
	return nil
}

func getShortHelp(bindings []key.Binding) string {
	var output string
	for _, km := range bindings {
		output += style.KeyHelpKey.Render(km.Help().Key) + " " + style.KeyHelpDescription.Render(km.Help().Desc) + "    "
	}
	output = strings.TrimSpace(output)
	return output
}

func changeKeyHelp(k *key.Binding, h string) {
	k.SetHelp(k.Help().Key, h)
}

func GetPageKeyHelp(currentPage Page, filterFocused, filterApplied, saving, enteringInput bool) string {
	firstRow := []key.Binding{keymap.KeyMap.Exit}

	if currentPage.DoesReload() && !saving && !filterFocused {
		firstRow = append(firstRow, keymap.KeyMap.Reload)
	}

	viewportKeyMap := viewport.GetKeyMap()
	secondRow := []key.Binding{viewportKeyMap.Save, keymap.KeyMap.Wrap}
	thirdRow := []key.Binding{viewportKeyMap.Down, viewportKeyMap.Up, viewportKeyMap.PageDown, viewportKeyMap.PageUp, viewportKeyMap.Bottom, viewportKeyMap.Top}

	var fourthRow []key.Binding
	if nextPage := currentPage.Forward(); nextPage != currentPage {
		changeKeyHelp(&keymap.KeyMap.Forward, currentPage.Forward().String())
		fourthRow = append(fourthRow, keymap.KeyMap.Forward)
	}

	if filterApplied {
		changeKeyHelp(&keymap.KeyMap.Back, "remove filter")
		fourthRow = append(fourthRow, keymap.KeyMap.Back)
	} else if prevPage := currentPage.Backward(); prevPage != currentPage {
		changeKeyHelp(&keymap.KeyMap.Back, fmt.Sprintf("%s", currentPage.Backward().String()))
		fourthRow = append(fourthRow, keymap.KeyMap.Back)
	}

	if currentPage == WorkflowsPage {
		fourthRow = append(fourthRow, keymap.KeyMap.Term)
	}

	if saving {
		changeKeyHelp(&keymap.KeyMap.Forward, "confirm save")
		changeKeyHelp(&keymap.KeyMap.Back, "cancel save")
		secondRow = []key.Binding{keymap.KeyMap.Back, keymap.KeyMap.Forward}
		return getShortHelp(firstRow) + "\n" + getShortHelp(secondRow)
	}

	if filterFocused {
		changeKeyHelp(&keymap.KeyMap.Forward, "apply filter")
		changeKeyHelp(&keymap.KeyMap.Back, "cancel filter")
		secondRow = []key.Binding{keymap.KeyMap.Back, keymap.KeyMap.Forward}
		return getShortHelp(firstRow) + "\n" + getShortHelp(secondRow)
	}

	var final string
	for _, row := range [][]key.Binding{firstRow, secondRow, thirdRow, fourthRow} {
		final += getShortHelp(row) + "\n"
	}

	return strings.TrimRight(final, "\n")
}
