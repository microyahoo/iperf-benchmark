# fio-benchmark

## Building and running

### Test and Build

```
make build 
```

### Running

Running the iperf-benchmark requires `iperf` development packages to be installed on the host.

```
bin/iperf-benchmark <flags>
```

#### Usage

```
bin/iperf-benchmark -h
```

#### Flags
| Name            |  Description |
|-----------------|--------------------------------------------------------------------------------------------------|
| --output-file   | redirect iperf benchmark result to output file                                                     |
| --config-file   | iperf benchmark config file                                                                        |
| --v             | number for the log level verbosity                                                               |

### Config file
```yaml
iperf_settings:
  servers:
    - 192.168.1.1
    - 192.168.1.2
    - 192.168.1.3
  user: root
  password: password
  iperf_threads: 10
  time: 30 # second
```

