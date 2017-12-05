# trailer

## a daemon tool 

### daemon
```bash
Usage of ./trailer:
  -c    Console
  -console
        Console
  -q    Stop service
  -r    Restart service
  -restart
        Restart service
  -run
        Run as service
  -s    Start service
  -start
        Start service
  -stop
        Stop service
  -v    Display version
  -version
        Display version

```

### console
```bash
Usage: add | remove | list | start | stop | restart | status
```

### config
> path services
```json
{
  "name": "demo",
  "command": ["/Users/yanggang/go/src/github.com/gimke/trailer/demo/demo"],
  "runAtLoad": true,
  "keepAlive": true
}
```