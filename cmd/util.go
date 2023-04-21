package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	charmssh "github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/neomantra/tempted/internal/tui/components/app"
	"github.com/neomantra/tempted/internal/tui/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DEFAULT_TEMPORAL_ADDRESS   = "localhost:7233"
	DEFAULT_TEMPORAL_NAMESPACE = "default"
)

var (
	// Version contains the application version number. It's set via ldflags
	// in the .goreleaser.yaml file when building
	Version = ""

	// CommitSHA contains the SHA of the commit that this application was built
	// against. It's set via ldflags in the .goreleaser.yaml file when building
	CommitSHA = ""
)

func retrieveWithDefault(cmd *cobra.Command, a arg, defaultVal string) string {
	val := cmd.Flag(a.cliLong).Value.String()
	if val == "" {
		val = viper.GetString(a.cfgFileEnvVar)
	}
	if val == "" {
		return defaultVal
	}
	return val
}

func retrieveNonCLIWithDefault(a arg, defaultVal string) string {
	val := viper.GetString(a.cfgFileEnvVar)
	if val == "" {
		return defaultVal
	}
	return val
}

func retrieveAddress(cmd *cobra.Command) string {
	// TODO: validate host-port format
	return retrieveWithDefault(cmd, addrArg, DEFAULT_TEMPORAL_ADDRESS)
}

func retrieveNamespace(cmd *cobra.Command) string {
	return retrieveWithDefault(cmd, namespaceArg, DEFAULT_TEMPORAL_NAMESPACE)
}

func retrieveUpdateSeconds(cmd *cobra.Command) int {
	updateSecondsString := retrieveWithDefault(cmd, updateSecondsArg, "2")
	updateSeconds, err := strconv.Atoi(updateSecondsString)
	if err != nil {
		fmt.Println(fmt.Errorf("update value %s cannot be converted to an integer", updateSecondsString))
		os.Exit(1)
	}
	return updateSeconds
}

// customLoggingMiddleware provides basic connection logging. Connects are logged with the
// remote address, invoked command, TERM setting, window dimensions and if the
// auth was public key based. Disconnect will log the remote address and
// connection duration. It is custom because it excludes the ssh Command in the log.
func customLoggingMiddleware() wish.Middleware {
	return func(sh charmssh.Handler) charmssh.Handler {
		return func(s charmssh.Session) {
			ct := time.Now()
			hpk := s.PublicKey() != nil
			pty, _, _ := s.Pty()
			log.Printf("%s connect %s %v %v %v %v\n", s.User(), s.RemoteAddr().String(), hpk, pty.Term, pty.Window.Width, pty.Window.Height)
			sh(s)
			log.Printf("%s disconnect %s\n", s.RemoteAddr().String(), time.Since(ct))
		}
	}
}

func setup(cmd *cobra.Command, overrideToken string) (app.Model, []tea.ProgramOption) {
	temporalAddr := retrieveAddress(cmd)
	temporalNamespace := retrieveNamespace(cmd)
	updateSeconds := retrieveUpdateSeconds(cmd)
	logoColor := retrieveNonCLIWithDefault(logoColorArg, "")

	initialModel := app.InitialModel(app.Config{
		Version:       Version,
		SHA:           CommitSHA,
		HostPort:      temporalAddr,
		Namespace:     temporalNamespace,
		UpdateSeconds: time.Second * time.Duration(updateSeconds),
		LogoColor:     logoColor,
	})
	return initialModel, []tea.ProgramOption{tea.WithAltScreen()}
}

func getVersion() string {
	if Version == "" {
		return constants.NoVersionString
	}
	return Version
}
