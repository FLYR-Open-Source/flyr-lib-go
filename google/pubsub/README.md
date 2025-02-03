# flyr-lib-go/google/pubsub

## Overview
This package contains functionality for interacting with Google Pub/Sub.

## Usage
The primary functionality of this package is provided by the `GooglePubSubProvider{}`. The provider contains a variety of methods for interacting with pub/sub topics and subscriptions.

### Mocking
The `GooglePubSubProvider{}` is abstracted by the interface defined in the `flyr_google` package:
```
type PubSubProvider interface {
	ProcessSubMessages(ctx context.Context, subscriptionName string, f func(c context.Context, m PubSubMessage)) error
	SendPubSubMessage(ctx context.Context, topicName string, message []byte, attributes map[string]string) error
}
```
This prevents tightly coupling applications to this package, and also helps mock dependencies for unit testing.

### Initializing
In order to add the provider and client for use in an application, the function `AddPubSubToContext()` should be called during application startup. (Note that this function requires the Google project ID where the topics and/or subscriptions used by the application reside.)
```
import (
	"context"

	flyrPubSub "github.com/FlyrInc/flyr-lib-go/google/pubsub"
)

func main() {
    ctx := context.Background()

    googleProjectID := ...

    ctx, pubsubClose, err := flyrPubSub.AddPubSubToContext(ctx, googleProjectID)
    if err != nil {
        // error handling
    }
    defer pubsubClose()

    // subsequent operations
}
```
Additionally, a similar mechanism exists for adding the pub/sub resources to a Gin engine. This allows for retrieval of the pub/sub resources from a Gin context.
```
import (
	"context"

	flyrPubSub "github.com/FlyrInc/flyr-lib-go/google/pubsub"

    "github.com/gin-gonic/gin"
)

func main() {
    ctx := context.Background()

    googleProjectID := ...

    ginEngine := ...

    ctx, pubsubClose, err := flyrPubSub.AddPubSubToContextWithGin(ctx, ginEngine, googleProjectID)
    if err != nil {
        // error handling
    }
    defer pubsubClose()

    // subsequent operations
}
```

### Retrieving
The `GooglePubSubProvider{}` can be retrieved from the Go context (i.e., `context.Context{}`). (Note that the relevant interfaces should be referenced as opposed to the concrete implementation types.) This is done by providing the appropriate `ContextKey{}` (from the `flyr_context` package) to the `context.Context.Value()` method. The key for the provider (`PubSubProviderKey`) is contained in the `flyr_google` package.
```
func Example() {
    // Retrieve the provider and cast it to the relevant interface
    object, ok := ctx.Value(config.ContextKey(flyr_google.PubSubProviderKey)).(flyr_google.PubSubProvider)
    if !ok {
		// error handling   
	}

    // subsequent operations
}
```
The `GetObjectFromContext()` function from the `flyr_context` package can be used to simplify object retrieval.
```
func Example() {
    // Retrieve the provider by specifying the return type and the ContextKey
    provider, err := config.GetObjectFromContext[flyr_google.PubSubProvider](ctx, config.ContextKey(flyr_google.PubSubProviderKey))
	if err != nil {
		// error handling
	}

    // subsequent operations
}
```
If using Gin, the `GetObjectFromGinContext()` function from the `flyr_context` package can be used instead.
```
func Example() {
    // Retrieve the provider by specifying the return type and the context key
    provider, err := config.GetObjectFromGinContext[flyr_google.PubSubProvider](ctx, flyr_google.PubSubProviderKey)
	if err != nil {
		// error handling
	}

    // subsequent operations
}
```

### Read from a subscription
Messages can be read from a subscription by calling the `ProcessSubMessages()` method on the `PubSubProvider{}`. This method requires passing a processing function that performs an operation on a message and then calls `message.Ack()` (to acknowledge the message) or `message.Nack()` (to refuse message acknowledgement) as appropriate based on the application logic.
```
func Example() {
    provider :=  // retrieve the provider from the context

    processingFunction := func(ctx context.Context, message PubSubMessage) {
        // do something with the message

        if (...) {
            // if processing is successful
            message.Ack()
        } else {
            // else if processing is unsuccessful
            message.Nack()
        }
    }

    err := provider.ProcessSubMessages(ctx, "my_subscription", processingFunction)
}
```

### Send a message to a topic
A message can be sent to a pub/sub topic by calling the `SendPubSubMessage()` method on the `PubSubProvider{}`.
```
func Example() {
    provider :=  // retrieve the provider from the context

    message :=  // your message content as a byte array

    attributes := // your message attributes as a map[string]string

    err := provider.SendPubSubMessage(ctx, "my_topic", message, attributes)
}
```
