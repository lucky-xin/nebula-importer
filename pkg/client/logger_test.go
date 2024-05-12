package client

import (
	"github.com/lucky-xin/nebula-importer/pkg/logger"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("nebulaLogger", func() {
	It("newNebulaLogger", func() {
		l := newNebulaLogger(logger.NopLogger)
		l.Info("")
		l.Warn("")
		l.Error("")
		l.Fatal("")
	})
})
