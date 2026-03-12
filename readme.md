# Graceful Shutdown in Go HTTP Server

## What is Graceful Shutdown?

Graceful shutdown means **stopping the server without interrupting active requests**.

Instead of immediately killing the server, it:

1. Stops accepting new requests
2. Allows ongoing requests to finish
3. Closes the server safely

This prevents problems like:

* Lost responses
* Interrupted database operations
* Partial transactions

---

# Basic Graceful Shutdown Flow

```
Server running
      ↓
User presses Ctrl + C / SIGTERM received
      ↓
Stop accepting new requests
      ↓
Wait for existing requests to complete
      ↓
Shutdown server
```

---

# Example Implementation

```go
done := make(chan os.Signal, 1)

signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

go func() {
    err := server.ListenAndServe()
    if err != nil {
        panic(err)
    }
}()

<-done

slog.Info("shutting down server")

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := server.Shutdown(ctx)
if err != nil {
    slog.Error("failed to shutdown server", slog.String("error", err.Error()))
}

slog.Info("server shut down successfully")
```

---

# Step-by-Step Explanation

## 1. Listen for OS signals

```go
done := make(chan os.Signal, 1)
signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
```

This listens for system signals like:

* `Ctrl + C`
* `SIGTERM`
* `SIGINT`

When the signal is received, it is sent to the `done` channel.

---

## 2. Run the server in a goroutine

```go
go func() {
    err := server.ListenAndServe()
}()
```

The server runs in a **separate goroutine** so the main program can wait for shutdown signals.

---

## 3. Wait for shutdown signal

```go
<-done
```

This blocks the program until a signal is received.

---

## 4. Create a timeout context

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

This creates a **5-second timeout** for the shutdown process.

Meaning:

```
Allow active requests to finish within 5 seconds
```

---

## 5. Shutdown the server

```go
server.Shutdown(ctx)
```

This performs graceful shutdown:

* Stops accepting new requests
* Waits for active requests to finish
* Exits the server

If requests take longer than the timeout, they are forcefully stopped.

---

# Why Context is Used

Context helps control long-running operations.

In shutdown:

```
Context → defines how long the server should wait before stopping.
```

Example:

```
5 second timeout
```

---

# Visual Flow

```
Ctrl + C
   ↓
Signal received
   ↓
Program unblocks from <-done
   ↓
Create timeout context (5s)
   ↓
server.Shutdown(ctx)
   ↓
Existing requests finish
   ↓
Server stops safely
```

---

# Key Benefits

Graceful shutdown prevents:

* Broken client responses
* Unfinished HTTP requests
* Corrupted database writes
* Abrupt server termination

---

# Simple Summary

Graceful shutdown in Go allows a server to **stop safely by finishing active requests before exiting**, using **signals and a context timeout**.
