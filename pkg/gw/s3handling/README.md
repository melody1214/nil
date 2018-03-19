# S3 Handling

S3 handling package offers methods that can handle s3 requests.

## Use cases

```
// Get the new s3 handler.
handler := s3handling.NewHandler()

// Handler registered to nil multiplexer.
handler.RegisteredTo(nilmux)
```