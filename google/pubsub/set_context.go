package flyr_pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	flyrContext "github.com/FlyrInc/flyr-lib-go/context"
	flyrGoogle "github.com/FlyrInc/flyr-lib-go/google"
	"github.com/gin-gonic/gin"
)

// AddPubSubToContext adds the pub/sub provider and client to the given context and returns it.
// These objects can be retrieved from the context using the keys provided in the integrations_google package.
// These objects are safe for concurrent use by multiple goroutines, and as such this function only needs to be called once.
// If AddPubSubToContext is unable to create the underlying Google pubsub.Client{}, it will fail and return an error.
// AddPubSubToContext also returns a Close() function handle that should be executed by the caller when finished with the pub/sub resources.
func AddPubSubToContext(ctx context.Context, googleProjectID string) (context.Context, func() error, error) {
	// Set the pub/sub client
	pubsubClient, err := pubsub.NewClient(ctx, googleProjectID)
	if err != nil {
		return ctx, nil, err
	}

	client := &clientWrapper{client: pubsubClient}
	closingHandler := func() error { return client.Close() }

	// Create provider with client
	provider := NewGooglePubSubProvider(client)

	// Set the pub/sub provider
	ctx = context.WithValue(ctx, flyrContext.ContextKey(flyrGoogle.PubSubProviderKey), provider)

	return ctx, closingHandler, nil
}

// AddPubSubToContext adds the pub/sub provider and client to the given context and returns it. In addition, this function also adds the provider to the given Gin engine.
// These objects can be retrieved from the context using the keys provided in the integrations_google package.
// These objects are safe for concurrent use by multiple goroutines, and as such this function only needs to be called once.
// If AddPubSubToContextWithGin is unable to create the underlying Google pubsub.Client{}, it will fail and return an error.
// AddPubSubToContextWithGin also returns a Close() function handle that should be executed by the caller when finished with the pub/sub resources.
func AddPubSubToContextWithGin(ctx context.Context, ginEngine *gin.Engine, googleProjectID string) (context.Context, func() error, error) {
	ctx, closingHandler, err := AddPubSubToContext(ctx, googleProjectID)
	if err != nil {
		return ctx, nil, err
	}

	// Add provider to Gin engine
	provider, err := flyrContext.GetObjectFromContext[flyrGoogle.PubSubProvider](ctx, flyrContext.ContextKey(flyrGoogle.PubSubProviderKey))
	if err != nil {
		return ctx, nil, err
	}

	providerHandlerFunc := func(c *gin.Context) {
		c.Set(flyrGoogle.PubSubProviderKey, provider)
	}

	ginEngine.Use(providerHandlerFunc)

	return ctx, closingHandler, nil
}
