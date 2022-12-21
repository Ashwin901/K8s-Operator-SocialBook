package main

import (
	"fmt"

	"github.com/ashwin901/social-book-operator/pkg/apis/operators/v1alpha1"
)

func main() {
	sb := v1alpha1.SocialBook{}
	fmt.Println(sb)
}
