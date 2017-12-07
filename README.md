# trailer

## a daemon tool 

### daemon

```bash
go get -u -d github.com/gimke/trailer
cd trailer
go build
cd demo
go build
cd ..
./trailer -s
./trailer -c
```
```txt
Usage of ./trailer:
  -c	Console
  -console
    	Console
  -daemon
    	Run as service (Do not run in the terminal Run -s instead)
  -q	Stop service
  -r	Restart service
  -restart
    	Restart service
  -s	Start service
  -start
    	Start service
  -stop
    	Stop service
  -v	Display version
  -version
    	Display version
```

### console
```bash
Usage: list | start | stop | restart | status
```

### config
> config path ./services

#### json
```json
{
  "name": "ping",
  "command": ["ping","192.168.1.1"],
  "runAtLoad": true,
  "keepAlive": false
}
```
#### yaml
```yaml
name: demo
env:
  - MY_ENV=hello

command:
  - ./demo/demo

pid_file: ./demo/demo.pid
grace: false
run_at_load: true
keep_alive: false
```