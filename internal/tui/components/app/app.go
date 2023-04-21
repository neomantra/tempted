package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"

	temporalClient "go.temporal.io/sdk/client"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neomantra/tempted/internal/dev"
	"github.com/neomantra/tempted/internal/tui/components/header"
	"github.com/neomantra/tempted/internal/tui/components/page"
	"github.com/neomantra/tempted/internal/tui/constants"
	"github.com/neomantra/tempted/internal/tui/formatter"
	"github.com/neomantra/tempted/internal/tui/keymap"
	"github.com/neomantra/tempted/internal/tui/message"
	"github.com/neomantra/tempted/internal/tui/temporaltui"
)

type TLSConfig struct {
	CACert, CAPath, ClientCert, ClientKey, ServerName string
	SkipVerify                                        bool
}

type Config struct {
	Version, SHA        string
	HostPort, Namespace string
	HTTPAuth            string
	TLS                 TLSConfig
	CopySavePath        bool
	UpdateSeconds       time.Duration
	LogoColor           string
}

type Model struct {
	config Config
	client temporalClient.Client

	header      header.Model
	currentPage temporaltui.Page
	pageModels  map[temporaltui.Page]*page.Model

	workflowKey temporaltui.WorkflowKey

	updateID int

	width, height int
	initialized   bool
	err           error
}

func InitialModel(c Config) Model {
	firstPage := temporaltui.WorkflowsPage
	initialHeader := header.New(
		constants.LogoString,
		c.LogoColor,
		c.HostPort,
		getVersionString(c.Version, c.SHA),
		temporaltui.GetPageKeyHelp(firstPage, false, false, false, false),
	)

	return Model{
		config:      c,
		header:      initialHeader,
		currentPage: firstPage,
		updateID:    nextUpdateID(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("main %T", msg))
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	currentPageModel := m.getCurrentPageModel()
	if currentPageModel != nil && currentPageModel.EnteringInput() {
		*currentPageModel, cmd = currentPageModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case message.CleanupCompleteMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		cmd = m.handleKeyMsg(msg)
		if cmd != nil {
			return m, cmd
		}

	case message.ErrMsg:
		m.err = msg
		return m, nil

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if !m.initialized {
			err := m.initialize()
			if err != nil {
				m.err = err
				return m, nil
			}
			cmds = append(cmds, m.getCurrentPageCmd())
		} else {
			m.setPageWindowSize()
		}

	case temporaltui.PageLoadedMsg:
		if msg.Page == m.currentPage {
			m.getCurrentPageModel().SetHeader(msg.TableHeader)
			m.getCurrentPageModel().SetAllPageData(msg.AllPageRows)
			if m.currentPageLoading() {
				m.getCurrentPageModel().SetViewportXOffset(0)
			}
			m.getCurrentPageModel().SetLoading(false)

			switch m.currentPage {
			case temporaltui.WorkflowsPage:
				if m.currentPage == temporaltui.WorkflowsPage && len(msg.AllPageRows) == 0 {
					// oddly, nomad http api errors when one provides the wrong token, but returns empty results when one provides an empty token
					m.getCurrentPageModel().SetAllPageData([]page.Row{
						{"", "No job results. Is the cluster empty or no nomad token provided?"},
						{"", "Press q or ctrl+c to quit."},
					})
					m.getCurrentPageModel().SetViewportSelectionEnabled(false)
				}
			}
			cmds = append(cmds, temporaltui.UpdatePageDataWithDelay(m.updateID, m.currentPage, m.config.UpdateSeconds))
		}

	case temporaltui.UpdatePageDataMsg:
		if msg.ID == m.updateID && msg.Page == m.currentPage {
			cmds = append(cmds, m.getCurrentPageCmd())
			m.updateID = nextUpdateID()
		}

	case message.PageInputReceivedMsg:
		// if m.currentPage == temporaltui.ExecPage {
		// 	m.getCurrentPageModel().SetLoading(true)
		// 	return m, temporaltui.InitiateWebSocket(m.config.URL, m.config.Token, m.alloc.ID, m.taskName, msg.Input)
		// }
	}

	currentPageModel = m.getCurrentPageModel()
	if currentPageModel != nil && !currentPageModel.EnteringInput() {
		*currentPageModel, cmd = currentPageModel.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.updateKeyHelp()

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err) + "\n\nif this seems wrong, consider opening an issue here: https://github.com/neomantra/tempted/issues/new/choose" + "\n\nq/ctrl+c to quit"
	} else if !m.initialized {
		return ""
	}

	pageView := m.header.View() + "\n" + m.getCurrentPageModel().View()

	return pageView
}

func (m *Model) initialize() error {
	client, err := m.config.client()
	if err != nil {
		return err
	}
	m.client = *client

	m.pageModels = make(map[temporaltui.Page]*page.Model)
	for k, c := range temporaltui.GetAllPageConfigs(m.width, m.getPageHeight(), m.config.CopySavePath) {
		p := page.New(c)
		m.pageModels[k] = &p
	}

	m.initialized = true
	return nil
}

func (m *Model) cleanupCmd() tea.Cmd {
	return func() tea.Msg {
		// if m.execWebSocket != nil {
		// 	temporaltui.CloseWebSocket(m.execWebSocket)()
		// }
		return message.CleanupCompleteMsg{}
	}
}

