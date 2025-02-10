# flyr-lib-go/context

## Overview
This package contains functions interacting with various context types.

## Usage
This package contains support for adding and retrieving objects from the base Go context and the Gin context.

### Base Go context
Two pieces of functionality are supplied in order to facilitate adding and retrieving objects from the base Go context. The `ContextKey` type extends `string` and should be used as a key for the aforementioned operations (this is recommended to avoid type collision in the base go Context). The `GetObjectFromContext()` generic function attempts to retrieve the object with the given key and cast it to the specified type.
```
import (
    "context"

    flyrContext "github.com/FlyrInc/flyr-lib-go/context/base"
)

func Example(ctx context.Context, myObject any) {
    // Add an object to the context using a ContextKey
    ctx = context.WithValue(ctx, flyrContext.ContextKey(myPackage.ObjectKey), myObject)

    // Retrieve an object previously added to the context by specifying the return type and the ContextKey
    object, err := flyrContext.GetObjectFromContext[myPackage.MyType](ctx, flyrContext.ContextKey(myPackage.ObjectKey))
    if err != nil {
        // error handling
    }

    // subsequent operations
}
```

### Gin context
One piece of functionality is supplied in order to facilitate adding and retrieving objects from the Gin context. The `GetObjectFromGinContext()` generic function attempts to retrieve the object with the given key and cast it to the specified type.
```
import (
    "context"

    flyrContext "github.com/FlyrInc/flyr-lib-go/context/gin"

    "github.com/gin-gonic/gin"
)

func MyObject() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Set(myPackage.ObjectKey, &myPackage.MyType{})
    }
}

func Example() {
    // Retrieve an object previously added to the context by specifying the return type and the context key
    object, err := flyrContext.GetObjectFromGinContext[myPackage.MyType](ctx, myPackage.ObjectKey)
    if err != nil {
        // error handling
    }

    // subsequent operations
}
```
