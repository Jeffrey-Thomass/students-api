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


# Understanding `validator.ValidationErrors` in Go

This note explains what happens in the following Go code:

```go
if err := validator.New().Struct(student); err != nil {

	validateErrs := err.(validator.ValidationErrors)
	response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
	return
}
```

The goal is to understand:

1. What `err` contains
2. Why `err.(validator.ValidationErrors)` is needed
3. How validation errors are processed

---

# 1. What `validator.New().Struct(student)` Does

The `validator` package checks a struct against validation rules defined in struct tags.

Example struct:

```go
type Student struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"gte=18"`
}
```

Validation rules:

| Field | Rule         |
| ----- | ------------ |
| Name  | required     |
| Age   | must be ≥ 18 |

---

# 2. Running Validation

```go
err := validator.New().Struct(student)
```

Return type:

```go
error
```

Possible outcomes:

| Situation         | Value of `err`               |
| ----------------- | ---------------------------- |
| Validation passes | `nil`                        |
| Validation fails  | `validator.ValidationErrors` |

Example invalid data:

```go
student := Student{
	Name: "",
	Age: 15,
}
```

Validation errors:

* Name is required
* Age must be ≥ 18

---

# 3. What `err` Actually Contains

Even though the function returns `error`, internally the real type is:

```go
validator.ValidationErrors
```

So conceptually:

```
err (type: error interface)
        ↓
actual underlying type
validator.ValidationErrors
```

Example internal structure:

```
validator.ValidationErrors{
   {
     Field: "Name",
     Tag: "required",
     Value: ""
   },
   {
     Field: "Age",
     Tag: "gte",
     Value: 15
   }
}
```

---

# 4. Printing the Error

If you run:

```go
fmt.Println(err)
```

You might see:

```
Key: 'Student.Name' Error:Field validation for 'Name' failed on the 'required' tag
Key: 'Student.Age' Error:Field validation for 'Age' failed on the 'gte' tag
```

But this is only a **string representation** of the error.

Internally it is still structured data.

---

# 5. Why Type Assertion is Needed

Since `err` is declared as type:

```go
error
```

Go does not allow direct access to validation details.

So we use **type assertion**:

```go
validateErrs := err.(validator.ValidationErrors)
```

This converts:

```
error → validator.ValidationErrors
```

Now `validateErrs` contains the real validation error objects.

---

# 6. What `validator.ValidationErrors` Is

It is a **slice of validation errors**.

Conceptually:

```go
[]FieldError
```

Example:

```
[
  {Field: "Name", Tag: "required"},
  {Field: "Age", Tag: "gte"}
]
```

---

# 7. Accessing Error Details

Now you can loop through validation errors.

Example:

```go
for _, err := range validateErrs {
	fmt.Println(err.Field())
	fmt.Println(err.ActualTag())
}
```

Output:

```
Name
required
Age
gte
```

Useful methods:

| Method            | Description                  |
| ----------------- | ---------------------------- |
| `err.Field()`     | field that failed validation |
| `err.ActualTag()` | validation rule that failed  |
| `err.Value()`     | invalid value                |
| `err.Type()`      | field type                   |

---

# 8. How Your Code Uses It

Your code loops through errors and generates readable messages.

Example:

```go
for _, err := range errs {
	switch err.ActualTag() {
	case "required":
		errMsg = append(errMsg, fmt.Sprintf("field %s is required", err.Field()))
	default:
		errMsg = append(errMsg, fmt.Sprintf("field %s is invalid", err.Field()))
	}
}
```

Example messages generated:

```
field Name is required
field Age is invalid
```

Then they are combined:

```go
strings.Join(errMsg, ", ")
```

Result:

```
field Name is required, field Age is invalid
```

---

# 9. Final API Response

Your API sends a structured JSON response.

Example:

```json
{
  "status": "error",
  "error": "field Name is required, field Age is invalid"
}
```

HTTP Status:

```
400 Bad Request
```

---

# 10. Visual Flow

```
Client sends request
        ↓
Decode JSON into struct
        ↓
validator.Struct(student)
        ↓
Validation fails
        ↓
Error returned (type: error)
        ↓
Type assertion
err.(validator.ValidationErrors)
        ↓
Slice of validation errors
        ↓
Loop through errors
        ↓
Generate readable messages
        ↓
