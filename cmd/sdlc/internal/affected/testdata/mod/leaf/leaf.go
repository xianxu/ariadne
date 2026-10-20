package leaf

import _ "embed"

//go:embed assets/banner.txt
var Banner string

func Leaf() int { return 1 }
