# avcontrol_exporter
AV control system state exporter for Prometheus

This exporter collects engaged system state data from av control systems. This enables simple state monitoring of av systems, e.g. from extron or crestron.
The only requirement is to push UDP messages with state codes from the system to this exporters engage port.

The engaged data is cached within a redis database.

## Usage

```sh
./avcontrol_exporter
```

Regulary, you should specify your redis host,  password, and database.
```sh
./avcontrol_exporter --redis.address=redishost:6379 --redis.password=mycanarypass --redis.db=0
```

Visit http://localhost:2113/control?target=devicename.localnetwork where devicename.localnetwork is the DNS-Name of the control system.

## Enable avcontrol on your control system
All you have to do is sending a simple UDP message to port 2114.

Here is a list of messages available. Place it  to the appropriate place in your control logic.
Every message must be terminated with \n

```
### Driver internal services:
# keepalive (send every 15sec.)
system.keepalive = 1

# system power state
system.power.state=0	# system is off
system.power.state=1	# system is on
system.power.state=2    # system is booting
system.power.state=3	# system is in shutdown

# system initialized
system.init=1           # system is initialized. This sets a timestamp in the database and calculates the uptime

# nightly shutdown
system.power.nightly=1  # system is running a nightly shutdown. This sets a timestamp in the database. Result will be 1 for 5 minutes

# emergency shutdown from fire detection unit
system.firealarm.state=1 # ok
system.firealarm.state=2 # alarm
system.firealarm.state=3 # emergency shutdown active

# touchpanel page select
system.touchpanel.page=0 # off
system.touchpanel.page=1 # presentation
system.touchpanel.page=2 # video
system.touchpanel.page=3 # audio
system.touchpanel.page=4 # lights
...
system.touchpanel.page=n # the proposed number -> value mapping is only an example. It's recomendet to set up your own convention for your institution.

# connection state of peripherieal devices
system.connected.[device]=0 # device is not connected
system.connected.[device]=1 # device is connected
## [device] can be the dns-name of the device.
## It's recommended to set a monitior for every device checking it's connection state in your control logic.

# video input selection
video.input.select.[target]=0 # no input
video.input.select.[target]=1 # vga 1
video.input.select.[target]=2 # vga 2
video.input.select.[target]=3 # vga 3
video.input.select.[target]=4 # vga 4
video.input.select.[target]=5 # vga 5
video.input.select.[target]=6 # hdmi 1
video.input.select.[target]=7 # hdmi 2
video.input.select.[target]=8 # hdmi 3
video.input.select.[target]=9 # hdmi 4
video.input.select.[target]=10 # hdmi 5
video.input.select.[target]=11 # Digital 1
video.input.select.[target]=12 # Digital 2
video.input.select.[target]=13 # Digital 3
video.input.select.[target]=14 # Digital 4
video.input.select.[target]=15 # Digital 5
video.input.select.[target]=16 # SDI 1
video.input.select.[target]=17 # SDI 2
video.input.select.[target]=18 # SDI 3
video.input.select.[target]=19 # SDI 4
video.input.select.[target]=20 # SDI 5
video.input.select.[target]=21 # TV-Receiver
video.input.select.[target]=22 # IPTV
video.input.select.[target]=23 # Doc Cam
video.input.select.[target]=24 # PTZ CAM
video.input.select.[target]=25 # PC
video.input.select.[target]=26 # auto
## [target] ist a String. E.g. the Display-Name oder the dns-name or the IP.
## [target] is the name of the target device (e.g. a part of the dns-name, like projector-2) where your input selection is displayed. In regard to video matrixes, multible targets can be specified.

```

### Prometheus Config
```
 - job_name: 'avcontrol'
    static_configs:
      - targets:
          - ´lecturehall-1-controlprocessor.mycanarynetwork.tld
    metrics_path: /control
    params:
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: 127.0.0.1:2113  # The avcontrol exporter's real hostname:port.

```