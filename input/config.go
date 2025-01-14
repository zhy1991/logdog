package input

import (
	"github.com/zhjx922/alert/output"
)

type Inputs struct {
	Name          string   `yaml:"name"`
	ScanFrequency int64    `yaml:"scan_frequency"`
	AlertCount    int      `yaml:"alert_count"` // 新增 alert_count 字段
	Paths         []string `yaml:"paths"`
	IncludeLines  []string `yaml:"include_lines"`
	ExcludeLines  []string `yaml:"exclude_lines"`
}

type Config struct {
	Inputs     []*Inputs    `yaml:"inputs"`
	OutputHttp *output.Http `yaml:"output.http"`
}
