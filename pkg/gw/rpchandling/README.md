# RPC Handling

RPC handling package offers methods that can handle rpc requests.

## Use cases

```
// Get new rpc handler working with the nilmux.
handler := rpchandling.NewHandler(nilmux)

// Handler do proxying connection to other node.
handler.Proxying(conn net.Conn)
```
