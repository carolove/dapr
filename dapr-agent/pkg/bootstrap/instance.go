package bootstrap

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
	"text/template"

	"github.com/dapr/dapr/dapr-agent/pkg/model"
)

const (
	// EpochFileTemplate is a template for the root config JSON
	EpochFileTemplate = "daprd-rev%d.%s"
	DefaultCfgDir     = "/var/lib/dapr/runtime/grpc_bootstrap.json"
)

// Instance of a configured dapr runtime bootstrap writer.
type Instance interface {
	// WriteTo writes the content of the dapr runtime bootstrap to the given writer.
	WriteTo(templateFile string, w io.Writer) error

	// CreateFileForEpoch generates an dapr runtime bootstrap file for a particular epoch.
	CreateFileForEpoch(epoch int) (string, error)
}

// New creates a new Instance of an dapr runtime bootstrap writer.
func New(cfg BootStrapConfig) Instance {
	return &instance{
		BootStrapConfig: cfg,
	}
}

type instance struct {
	BootStrapConfig
}

func (i *instance) WriteTo(templateFile string, w io.Writer) error {
	// Get the input bootstrap template.
	t, err := newTemplate(templateFile)
	if err != nil {
		return err
	}

	// Create the parameters for the template.
	templateParams, err := i.toTemplateParams()
	if err != nil {
		return err
	}

	// Execute the template.
	return t.Execute(w, templateParams)
}

func toJSON(i interface{}) string {
	if i == nil {
		return "{}"
	}

	ba, err := json.Marshal(i)
	if err != nil {
		log.Warnf("Unable to marshal %v: %v", i, err)
		return "{}"
	}

	return string(ba)
}

func (i *instance) CreateFileForEpoch(epoch int) (string, error) {
	// Create the output file.
	if err := os.MkdirAll(i.Metadata.ConfigPath, 0700); err != nil {
		return "", err
	}

	templateFile := getEffectiveTemplatePath(i.Metadata)

	outputFilePath := model.ConfigFile(i.Metadata.ConfigPath, epoch)
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = outputFile.Close() }()

	// Write the content of the file.
	if err := i.WriteTo(templateFile, outputFile); err != nil {
		return "", err
	}

	return outputFilePath, err
}

// getEffectiveTemplatePath gets the template file that should be used for bootstrap
func getEffectiveTemplatePath(pc *model.NodeMetadata) string {
	var templateFilePath string
	switch {
	case pc.CustomConfigFile != "":
		templateFilePath = pc.CustomConfigFile
	default:
		templateFilePath = DefaultCfgDir
	}
	return templateFilePath
}

func newTemplate(templateFilePath string) (*template.Template, error) {
	cfgTmpl, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		return nil, err
	}

	funcMap := template.FuncMap{
		"toJSON": toJSON,
	}
	return template.New("bootstrap").Funcs(funcMap).Parse(string(cfgTmpl))
}

// isIPv6Proxy check the addresses slice and returns true for a valid IPv6 address
// for all other cases it returns false
func isIPv6Proxy(ipAddrs []string) bool {
	for i := 0; i < len(ipAddrs); i++ {
		addr := net.ParseIP(ipAddrs[i])
		if addr == nil {
			// Should not happen, invalid IP in proxy's IPAddresses slice should have been caught earlier,
			// skip it to prevent a panic.
			continue
		}
		if addr.To4() != nil {
			return false
		}
	}
	return true
}