Send JSON response
```

---

# Key Idea

`validator.Struct()` returns a **generic `error` interface**, but when validation fails the actual type stored inside it is:

```
validator.ValidationErrors
```

Using type assertion extracts the specific validation error details so they can be processed.



# Understanding `err`, `validateErrs`, and `FieldError` in Go Validator

This note explains how validation errors work when using the Go package:

```
github.com/go-playground/validator/v10
```

Particularly how this code works:

```go
if err := validator.New().Struct(student); err != nil {

	validateErrs := err.(validator.ValidationErrors)
	response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
	return
}
```

---

# 1. What `validator.Struct()` Returns

The function:

```go
validator.New().Struct(student)
```

returns:

```
error
```

Possible outcomes:

| Case              | Result                           |
| ----------------- | -------------------------------- |
| Validation passes | `err = nil`                      |
| Validation fails  | `err` contains validation errors |

Example validation failure output when printed:

```
Key: 'Student.Email' Error:Field validation for 'Email' failed on the 'required' tag
Key: 'Student.Age' Error:Field validation for 'Age' failed on the 'required' tag
```

However, **this is only the string representation** of the error.

---

# 2. What `err` Actually Contains

Even though `err` is declared as type `error`, the **actual underlying value** stored inside it is:

```
validator.ValidationErrors
```

Conceptually:

```
err (type: error interface)
      ↓
actual underlying value
validator.ValidationErrors
```

So the error interface is just **wrapping the real value**.

---

# 3. Converting `err` Using Type Assertion

This line extracts the real value:

```go
validateErrs := err.(validator.ValidationErrors)
```

Now the type becomes:

```
validator.ValidationErrors
```

Inside the validator library it is defined roughly like this:

```go
type ValidationErrors []FieldError
```

So `validateErrs` is a **slice of validation errors**.

Conceptually:

```
validateErrs = [
  FieldError,
  FieldError,
  FieldError
]
```

Each element represents **one failed field validation**.

---

# 4. Structure of `FieldError`

Each `FieldError` object stores information about a validation failure.

Conceptually it contains:

```
FieldError
   fieldName
   validationRule
   invalidValue
```

The library does **not expose fields directly**, but provides methods to access them.

Important methods:

| Method        | Meaning                 |
| ------------- | ----------------------- |
| `Field()`     | returns field name      |
| `ActualTag()` | returns validation rule |
| `Value()`     | returns invalid value   |

---

# 5. Example Validation Failure

Struct:

```go
type Student struct {
	Name  string `validate:"required"`
	Email string `validate:"required"`
	Age   int    `validate:"required"`
}
```

Invalid input:

```json
{
  "name": "Jeffrey"
}
```

Decoded struct:

```
Name  = "Jeffrey"
Email = ""
Age   = 0
```

Validator detects:

```
Email required
Age required
```

So internally:

```
validateErrs = [
  FieldError(Email required),
  FieldError(Age required)
]
```

---

# 6. Looping Through Errors

Your code:

```go
for _, e := range validateErrs {
	fmt.Println(e.Field())
	fmt.Println(e.ActualTag())
}
```

Iteration 1:

```
e.Field()      → "Email"
e.ActualTag()  → "required"
```

Iteration 2:

```
e.Field()      → "Age"
e.ActualTag()  → "required"
```

---

# 7. Why `fmt.Println()` Shows a String

If you print:

```go
fmt.Println(validateErrs)
```

Go prints:

```
Key: 'Student.Email' Error:Field validation for 'Email' failed on the 'required' tag
Key: 'Student.Age' Error:Field validation for 'Age' failed on the 'required' tag
```

This happens because the validator library implements:

```go
Error() string
```

So when Go prints the value, it calls:

```
Error()
```

and displays the formatted message.

Important:

> This printed text is **not the actual structure** — it is just the formatted output.

---

# 8. Real Internal Structure

Internally it looks more like:

```
validateErrs
│
├── FieldError
│      Field() → "Email"
│      ActualTag() → "required"
│
└── FieldError
       Field() → "Age"
       ActualTag() → "required"
```

---

# 9. Why Type Assertion is Necessary

When `err` is type `error`, Go only allows methods defined by the `error` interface:

```
Error() string
```

So this will **not work**:

```
err.Field()
```

Because the `error` interface doesn't know about `Field()`.

After conversion:

```go
validateErrs := err.(validator.ValidationErrors)
```

Go now knows the real type, so you can access:

```
Field()
ActualTag()
Value()
```

---

# 10. Final Flow

```
Client sends request
        ↓
JSON decoded into struct
        ↓
validator.Struct(student)
        ↓
Validation fails
        ↓
err returned (type: error interface)
        ↓
Type assertion extracts real value
err.(validator.ValidationErrors)
        ↓
Slice of FieldError objects
        ↓
Loop through errors
        ↓
Generate readable messages
        ↓
Return API response
```

---

# Key Idea

The string you see when printing validation errors is **only the formatted output**.
Internally, `validator.ValidationErrors` contains **structured `FieldError` objects** that allow access to the field name and validation rule using methods like `Field()` and `ActualTag()`.
