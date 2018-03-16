# S3 Handling

S3 handling package offers methods that can handle s3 requests.

## Use cases

```
// Get new s3 handler working with the nilmux.
handler := s3handling.NewHandler(nilmux)

// Register 
// Start to serve.
handler.Serve()
```