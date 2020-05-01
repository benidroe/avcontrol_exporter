# avcontrol_exporter
AV control system state exporter for Prometheus

This exporter collects engaged system state data from av control systems. This enables simple state monitoring of av systems, e.g. from extron or crestron.
The only requirement is to push UDP messages with state codes from the system to this exporters engage port.

## Usage

```sh
./avcontrol_exporter
```

Visit http://localhost:2113/control?target=devicename.localnetwork where devicename.localnetwork is the DNS-Name of the control system.

## Enable avcontrol on your control system
All you have to do is sending a simple UDP message to port 2114.

Here is a list of messages available. Place it  to the appropriate place in your control logic.
Every message must be terminated with \n

```
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
video.input.select.[target]=0	# no input
video.input.select.[target]=1	# vga
video.input.select.[target]=2	# hdmi1
video.input.select.[target]=3	# hdmi2
video.input.select.[target]=4	# display port
video.input.select.[target]=5	# sdi1
video.input.select.[target]=6	# sdi2
video.input.select.[target]=7	# fbas1
video.input.select.[target]=8	# fbas2
video.input.select.[target]=9	# doc cam
video.input.select.[target]=10	# local computer
...
video.input.select.[target]=n	# something else, specified for your institution
## It's recommended to set up your custom number->input convention, suitable for your application.
## [target] is the name of the target device (e.g. a part of the dns-name, like projector-2) where your input selection is displayed. In regard to video matrixes, multible targets can be specified.

```