func (m *Model) setPageWindowSize() {
	for _, pm := range m.pageModels {
		pm.SetWindowSize(m.width, m.getPageHeight())
	}
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	var cmds []tea.Cmd
	currentPageModel := m.getCurrentPageModel()

	// always exit if desired, or don't respond if typing "q" legitimately in some text input
	if key.Matches(msg, keymap.KeyMap.Exit) {
		addingQToFilter := m.currentPageFilterFocused()
		saving := m.currentPageViewportSaving()
		enteringInput := currentPageModel != nil && currentPageModel.EnteringInput()
		typingQLegitimately := msg.String() == "q" && (addingQToFilter || saving || enteringInput)
		if !typingQLegitimately || m.err != nil {
			return m.cleanupCmd()
		}
	}

	if !m.currentPageFilterFocused() && !m.currentPageViewportSaving() {
		switch {
		case key.Matches(msg, keymap.KeyMap.Forward):
			if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
				switch m.currentPage {
				case temporaltui.WorkflowsPage:
					m.workflowKey = temporaltui.WorkflowKeyFromString(selectedPageRow.Key)
				}
				nextPage := m.currentPage.Forward()
				if nextPage != m.currentPage {
					m.setPage(nextPage)
					return m.getCurrentPageCmd()
				}
			}

		case key.Matches(msg, keymap.KeyMap.Back):
			if !m.currentPageFilterApplied() {
				switch m.currentPage {
				}

				backPage := m.currentPage.Backward()
				if backPage != m.currentPage {
					m.setPage(backPage)
					cmds = append(cmds, m.getCurrentPageCmd())
					return tea.Batch(cmds...)
				}
			}

		case key.Matches(msg, keymap.KeyMap.Reload):
			if m.currentPage.DoesReload() {
				m.getCurrentPageModel().SetLoading(true)
				return m.getCurrentPageCmd()
			}
		}

		if key.Matches(msg, keymap.KeyMap.Term) {
			if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
				switch m.currentPage {
				case temporaltui.WorkflowsPage:
					m.workflowKey = temporaltui.WorkflowKeyFromString(selectedPageRow.Key)
					m.setPage(temporaltui.WorkflowTermPage)
					return m.getCurrentPageCmd()
				}
			}
		}

		//if key.Matches(msg, keymap.KeyMap.JobEvents) && m.currentPage == temporaltui.JobsPage {
		// if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
		// 	m.jobID, m.jobNamespace = temporaltui.JobIDAndNamespaceFromKey(selectedPageRow.Key)
		// 	m.setPage(temporaltui.JobEventsPage)
		// 	return m.getCurrentPageCmd()
		// }
		//}

	}

	return nil
}

func (m *Model) setPage(page temporaltui.Page) {
	m.getCurrentPageModel().HideToast()
	m.currentPage = page
	m.getCurrentPageModel().SetFilterPrefix(m.getFilterPrefix(page))
	if page.DoesLoad() {
		m.getCurrentPageModel().SetLoading(true)
	} else {
		m.getCurrentPageModel().SetLoading(false)
	}
}

func (m *Model) getCurrentPageModel() *page.Model {
	return m.pageModels[m.currentPage]
}

func (m *Model) appendToViewport(content string, startOnNewLine bool) {
	stringRows := strings.Split(content, "\n")
	var pageRows []page.Row
	for _, row := range stringRows {
		stripOS := formatter.StripOSCommandSequences(row)
		stripped := formatter.StripANSI(stripOS)
		// bell seems to mess with parent terminal
		if stripped != "\a" {
			pageRows = append(pageRows, page.Row{Row: stripped})
		}
	}
	m.getCurrentPageModel().AppendToViewport(pageRows, startOnNewLine)
	m.getCurrentPageModel().ScrollViewportToBottom()
}

func (m *Model) updateKeyHelp() {
	m.header.KeyHelp = temporaltui.GetPageKeyHelp(m.currentPage, m.currentPageFilterFocused(), m.currentPageFilterApplied(), m.currentPageViewportSaving(), m.getCurrentPageModel().EnteringInput())
}

func (m Model) getCurrentPageCmd() tea.Cmd {
	switch m.currentPage {
	case temporaltui.WorkflowsPage:
		return temporaltui.FetchWorkflowExecutions(m.client)
	case temporaltui.WorkflowDetailsPage:
		return temporaltui.FetchWorkflowDetails(m.workflowKey, m.client)
	// case temporaltui.WorkflowTermPage:
	// return temporaltui.FetchWorkflowDetails(m.workflowKey, m.client)
	default:
		panic("page load command not found")
	}
}

func (m Model) getPageHeight() int {
	return m.height - m.header.ViewHeight()
}

func (m Model) currentPageLoading() bool {
	return m.getCurrentPageModel().Loading()
}

func (m Model) currentPageFilterFocused() bool {
	return m.getCurrentPageModel().FilterFocused()
}

func (m Model) currentPageFilterApplied() bool {
	return m.getCurrentPageModel().FilterApplied()
}

func (m Model) currentPageViewportSaving() bool {
	return m.getCurrentPageModel().ViewportSaving()
}

func (m Model) getFilterPrefix(page temporaltui.Page) string {
	return page.GetFilterPrefix(m.workflowKey.WorkflowID)
}

func getVersionString(v, s string) string {
	if v == "" {
		return constants.NoVersionString
	}
	if len(s) >= 7 {
		s = s[:7]
	}

	return fmt.Sprintf("%s (%s)", v, s)
}
