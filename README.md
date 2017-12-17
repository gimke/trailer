# trailer

## a daemon tool 

### daemon

```bash
go get -u -d github.com/gimke/trailer
cd ~/go/src/github.com/gimke/trailer
go build
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
#name: service name
#env:
#  - CART_MODE=release

#command:
#  - ./home/cartdemo/cartdemo

#pid_file: ./home/cartdemo/cartdemo.pid
#grace: true
#run_at_load: false
#keep_alive: false

#deploy:
#  provider: github (only support github gitlab)
#  token: Personal access tokens (visit https://github.com/settings/tokens or https://gitlab.com/profile/personal_access_tokens and generate a new token)
#  repository: repository address (https://github.com/gimke/cartdemo)
#  version: branchName (e.g master), latest release (e.g latest）or a release described in a file (e.g master:filepath/version.txt)
#  payload: payload url when update success

name: demo

command:
  - ping
  - -c
  - 3
  - 192.168.1.1

run_at_load: true
keep_alive: false
```