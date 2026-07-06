# JS Runtime Package

## Overview

The `jsruntime` package provides a QuickJS-based JavaScript runtime environment, allowing JavaScript code to be executed inside a Go application.

## Directory Structure

```
internal/jsruntime/
├── runtime.go      # Core runtime manager and main features
├── pendingjob.go   # Low-level JS_ExecutePendingJob calls (handling Promise microtasks)
└── polyfill.go     # JavaScript polyfill code
```

## Main Types

### JSEnvManager

The JS runtime environment manager, responsible for creating, managing, and destroying multiple JS runtime environments.

```go
mgr := jsruntime.NewJSEnvManager()
defer mgr.Close()
```

### JSEnv

A single JS runtime environment, containing an independent QuickJS VM instance.

```go
type JSEnv struct {
    vm       *quickjs.VM
    envID    string
    pluginID int64
    created  time.Time
    mu       sync.Mutex
    events   chan JSEventResult
}
```

### ExecuteResult

A wrapper for JS execution results.

```go
type ExecuteResult struct {
    Result string         // Execution result string
    Events []JSEventResult // Events produced during execution
}
```

### JSEventResult

A wrapper for JS event results.

```go
type JSEventResult struct {
    EnvID string // Environment ID
    Name  string // Event name
    Data  string // Event data
}
```

## Main Methods

### Creating an Environment

```go
err := mgr.CreateEnv(envID, initCode, pluginID)
```

- `envID`: Unique identifier of the environment
- `initCode`: JS code executed during initialization (optional)
- `pluginID`: ID of the plugin that created this environment

### Executing JS Code

```go
result, err := mgr.ExecuteJS(envID, code, timeoutMs)
```

- `envID`: Target environment ID
- `code`: JS code to execute
- `timeoutMs`: Timeout in milliseconds; 0 means use the default timeout (30 seconds)

### Executing JS and Waiting for Events

```go
result, err := mgr.ExecuteJSAndWaitEvents(envID, code, timeoutMs, waitEventNames)
```

- `waitEventNames`: List of event names to wait for

### Destroying an Environment

```go
// Destroy a single environment
err := mgr.DestroyEnv(envID)

// Destroy all environments created by a plugin
err := mgr.DestroyPluginEnvs(pluginID)
```

## Built-in Polyfills

The JS runtime provides a rich set of polyfills, enabling JS code to run in the QuickJS environment:

### Console

```javascript
console.log('message')
console.error('error')
console.warn('warning')
console.info('info')
console.debug('debug')
console.trace('trace')
```

### Fetch (synchronous HTTP)

```javascript
fetch(url, {
    method: 'GET',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data)
}).then(response => response.json())
```

### Timer

```javascript
setTimeout(() => {
    console.log('timer fired')
}, 1000)

clearTimeout(timerId)
```

### Buffer

```javascript
const buf = Buffer.from('hello', 'utf8')
console.log(buf.toString('base64'))
```

### Crypto

```javascript
// MD5
const hash = crypto.md5('data')

// AES encryption
const aesEncrypted = crypto.aesEncrypt(buffer, 'cbc', key, iv)

// AES decryption (string ciphertext is parsed as base64 by default; Buffer input is parsed as raw bytes)
const decrypted = crypto.aesDecrypt(aesEncrypted, 'cbc', key, iv)

// RSA encryption
const rsaEncrypted = crypto.rsaEncrypt(buffer, publicKeyPEM)

// Random bytes
const randomBytes = crypto.randomBytes(32)
```

### Zlib

```javascript
const compressed = zlib.deflate(buffer)
const decompressed = zlib.inflate(compressed)
```

### URL / URLSearchParams

```javascript
const url = new URL('https://example.com/path?query=value')
console.log(url.hostname)
console.log(url.searchParams.get('query'))

const params = new URLSearchParams('key=value&foo=bar')
console.log(params.toString())
```

### TextEncoder / TextDecoder

```javascript
const encoder = new TextEncoder()
const uint8array = encoder.encode('hello')

const decoder = new TextDecoder('utf-8')
const str = decoder.decode(uint8array)
```

## Go Bridge Functions

The JS runtime exposes system-level capabilities through the following Go bridge functions:

