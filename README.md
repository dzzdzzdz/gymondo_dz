# Gymondo Subscription Service

My interpretation of Gymondo Subscription Service assessment

## How to Run

1. Make sure you Go (1.20+) is installed
2. Install dependencies: `go mod tidy`
3. Setup database in `config.yaml`
4. Run service using `go run cmd/main.go` 

## API Endpoints

### After running the service, check the docs out at: `http://localhost:8080/swagger/index.html`

Products
GET /products - List all products (paginated)

GET /products/:id - Get product details

Subscriptions
POST /products/:product_id/subscriptions - Create new subscription

GET /subscriptions/:id - Get subscription details

PATCH /subscriptions/:id/pause - Pause subscription (needs If-Match header)

PATCH /subscriptions/:id/unpause - Unpause subscription (needs If-Match header)

DELETE /subscriptions/:id - Cancel subscription

## Testing
To run all tests: `go test -v ./...`
To test a specific package: `go test ./pkg/[handlers|repositories]`

## Notes
* The pause/unpause function uses optimistic concurrency control with version numbers
* Subscription end dates adjust automatically when unpausing with time elapsed
* Subscriptions auto-expire
* Uses Postgres as DB but tests use in-memory SQLite

## Further Considerations Not Developed (Out of Scope)
* Implement caching layer
```
type ProductCache interface {
    Get(key string) (*models.Product, error)
    Set(key string, product *models.Product, ttl time.Duration) error
    Delete(key string) error
}
```
* Rate Limiting
```
func RateLimitMiddleware(limiter *redis.RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        if !limiter.Allow(clientIP) {
            c.JSON(http.StatusTooManyRequests, api.ErrorResponse("too many requests", "rate_limit"))
            c.Abort()
            return
        }
        c.Next()
    }
}
```
* Enhanced Error Handling (custom error types)

* Request Validation middleware
``` 
type ProductRequest struct {
    Name        string               `json:"name" binding:"required,min=3,max=100"`
    Description string               `json:"description" binding:"max=255"`
    Price       float64              `json:"price" binding:"required,gt=0"`
    Duration    models.SubscriptionDuration `json:"duration" binding:"required"`
}

func CreateProduct(c *gin.Context) {
    var req ProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        api.RespondWithError(c, api.ValidationError(err))
        return
    }
    // ...
}
```

* Metrics

* Circuit Breaking using gobreaker