package agent

import (
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/dapr/dapr/dapr-agent/pkg/bootstrap"
	"github.com/dapr/dapr/dapr-agent/pkg/model"
	"github.com/gogo/protobuf/types"
)

const (
	// epochFileTemplate is a template for the root config JSON
	epochFileTemplate = "daprd-rev%d.json"
)

type daprd struct {
	Metadata *model.NodeMetadata
}

// NewDaprd creates an instance of the proxy control commands
func NewDaprd(cfg *model.NodeMetadata) Proxy {
	// inject tracing flag for higher levels
	//var args []string
	//if cfg.LogLevel != "" {
	//	args = append(args, "-l", cfg.LogLevel)
	//}
	//if cfg.ComponentLogLevel != "" {
	//	args = append(args, "--component-log-level", cfg.ComponentLogLevel)
	//}
	//cfg.ExtraArgs = args
	return &daprd{
		Metadata: cfg,
	}
}

func (e *daprd) IsLive() bool {

	return true
}

func (e *daprd) Drain() error {
	return nil
}

func (e *daprd) args(fname string, epoch int) []string {
	startupArgs := []string{
		"--mode", e.Metadata.Mode,
		"--dapr-http-port", e.Metadata.DaprHTTPPort,
		"--dapr-grpc-port", e.Metadata.DaprAPIGRPCPort,
		"--dapr-internal-grpc-port", e.Metadata.DaprInternalGRPCPort,
		"--app-port", e.Metadata.AppPort,
		"--app-id", e.Metadata.AppID,
		"--control-plane-address", e.Metadata.ControlPlaneAddress,
		"--app-protocol", e.Metadata.AppProtocol,
		"--placement-host-address", e.Metadata.PlacementServiceHostAddr,
		"--config", "",
		//"--log-level", e.Metadata.LogLevel,
		"--app-max-concurrency", strconv.Itoa(e.Metadata.AppMaxConcurrency),
		"--sentry-address", e.Metadata.SentryAddress,
	}
	if e.Metadata.EnableMesh {
		startupArgs = append(startupArgs, "--enable-mesh")
	}

	startupArgs = append(startupArgs, e.Metadata.ExtraArgs...)

	return startupArgs
}

func (e *daprd) Run(config interface{}, epoch int, abort <-chan error) error {
	var fname string
	// Note: the cert checking still works, the generated file is updated if certs are changed.
	// We just don't save the generated file, but use a custom one instead. Pilot will keep
	// monitoring the certs and restart if the content of the certs changes.

	out, err := bootstrap.New(bootstrap.BootStrapConfig{
		Metadata: e.Metadata,
	}).CreateFileForEpoch(epoch)

	if err != nil {
		log.Error("Failed to generate bootstrap config: ", err)
		os.Exit(1) // Prevent infinite loop attempting to write the file, let k8s/systemd report
	}
	fname = out

	// spin up a new dapr runtime process
	args := e.args(fname, epoch)
	log.Infof("dapr runtime command: %v", args)

	/* #nosec */
	cmd := exec.Command(e.Metadata.BinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-abort:
		log.Warnf("Aborting epoch %d", epoch)
		if errKill := cmd.Process.Kill(); errKill != nil {
			log.Warnf("killing epoch %d caused an error %v", epoch, errKill)
		}
		return err
	case err := <-done:
		return err
	}
}

func (e *daprd) Cleanup(epoch int) {

	filePath := model.ConfigFile(e.Metadata.ConfigPath, epoch)
	cfg, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Warnf("Failed to open config file %s for %d, %v", filePath, epoch, err)
	}else {
		log.Infof("dapr runtime cleaned, file:%s", string(cfg))
	}
	if err := os.Remove(filePath); err != nil {
		log.Warnf("Failed to delete config file %s for %d, %v", filePath, epoch, err)
	}
}

// convertDuration converts to golang duration and logs errors
func convertDuration(d *types.Duration) time.Duration {
	if d == nil {
		return 0
	}
	dur, err := types.DurationFromProto(d)
	if err != nil {
		log.Warnf("error converting duration %#v, using 0: %v", d, err)
	}
	return dur
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
