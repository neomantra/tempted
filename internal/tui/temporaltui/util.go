package temporaltui

import (
	"errors"
	"io/ioutil"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neomantra/tempted/internal/tui/components/page"
	"github.com/neomantra/tempted/internal/tui/formatter"
)

func doQuery(url, token string, params [][2]string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Nomad-Token", token)

	query := req.URL.Query()
	for _, p := range params {
		query.Add(p[0], p[1])
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func get(url, token string, params [][2]string) ([]byte, error) {
	resp, err := doQuery(url, token, params)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if string(body) == "ACL token not found" {
		return nil, errors.New("token not authorized")
	}
	return body, nil
}

func PrettifyLine(l string, p Page) tea.Cmd {
	return func() tea.Msg {
		// nothing async actually happens here, but this fits the PageLoadedMsg pattern
		pretty := formatter.PrettyJsonStringAsLines(l)

		var rows []page.Row
		for _, row := range pretty {
			rows = append(rows, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{
			Page:        p,
			TableHeader: []string{},
			AllPageRows: rows,
		}
	}
}
