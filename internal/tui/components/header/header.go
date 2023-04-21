package header

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/neomantra/tempted/internal/tui/style"
)

type Model struct {
	logo, logoColor, nomadUrl, version, KeyHelp string
}

func New(logo string, logoColor string, nomadUrl, version, keyHelp string) (m Model) {
	return Model{logo: logo, logoColor: logoColor, nomadUrl: nomadUrl, version: version, KeyHelp: keyHelp}
}

func (m Model) View() string {
	logoStyle := style.Logo
	if m.logoColor != "" {
		logoStyle.Foreground(lipgloss.Color(m.logoColor))
	}
	logo := logoStyle.Render(m.logo)
	clusterUrl := style.ClusterUrl.Render(m.nomadUrl)
	left := style.Header.Render(lipgloss.JoinVertical(lipgloss.Center, logo, m.version, clusterUrl))
	styledKeyHelp := style.KeyHelp.Render(m.KeyHelp)
	return lipgloss.JoinHorizontal(lipgloss.Center, left, styledKeyHelp)
}

func (m Model) ViewHeight() int {
	return len(strings.Split(m.View(), "\n"))
}
