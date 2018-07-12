```
---
applications:
  - name: cfusers
    no-route: true
    memory: 64M
    disk: 128M
    env:
      GOPACKAGENAME: github.com/mxplusb/cfusers
      UAA_TARGET:
      UAA_USER:
      UAA_PASSWORD:
      CAPI_TARGET:
      CAPI_USER:
      CAPI_PASSWORD:
      USER_KEEPALIVE:
      DEFAULT_PASSWORD:
```