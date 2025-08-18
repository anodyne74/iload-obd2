# Copilot Instructions for iload-obd2

This document provides guidelines for GitHub Copilot and human developers working on the iload-obd2 project.

## Project Structure

```
iload-obd2/
├── internal/
│   ├── transport/     # Transport layer implementations
│   ├── analysis/      # Data analysis tools
│   └── vehicle/       # Vehicle-specific implementations
├── cmd/
│   ├── analyze/       # Analysis tool commands
│   ├── query/         # Vehicle query commands
│   └── replay/        # Session replay tools
├── testing/
│   └── simulator.go   # OBD-II and CAN simulator
└── static/           # Web interface files
```

## Code Organization Guidelines

### 1. Transport Layer (internal/transport/)

When implementing new transport types:
- Implement the `Transport` interface:
  ```go
  type Transport interface {
    io.ReadWriteCloser
  }
  ```
- Add configuration in `Config` struct
- Update `NewConnection` function to handle the new type
- Include comprehensive error handling
- Add tests in `transport_test.go`

### 2. CAN Frame Handling

When working with CAN frames:
- Use the `CANFrame` struct for all frame operations
- Include timestamp information
- Handle both standard and extended frame formats
- Implement proper error handling for frame conversion
- Buffer frames appropriately to prevent data loss

Example frame handling:
```go
// Correct way to handle CAN frames
frame := CANFrame{
    ID:        uint32(rawFrame.ID),
    Data:      make([]byte, len(rawFrame.Data)),
    Timestamp: time.Now(),
}
copy(frame.Data, rawFrame.Data[:])
```

### 3. OBD-II Commands

When adding new OBD-II commands:
- Use the elmobd package's command interface
- Handle both standard and custom PIDs
- Include proper error handling
- Add retry logic for failed commands
- Document the command's purpose and expected response

Example custom command:
```go
// Custom OBD-II command template
type CustomCommand struct {
    Mode      string
    PID       string
    Name      string
    ValueFunc func(string) (interface{}, error)
}
```

### 4. WebSocket Communication

For WebSocket updates:
- Use structured JSON messages
- Include timestamp information
- Handle reconnection gracefully
- Buffer messages when appropriate
- Implement proper error handling

Message structure:
```go
type WSMessage struct {
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
    Error     string      `json:"error,omitempty"`
}
```

## Common Patterns

### 1. Error Handling

Use this pattern for error handling:
```go
if err := someFunction(); err != nil {
    return fmt.Errorf("context: %w", err)
}
```

### 2. Configuration

Use this pattern for configuration:
```go
type Config struct {
    Key      string `json:"key"`
    Required bool   `json:"required"`
    Default  string `json:"default,omitempty"`
}
```

### 3. Interface Implementation

When implementing interfaces:
```go
// Document the interface implementation
var _ Transport = (*MyTransport)(nil)

type MyTransport struct {
    // fields
}

func (t *MyTransport) Read(p []byte) (n int, err error) {
    // implementation
}
```

## Future Enhancements

### 1. Transport Layer
- [ ] Add Bluetooth transport
- [ ] Add WiFi transport
- [ ] Implement connection pooling
- [ ] Add transport metrics

### 2. Vehicle Support
- [ ] Add support for other Hyundai models
- [ ] Implement more manufacturer-specific PIDs
- [ ] Add DTC database
- [ ] Support multiple ECU queries

### 3. Analysis Tools
- [ ] Add real-time data analysis
- [ ] Implement machine learning for fault prediction
- [ ] Add performance benchmarking
- [ ] Create detailed reporting system

### 4. UI Improvements
- [ ] Add real-time graphs
- [ ] Implement customizable dashboards
- [ ] Add mobile-specific interface
- [ ] Support dark mode

## Testing Guidelines

1. Always include tests for:
   - New transport implementations
   - Custom OBD-II commands
   - CAN frame handlers
   - WebSocket message handlers

2. Test Structure:
   ```go
   func TestFeature(t *testing.T) {
       // Setup
       setup := func() {}
       
       // Teardown
       teardown := func() {}
       
       // Tests
       tests := []struct {
           name     string
           input    interface{}
           expected interface{}
           wantErr  bool
       }{
           // test cases
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // test implementation
           })
       }
   }
   ```

## Documentation Requirements

1. All new features must include:
   - Function documentation
   - Usage examples
   - Error scenarios
   - Performance considerations

2. Update these files when adding features:
   - README.md
   - API documentation
   - Configuration examples
   - Testing documentation

## Performance Considerations

1. CAN Frame Handling:
   - Use buffered channels
   - Implement frame filtering
   - Consider batch processing
   - Monitor memory usage

2. WebSocket Communication:
   - Implement message batching
   - Use compression when appropriate
   - Monitor connection health
   - Handle backpressure

3. Data Processing:
   - Use goroutines appropriately
   - Implement rate limiting
   - Consider data retention policies
   - Monitor system resources

## Security Guidelines

1. Input Validation:
   - Validate all CAN frames
   - Check message boundaries
   - Verify data integrity
   - Sanitize user input

2. Connection Security:
   - Use TLS when possible
   - Implement authentication
   - Monitor for unusual patterns
   - Log security events

## Maintenance Tasks

- Regular cleanup of old code
- Update dependencies
- Performance monitoring
- Security audits
- Documentation updates
