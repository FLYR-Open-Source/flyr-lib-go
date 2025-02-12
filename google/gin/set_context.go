package gin

import (
	"context"

	flyrContextBase "github.com/FlyrInc/flyr-lib-go/context/base"
	flyrGoogle "github.com/FlyrInc/flyr-lib-go/google/core"
	flyrPubSub "github.com/FlyrInc/flyr-lib-go/google/pubsub"

	"github.com/gin-gonic/gin"
)

// AddPubSubToContext adds the pub/sub provider to the given context and returns it. In addition, this function also adds the provider to the given Gin engine.
// The provider can be retrieved from the context using the keys provided in the flyr-lib-go/google/core package.
// The provider is safe for concurrent use by multiple goroutines, and as such this function only needs to be called once.
// If AddPubSubToContextWithGin is unable to create the underlying Google pubsub.Client{}, it will fail and return an error.
// AddPubSubToContextWithGin also returns a Close() function handle that should be executed by the caller when finished with the pub/sub resources.
func AddPubSubToContextWithGin(ctx context.Context, ginEngine *gin.Engine, googleProjectID string) (context.Context, func() error, error) {
	ctx, closingHandler, err := flyrPubSub.AddPubSubToContext(ctx, googleProjectID)
	if err != nil {
		return ctx, nil, err
	}

	// Add provider to Gin engine
	provider, err := flyrContextBase.GetObjectFromContext[flyrGoogle.PubSubProvider](ctx, flyrContextBase.ContextKey(flyrGoogle.PubSubProviderKey))
	if err != nil {
		return ctx, nil, err
	}

	providerHandlerFunc := func(c *gin.Context) {
		c.Set(flyrGoogle.PubSubProviderKey, provider)
	}

	ginEngine.Use(providerHandlerFunc)

	return ctx, closingHandler, nil
}
