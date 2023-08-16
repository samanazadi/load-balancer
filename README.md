# Go Load Balancer
A simple load balancer written in Go.

# Config Files
To run the project copy json config files from `configs` to `/etc/load-balancer/` and edit them:
- Change checker name in `config.json` to one of "tcp" or "http"
  - Change `checker.json` accordingly.
    - TCP checker doesn't need any parameters.
    - HTTP checker needs keys "path" and "keyPhrase" in the json file. 
- Change algorithm name in `config.json` based on required algorithm. At the time only round-robin algorithm is implemented ("rr").
  - Change `algorithm.json` accordingly.
    - Round-robin algorithm doesn't need any parameters.

# How to Use
Build and run `cmd/server/main.go`. Listening port, nodes and other configs will be read from config files.

# Todo
- Dockerization
- Nodes statistics
- Other algorithms: weighted round-robin, consistent hashing, the least connections
- Other types of configs: yaml, OS environment variables, command line flag
- Graceful shutdown
