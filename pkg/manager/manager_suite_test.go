package manager

import (
	"github.com/lucky-xin/nebula-importer/pkg/spec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestManager(t *testing.T) {
	var faileds []spec.Record
	var succeededs []spec.Record
	println(len(faileds))
	println(len(succeededs))
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg manager Suite")
}
