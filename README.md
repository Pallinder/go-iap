go-iap
======

A go implementation for verifying In App Purchases via apple.

### Documentation
http://godoc.org/github.com/Pallinder/go-iap

### Usage

``` 
package main

import (
	"fmt"
	"github.com/Pallinder/go-iap"
	"log"
)

func main() {
	receipt, err := goiap.VerifyReceipt("receipt",true) // Uses the sandbox environment

	if err != nil {
	  log.Fatal(err)
	}
	
	fmt.Println("Got receipt", receipt)
}
```

Or even

```
package main

import (
	"fmt"
	"github.com/Pallinder/go-iap"
	"log"
)

func main() {
	receipt, err := goiap.VerifyReceipt("receipt",false)

	goiapErr, ok := err.(goiap.ErrorWithCode)

	if ok && goiapErr.Code() == goiap.SandboxReceiptOnProd {
		receipt, err = goiap.VerifyReceipt("receipt", true)
	}

	if err != nil {
	  log.Fatal(err)
	}

	fmt.Println("Got receipt", receipt)
}
```
