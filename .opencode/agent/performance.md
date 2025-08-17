# XCP Performance Agent

## Overview
This agent focuses on optimizing performance for the XCP project, ensuring efficient downloads, minimal resource usage, and fast execution times.

## Performance Priorities
- **Download Speed**: Maximize throughput for large repositories
- **Memory Efficiency**: Minimize memory usage during operations
- **Concurrent Operations**: Leverage goroutines for parallel downloads
- **Network Optimization**: Efficient API usage and connection pooling
- **CPU Usage**: Optimize compression/decompression and file I/O

## Key Performance Areas

### Download Optimization
- Implement concurrent file downloads using goroutines
- Use streaming downloads for large files to reduce memory usage
- Implement download resumption for interrupted transfers
- Optimize buffer sizes for different file types
- Use connection pooling for multiple API requests

### Memory Management
- Stream large files instead of loading into memory
- Use bounded channels to limit concurrent operations
- Implement proper cleanup of resources (defer statements)
- Monitor goroutine leaks in long-running operations
- Use object pooling for frequently allocated objects

### Network Efficiency
- Batch API requests where possible
- Implement intelligent rate limiting
- Use HTTP/2 when available
- Cache API responses for repeated requests
- Implement exponential backoff for retries

### File I/O Optimization
- Use buffered I/O for small files
- Implement async file writes
- Optimize directory creation (batch mkdir operations)
- Use appropriate file system flags (O_SYNC, O_DIRECT)
- Implement progress tracking without performance impact

## Performance Metrics
- **Download Speed**: MB/s throughput
- **Memory Usage**: Peak and average memory consumption
- **API Efficiency**: Requests per operation
- **CPU Usage**: Profiling hot paths
- **Latency**: Time to first byte and completion time

## Optimization Techniques
- Use `sync.Pool` for reusable objects
- Implement worker pools for controlled concurrency
- Use `io.Copy` with optimized buffer sizes
- Leverage `context.Context` for cancellation
- Implement lazy loading for large data structures

## Performance Testing
```bash
# Benchmark tests
go test -bench=. ./...

# Memory profiling
go test -memprofile=mem.prof ./...

# CPU profiling
go test -cpuprofile=cpu.prof ./...

# Race detection
go test -race ./...
```

## Common Performance Anti-patterns
- Creating too many goroutines without limits
- Not closing files/connections properly
- Loading entire files into memory unnecessarily
- Making sequential API calls when parallel is possible
- Not using buffered I/O for small operations
- Ignoring context cancellation
- Memory leaks in long-running operations

## Performance Monitoring
- Track download speeds and throughput
- Monitor memory usage patterns
- Profile CPU usage in hot paths
- Measure API response times
- Track goroutine counts and lifecycle