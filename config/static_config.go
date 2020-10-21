package config

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/qiangyt/jog/util"
	"gopkg.in/yaml.v2"
)

// StaticConfigT ...
type StaticConfigT struct {
	// TODO: configurable
	Colorization    bool
	Replace         map[string]string
	Pattern         string
	fieldsInPattern map[string]bool
	StartupLine     StartupLine `yaml:"startup-line"`
	LineNo          Element     `yaml:"line-no"`
	UnknownLine     Element     `yaml:"unknown-line"`
	Prefix          Prefix
	Fields          FieldMap
}

// StaticConfig ...
type StaticConfig = *StaticConfigT

// UnmarshalYAML ...
func (i StaticConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return UnmarshalYAML(i, unmarshal)
}

// MarshalYAML ...
func (i StaticConfig) MarshalYAML() (interface{}, error) {
	return MarshalYAML(i)
}

// Init ...
func (i StaticConfig) Init(cfg StaticConfig) {
	if cfg != nil {
		panic(fmt.Errorf("root configure initialization"))
	}

	i.StartupLine.Init(i)
	i.LineNo.Init(i)
	i.UnknownLine.Init(i)
	i.Prefix.Init(i)
	i.Fields.Init(i)
}

// Reset ...
func (i StaticConfig) Reset() {
	i.Colorization = true
	i.Replace = make(map[string]string)
	i.Pattern = ""
	i.fieldsInPattern = make(map[string]bool)
	i.StartupLine.Reset()
	i.LineNo.Reset()
	i.UnknownLine.Reset()
	i.Prefix.Reset()
	i.Fields.Reset()
}

// HasFieldInPattern ...
func (i StaticConfig) HasFieldInPattern(fieldName string) bool {
	r, contains := i.fieldsInPattern[fieldName]
	if contains {
		return r
	}

	r = strings.Contains(i.Pattern, "${"+fieldName+"}")
	i.fieldsInPattern[fieldName] = r
	return r
}

// FromMap ...
func (i StaticConfig) FromMap(m map[string]interface{}) error {
	var v interface{}

	v = util.ExtractFromMap(m, "colorization")
	if v != nil {
		i.Colorization = util.ToBool(v)
	}

	v = util.ExtractFromMap(m, "replace")
	if v != nil {
		i.Replace = v.(map[string]string)
	}

	v = util.ExtractFromMap(m, "pattern")
	if v != nil {
		i.Pattern = v.(string)
	}

	i.fieldsInPattern = make(map[string]bool)

	v = util.ExtractFromMap(m, "startup-line")
	if v != nil {
		if err := util.UnmashalYAMLAgain(v, &i.StartupLine); err != nil {
			return err
		}
	}

	v = util.ExtractFromMap(m, "line-no")
	if v != nil {
		if err := util.UnmashalYAMLAgain(v, &i.LineNo); err != nil {
			return err
		}
	}

	v = util.ExtractFromMap(m, "unknown-line")
	if v != nil {
		if err := util.UnmashalYAMLAgain(v, &i.UnknownLine); err != nil {
			return err
		}
	}

	v = util.ExtractFromMap(m, "prefix")
	if v != nil {
		if err := util.UnmashalYAMLAgain(v, &i.Prefix); err != nil {
			return err
		}
	}

	v = util.ExtractFromMap(m, "fields")
	if v != nil {
		if err := util.UnmashalYAMLAgain(v, &i.Fields); err != nil {
			return err
		}
	}

	return nil
}

// ToMap ...
func (i StaticConfig) ToMap() map[string]interface{} {
	r := make(map[string]interface{})
	r["replace"] = i.Replace
	r["pattern"] = i.Pattern
	r["startup-line"] = i.StartupLine.ToMap()
	r["line-no"] = i.LineNo.ToMap()
	r["unknown-line"] = i.UnknownLine.ToMap()
	r["prefix"] = i.Prefix.ToMap()
	r["fields"] = i.Fields.ToMap()
	return r
}

func lookForConfigFile(dir string) string {
	log.Printf("looking for config files in: %s\n", dir)
	r := filepath.Join(dir, ".jog.yaml")
	if util.FileExists(r) {
		return r
	}
	r = filepath.Join(dir, ".jog.yml")
	if util.FileExists(r) {
		return r
	}
	return ""
}

// DetermineConfigFilePath return (file path)
func DetermineConfigFilePath() string {
	dir := util.ExeDirectory()
	r := lookForConfigFile(dir)
	if len(r) != 0 {
		return r
	}

	dir, err := homedir.Dir()
	if err != nil {
		log.Printf("failed to get home dir: %v\n", err)
		return ""
	}
	return lookForConfigFile(dir)
}

// WithDefaultYamlFile ...
func WithDefaultYamlFile() StaticConfig {
	path := DetermineConfigFilePath()

	if len(path) == 0 {
		log.Println("config file not found, take default config")
		return WithYaml(DefaultYAML)
	}

	log.Printf("config file: %s\n", path)
	return WithYamlFile(path)
}

// WithYamlFile ...
func WithYamlFile(path string) StaticConfig {
	log.Printf("config file: %s\n", path)

	yamlText := string(util.ReadFile(path))
	return WithYaml(yamlText)
}

// WithYaml ...
func WithYaml(yamlText string) StaticConfig {
	r := &StaticConfigT{
		Replace: map[string]string{
			"\\\"": "\"",
			"\\'":  "'",
			"\\\n": "\n",
			"\\\r": "",
			"\\\t": "\t",
		},
		Pattern:     "",
		StartupLine: &StartupLineT{},
		LineNo:      &ElementT{},
		UnknownLine: &ElementT{},
		Prefix:      &PrefixT{},
		Fields:      &FieldMapT{},
	}

	if err := yaml.Unmarshal([]byte(yamlText), &r); err != nil {
		panic(errors.Wrap(err, "failed to unmarshal yaml: \n"+yamlText))
	}

	r.Init(nil)

	return r
}
