package features

import (
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"github.com/nick-jones/gost/features/internal"
)

func TestMain(m *testing.M) {
	status := godog.TestSuite{
		Name:                "gost",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:    "pretty",
			Paths:     []string{"./"},
			Randomize: time.Now().UTC().UnixNano(),
			Tags:      "~@wip",
			Strict:    true,
		},
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func InitializeScenario(sc *godog.ScenarioContext) {
	internal.NewContext().RegisterHooks(sc)
}
