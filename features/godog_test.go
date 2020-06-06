package features

import (
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"github.com/nick-jones/gost/features/internal"
)

func TestMain(m *testing.M) {
	status := godog.RunWithOptions("godog", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format:    "pretty",
		Paths:     []string{"./"},
		Randomize: time.Now().UTC().UnixNano(),
		Tags:      "~@wip",
		Strict:    true,
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func FeatureContext(s *godog.Suite) {
	internal.NewContext().RegisterHooks(s)
}
