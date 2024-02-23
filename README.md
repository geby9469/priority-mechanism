# priority-mechanism
thesis experiment

# Building the source
1. `go build` in each folder
2. create .env file
* Leader folder
   - `MY_IP` Leader node's IP
   - `HttpPort` Leader node's http port
   - `Leader_IP` Leader node's IP
   - `Leader_TCP_Port` Leader node's tcp port
   - `nodeInfoVersionPort` heartbeat port of nodes
* Miner folder
   - `MY_IP` Miner node's IP
   - `MyHTTPPort` Miner node's http port
   - `MyTCPPort` Miner node's tcp port
   - `Leader_IP` Leader node's IP for TCP
   - `Leader_Port` Leader node's port for TCP
   - `ClosestNodeIP` same as Miner node's IP
   - `ClosestNodePort` same as Miner node's tcp port
   - `LeadernodeInfoVersionPort` heartbeat port of nodes
   - 
3. Run node in each terminal

You must execute Leader node first.
