# wgfailover
Wireguard interfaces failover handler

## Sample config file
```yaml
devices:
  - name: wg0
    health_check:
      url: https://api.ipify.org
      values:
        - 1.2.3.4
  - name: wg1
    latest_handshake_timeout: 60s
  - name: wg2
    health_check:
      url: https://api.ipify.org
      timeout: 60s
      values:
        - 5.6.7.8
        - 9.10.11.12
```