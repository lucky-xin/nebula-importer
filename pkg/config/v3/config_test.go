package configv3

import (
	stderrors "errors"
	"path/filepath"

	"github.com/lucky-xin/nebula-importer/pkg/client"
	configbase "github.com/lucky-xin/nebula-importer/pkg/config/base"
	"github.com/lucky-xin/nebula-importer/pkg/source"
	specv3 "github.com/lucky-xin/nebula-importer/pkg/spec/v3"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe(".Optimize", func() {
		It("c.Sources.OptimizePathWildCard failed", func() {
			c := &Config{
				Sources: Sources{
					Source{
						Source: configbase.Source{
							Config: source.Config{
								Local: &source.LocalConfig{
									Path: "[a-b",
								},
							},
						},
					},
				},
			}
			Expect(c.Optimize(".")).To(HaveOccurred())
		})

		It("successfully", func() {
			c := &Config{
				Sources: Sources{
					Source{
						Source: configbase.Source{
							Config: source.Config{
								Local: &source.LocalConfig{
									Path: filepath.Join("testdata", "file*"),
								},
							},
						},
					},
				},
			}
			Expect(c.Optimize(".")).NotTo(HaveOccurred())
		})
	})

	Describe(".Build", func() {
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

		It("BuildLogger failed", func() {
			c.Log = &Log{
				Files: []string{filepath.Join("testdata", "not-exists", "1.log")},
			}
			Expect(c.Build()).To(HaveOccurred())
		})

		It("BuildClientPool failed", func() {
			c.Client.Version = "v"
			Expect(c.Build()).To(HaveOccurred())
		})

		It("BuildManager failed", func() {
			c.Manager.GraphName = ""
			Expect(c.Build()).To(HaveOccurred())
		})

		It("successfully", func() {
			Expect(c.Build()).NotTo(HaveOccurred())
			Expect(c.GetLogger()).NotTo(BeNil())
			Expect(c.GetClientPool()).NotTo(BeNil())
			Expect(c.GetManager()).NotTo(BeNil())
		})
	})
})

var _ = Describe("clientInitFunc", func() {
	var (
		c            Config
		ctrl         *gomock.Controller
		mockClient   *client.MockClient
		mockResponse *client.MockResponse
	)

	BeforeEach(func() {
		c.Manager.GraphName = "graphName"
		ctrl = gomock.NewController(GinkgoT())
		mockClient = client.NewMockClient(ctrl)
		mockResponse = client.NewMockResponse(ctrl)
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	It("Execute failed", func() {
		mockClient.EXPECT().Execute("USE `graphName`").Return(nil, stderrors.New("execute error"))
		Expect(c.clientInitFunc(mockClient)).To(HaveOccurred())
	})

	It("Execute IsSucceed false", func() {
		mockClient.EXPECT().Execute("USE `graphName`").Return(mockResponse, nil)
		mockResponse.EXPECT().IsSucceed().Return(false)
		mockResponse.EXPECT().GetError().Return(stderrors.New("execute error"))
		Expect(c.clientInitFunc(mockClient)).To(HaveOccurred())
	})

	It("successfully", func() {
		mockClient.EXPECT().Execute("USE `graphName`").Return(mockResponse, nil)
		mockResponse.EXPECT().IsSucceed().Return(true)
		Expect(c.clientInitFunc(mockClient)).NotTo(HaveOccurred())
	})
})
