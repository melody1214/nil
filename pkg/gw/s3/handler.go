package s3

import (
	"fmt"
	"net/http"
)

func (s *s3APIHandlers) catchAll(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Gotya!")
}
