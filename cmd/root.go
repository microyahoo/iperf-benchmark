package cmd

import (
	"flag"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	genericServer "github.com/microyahoo/iperf-benchmark/pkg/server"
)

// iperfBenchmarkOptions defines the options of iperf benchmark
type iperfBenchmarkOptions struct {
	cfgFile    string
	outputFile string
}

func newIperfBenchmarkOptions() *iperfBenchmarkOptions {
	return &iperfBenchmarkOptions{
		// TODO
	}
}

func NewIperfCommand() *cobra.Command {
	o := newIperfBenchmarkOptions()
	cmds := &cobra.Command{
		Use: "iperf-benchmark",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(o.Run(genericServer.SetupSignalHandler()))
		},
	}
	cmds.Flags().SortFlags = false
	klog.InitFlags(nil)
	// Make cobra aware of select glog flags
	// Enabling all flags causes unwanted deprecation warnings
	// from glog to always print in plugin mode
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("v"))
	// pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("logtostderr"))
	pflag.CommandLine.Set("logtostderr", "true")

	cmds.Flags().StringVar(&o.cfgFile, "config-file", "", "iperf benchmark config file, which will be ignored if job file is specified")
	cmds.Flags().StringVar(&o.outputFile, "output-file", "", "redirect iperf benchmark result to output file")

	// cmds.AddCommand(versionCmd, chartsCmd)

	return cmds
}

func (o *iperfBenchmarkOptions) Run(stopCh <-chan struct{}) error {
	klog.Info("Starting iperf benchmark")
	klog.V(4).Infof("iperf benchmark options(config-file: %s)", o.cfgFile)

	server, err := genericServer.NewIperfServer(
		genericServer.WithCfgFile(o.cfgFile),
		genericServer.WithOutputFile(o.outputFile))
	if err != nil {
		return err
	}
	genericServer.RegisterInterruptHandler(server.Close)

	err = server.Run(stopCh)
	if err != nil {
		return err
	}
	return nil
}
