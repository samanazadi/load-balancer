# Go Load Balancer
A simple load balancer written in Go.

# Config Files
- Use command line flag "-c" for configs directory. Default configs directory path is `/etc/load-balancer/`.
- Change checker name in `config.json` to one of "tcp" or "http"
  - Change `checker.json` accordingly.
    - TCP checker doesn't need any parameters.
    - HTTP checker needs keys "path" and "keyPhrase" in the json file. 
- Change algorithm name in `config.json` to one of "rr" (round-robin) or "ch" (consistent hashing)
  - Change `algorithm.json` accordingly.
    - Round-robin algorithm doesn't need any parameters.
    - Consistent hashing need two parameters: "replicas" (e.g. 100) and "hashFunc" (e.g. "crc32")
- Sample config files can be found in `configs` directory
# How to Use
Build and run `cmd/server/main.go`. Listening port, nodes and other configs will be read from config files.

# Todo
- Dockerization
- Nodes statistics
- Other types of configs: yaml, OS environment variables, command line flag
