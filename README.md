# trailer

## a daemon tool 

### daemon
```bash
Usage of ./trailer:
  -c	Console
  -console
    	Console
  -daemon
    	Run as service
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
> path services
```json
{
  "name": "demo",
  "command": ["/home/demo/demo"],
  "runAtLoad": true,
  "keepAlive": false
}
```
```yaml
name: ping2
command:
  - ping
  - 192.168.2.1

run_at_load: true
keep_alive: false
```