package server

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"

	"github.com/microyahoo/iperf-benchmark/pkg/daemon/client"
)

type ServerOptions struct {
	cfgFile    string
	outputFile string
}

type ServerOption func(*ServerOptions)

func WithCfgFile(cfgFile string) ServerOption {
	return func(opts *ServerOptions) {
		opts.cfgFile = cfgFile
	}
}

func WithOutputFile(outputFile string) ServerOption {
	return func(opts *ServerOptions) {
		opts.outputFile = outputFile
	}
}

type IperfServer struct {
	cfgFile    string
	outputFile string

	ctx        context.Context
	cancelFunc context.CancelFunc

	settings *TestSettings

	lock *sync.Mutex
}

func NewIperfServer(options ...ServerOption) (*IperfServer, error) {
	opts := &ServerOptions{}
	for _, option := range options {
		option(opts)
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	s := &IperfServer{
		ctx:        ctx,
		cancelFunc: cancelFunc,
		cfgFile:    opts.cfgFile,
		outputFile: opts.outputFile,
		lock:       &sync.Mutex{},
	}
	return s, nil
}

func (s *IperfServer) Run(stopCh <-chan struct{}) (err error) {
	// TODO: check iperf
	settings, err := ParseSettings(s.cfgFile)
	if err != nil {
		return err
	}
	s.settings = settings
	err = s.doWork(settings)
	if err != nil {
		return err
	}
	// <-stopCh
	return nil
}

func (s *IperfServer) doWork(settings *TestSettings) (err error) {
	klog.Infof("fio test settings: %+v", settings.IperfSettings)

	var clients = make(map[string]*ssh.Client)
	// start iperf servers
	for _, server := range settings.IperfSettings.Servers {
		klog.Infof("start to run iperf server on %s", server)
		c, err := client.NewSSHClient(server, settings.IperfSettings.User, settings.IperfSettings.Password)
		if err != nil {
			return err
		}
		var tries = 10
		for tries > 0 {
			_, e := client.SshRemoteRunCommandWithTimeout(c, "ps cax | grep iperf", time.Minute)
			if e == nil {
				client.SshRemoteRunCommandWithTimeout(c, "pkill -x iperf", time.Minute)
			} else {
				break
			}
			tries--
			time.Sleep(2 * time.Second)
		}
		_, err = client.SshRemoteRunCommandWithTimeout(c, "iperf --server --daemon", time.Minute)
		if err != nil {
			return err
		}
		clients[server] = c
	}

	f := os.Stdout
	if s.outputFile != "" {
		f, err = os.OpenFile(s.outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			klog.Errorf("Failed to open file: %s", err)
			return err
		}
	}
	// writer := bufio.NewWriter(f)

	// run iperf test
	for _, s1 := range settings.IperfSettings.Servers {
		for _, s2 := range settings.IperfSettings.Servers {
			if s1 >= s2 {
				continue
			}
			klog.Infof("Start to run iperf between %s and %s", s1, s2)

			f.WriteString(fmt.Sprintf("=========%s, %s=============\n", s1, s2))
			f.WriteString(fmt.Sprintf("Start to run iperf between %s and %s\n", s1, s2))
			wg := &sync.WaitGroup{}
			cmd1 := fmt.Sprintf("iperf --client %s --time %d -i 1 -P %d",
				s2, settings.IperfSettings.Time, settings.IperfSettings.Threads)
			cmd2 := fmt.Sprintf("iperf --client %s --time %d -i 1 -P %d",
				s1, settings.IperfSettings.Time, settings.IperfSettings.Threads)

			runner := func(wg *sync.WaitGroup, sshClient *ssh.Client, cmd string) {
				defer wg.Done()

				output, e := client.SshRemoteRunCommandWithTimeout(sshClient, cmd, time.Hour)
				if e != nil {
					klog.Errorf("Failed to run command: %s", e)
					f.WriteString(e.Error() + "\n")
				}
				s.lock.Lock()
				f.WriteString(cmd + "\n")
				_, e = f.WriteString(output + "\n")
				s.lock.Unlock()
				if e != nil {
					klog.Errorf("Failed to write string: %s", e)
				}
			}
			wg.Add(2)
			go runner(wg, clients[s1], cmd1)
			go runner(wg, clients[s2], cmd2)
			wg.Wait()
		}
	}
	// writer.Flush()

	return nil
}

func (s *IperfServer) Close() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
}