- `__go_send(name, data)`: Send an event to Go
- `__go_console(level, msg)`: Console log output
- `__go_fetch_async(url, method, headers, body) -> id`: True asynchronous HTTP request;
  returns an id, and the result is posted back through the asyncResults channel, with the event loop resolving the corresponding Promise
  (plugin code always calls via `globalThis.fetch()`, which provides the Promise wrapper);
  supports the internal request header `X-Fetch-No-Redirect` to disable automatic redirect following;
  supports the internal request header `X-Fetch-Timeout-Ms` to set a per-request timeout (100-30000ms, default 30000ms);
  these internal headers are not forwarded to the target server
- `__go_bridge(action, data) -> id`: True asynchronous bridge call (storage/songs/playlists/comm/jsenv)
- `__go_pop_async_result() -> json|""`: Non-blocking interface for the main event loop to drain the async result queue
- `__go_now_ms()`: Current timestamp (milliseconds)
- `__go_buffer_from(data, encoding)`: Create a Buffer
- `__go_buffer_to_string(hex, encoding)`: Convert a Buffer to a string
- `__go_crypto_md5(str)`: MD5 hash
- `__go_crypto_random_bytes(size)`: Generate random bytes
- `__go_crypto_aes_encrypt(data, mode, key, iv)`: AES encryption
- `__go_crypto_aes_decrypt(data, mode, key, iv)`: AES decryption
- `__go_crypto_rsa_encrypt(data, keyPEM)`: RSA encryption
- `__go_zlib_inflate(data)`: zlib decompression
- `__go_zlib_deflate(data)`: zlib compression

## Usage Examples

### Basic Usage

```go
package main

import (
    "log"
    "songloft/internal/jsruntime"
)

func main() {
    // Create the manager
    mgr := jsruntime.NewJSEnvManager()
    defer mgr.Close()

    // Create an environment
    err := mgr.CreateEnv("test-env", "", 0)
    if err != nil {
        log.Fatal(err)
    }

    // Execute JS code
    result, err := mgr.ExecuteJS("test-env", `
        console.log('Hello from JS!')
        return 42
    `, 0)
    
    if err != nil {
        log.Printf("execution error: %v", err)
    }
    
    log.Printf("result: %s", result.Result)
    log.Printf("event count: %d", len(result.Events))

    // Destroy the environment
    mgr.DestroyEnv("test-env")
}
```

### Waiting for Events

```go
result, err := mgr.ExecuteJSAndWaitEvents("env-id", `
    setTimeout(() => {
        __go_send('ready', 'data')
    }, 1000)
`, 5000, []string{"ready"})

if err != nil {
    log.Printf("execution error: %v", err)
}

for _, evt := range result.Events {
    log.Printf("event: %s - %s", evt.Name, evt.Data)
}
```

### Asynchronous Operations

```go
result, err := mgr.ExecuteJS("env-id", `
    fetch('https://api.example.com/data')
        .then(res => res.json())
        .then(data => {
            console.log('fetched data:', data)
            __go_send('data-received', JSON.stringify(data))
        })
`, 10000)

// Handle the returned events
for _, evt := range result.Events {
    if evt.Name == 'data-received' {
        var data map[string]interface{}
        json.Unmarshal([]byte(evt.Data), &data)
        // process the data...
    }
}
```

## Notes

1. **Thread safety**: `JSEnvManager` is thread-safe, but the VM inside each `JSEnv` is not thread-safe. The manager automatically serializes access to the same environment.

2. **Resource management**: 
   - Calling `Close()` on the manager destroys all environments
   - Call `DestroyEnv()` or `DestroyPluginEnvs()` promptly to release environments that are no longer needed
   - Avoid memory leaks

3. **Timeout control**: 
   - The default timeout is 30 seconds
   - You can customize the timeout via the `timeoutMs` parameter
   - Long-running JS code should set an appropriate timeout

4. **Event handling**: 
   - The event channel has a buffer size of 64
   - If the channel is full, new events are dropped
   - Collect and process events promptly

5. **Error handling**: 
   - JS execution errors return an error
   - You can obtain events produced during execution via `result.Events`
   - Use `ExecuteJSAndWaitEvents` to wait for specific events

## Integration with the jsplugin Package

`jsruntime` provides the low-level VM capabilities, while `internal/jsplugin` builds business logic such as plugin lifecycle, permissions, and hot updates on top of it:

```go
import "songloft/internal/jsruntime"

type Manager struct {
    jsRuntime *jsruntime.JSEnvManager
    // ...
}

func NewManager(...) *Manager {
    m := &Manager{
        jsRuntime: jsruntime.NewJSEnvManager(),
        // ...
    }
    return m
}
```

This separates the JS runtime logic from the plugin management logic, improving maintainability and testability of the code.
