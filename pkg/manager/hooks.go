package manager

import "time"

const (
	BeforeHook = HookName("before")
	AfterHook  = HookName("after")
)

type (
	Hooks struct {
		Before []*Hook `yaml:"before,omitempty" json:"before,omitempty,optional"`
		After  []*Hook `yaml:"after,omitempty" json:"after,omitempty,optional"`
	}

	HookName string

	Hook struct {
		Statements []string      `yaml:"statements" json:"statements,omitempty,optional"`
		Wait       time.Duration `yaml:"wait,omitempty" json:"wait,omitempty,optional"`
	}
)
