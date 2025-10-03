# Error logging

Event-Driven architecture helps you build resilient systems, but without the right patterns, it's easy to overcomplicate things.
In the next several exercises, we'll look at a few error-handling patterns.

Middleware functions are a great place to keep the error-handling logic. 
There are two ways you can capture errors in middleware.

The first one is to store the return values in variables and return them.

```go
func HandleErrors(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		msgs, err := next(msg)
		
		if err != nil {
			// Handle the error 
		}
		
		return msgs, err
	}
}
```

The second one is to use `defer` and named returns. This is a different flavor of the same thing.

```go
func HandleErrors(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) (msgs []*message.Message, err error) {
		defer func() {
			if err != nil { 
				// Handle the error
			}
		}()

		return next(msg)
	}
}
```

No matter when in the sequence the middleware is added,
the error handling is done at the end, after the handler and all other middleware functions are executed.
Previously, we used middleware that executed before the handler.
This pattern is a way to **run some code after the handler returns.**

{{tip}}

Don't forget to return the produced messages (`msgs`). Otherwise, you will lose them!

Similarly, be mindful if you return the `err` or `nil` from the function.
You could acknowledge the message by mistake.

{{endtip}}

## Exercise

Exercise path: ./project

**Extend the logging middleware to log errors.**

The log message should be:

```text
Error while handling a message
```

It should include two log fields: `error` with the error and `message_id` with the message UUID.

{{tip}}

You can add multiple keys to the logger like this:

```go
logger.With(
	"key1", value1,
	"key2", value2,
).Info("Log message")
```

{{endtip}}
