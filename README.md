# trailer

## a daemon tool 

### daemon

```bash
go get -u -d github.com/gimke/trailer
cd ~/go/src/github.com/gimke/trailer
go build
cd demo
go build
cd ..
./trailer -s
./trailer -l
```
```txt
Usage of trailer:

  -l,-list          List services
                    +--------------------------------------------+
                    |  list all services ./trailer -l            |
                    +--------------------------------------------+

  -s,-start         Start service
                    +--------------------------------------------+
                    |  start normal service ./trailer -s demo    |
                    |  start daemon service ./trailer -s         |
                    +--------------------------------------------+

  -q,-stop          Stop service
                    +--------------------------------------------+
                    |  stop normal service ./trailer -q demo     |
                    |  stop daemon service ./trailer -q          |
                    +--------------------------------------------+

  -r,-restart       Restart service
                    +--------------------------------------------+
                    |  restart normal service ./trailer -r demo  |
                    |  restart daemon service ./trailer -r       |
                    +--------------------------------------------+

  -v,-version       Display version
                    +--------------------------------------------+
                    |  show trailer version ./trailer -v         |
                    +--------------------------------------------+

```

### config
> config path services

#### yaml
```yaml
name: cartdemo
env:
  - CART_MODE=release

command:
  - ./home/cartdemo/cartdemo

pid_file: ./home/cartdemo/cartdemo.pid
grace: true
run_at_load: true
keep_alive: false

deployment:
  config_headers:
    - 'Accept: application/vnd.github.VERSION.raw'
  config_path: https://api.github.com/repos/gimke/cartdemo/contents/trailer.yaml?ref=master
  version: v1.0.3
  zip: https://api.github.com/repos/gimke/cartdemo/zipball/{{version}}
```

```yaml
name: ping
env:
  - MY_ENV=hello

command:
  - ping
  - 192.168.3.1

run_at_load: true
keep_alive: false
```

```yaml
name: demo
env:
  - MY_ENV=Test

command:
  - ./home/demo/demo

pid_file: ./home/demo/demo.pid

run_at_load: true
keep_alive: false
```