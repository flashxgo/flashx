# FlashX

FlashX provides a Go package that helps easily setup a reverse proxy for your server/s.

It supports the following features:
- Reverse Proxy
- Reverse Proxy based on custom logic
- Load Balancing
  - Round Robin
  - Weighted Round Robin
  - Least Connections
  - TODO: Weighted Least Connections
  - TODO: Weighted Response Time
- Blacklist IPs
- Rate Limiting (requests per second)
- Modify Request and Response
- Buffer Pool
- Custom Error Handler
- Custom Logger
- Flush Interval
