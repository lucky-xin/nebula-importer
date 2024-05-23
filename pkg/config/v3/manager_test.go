package configv3

import (
	"path/filepath"

	configbase "github.com/lucky-xin/nebula-importer/pkg/config/base"
	"github.com/lucky-xin/nebula-importer/pkg/source"
	specv3 "github.com/lucky-xin/nebula-importer/pkg/spec/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	Describe(".BuildManager", func() {
		var c Config
		BeforeEach(func() {
			c = Config{
				Manager: Manager{
					GraphName: "graphName",
				},
				Sources: Sources{
					{
						Source: configbase.Source{
							Config: source.Config{
								Local: &source.LocalConfig{
									Path: filepath.Join("testdata", "file10"),
								},
							},
						},
						Nodes: specv3.Nodes{
							&specv3.Node{
								Name: "n1",
								ID: &specv3.NodeID{
									Name:  "id",
									Type:  specv3.ValueTypeString,
									Index: 0,
								},
							},
						},
					},
				},
			}
		})

		It("BuildImporters failed", func() {
			c.Manager.GraphName = ""
			Expect(c.Build()).To(HaveOccurred())
		})

		It("Importer failed", func() {
			c.Sources[0].Config.Local.Path = filepath.Join("testdata", "not-exists.csv")
			Expect(c.Build()).To(HaveOccurred())
		})

		It("successfully", func() {
			Expect(c.Build()).NotTo(HaveOccurred())
		})
	})
})
