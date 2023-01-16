# Traefik JSON-RPC filter

## Description

This plugin blocks or allows requests based on the JSON-RPC method in the request. It also supports limiting the amount of JSON-RPC requests in one batched HTTP request.

## Example configuration

```
## dynamic
http:
  middlewares:
    filter:
      plugin:
        jsonrpc-filter:
          batchedRequestLimit: 5
          allowlist:
            - "getProduct"
            - "getStore"
```
