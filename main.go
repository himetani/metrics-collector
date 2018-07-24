package main

import (
	"fmt"
)

func main() {
	vmstatCh := genVmstat()

	for i := 0; i < 10; i++ {
		fmt.Println(<-vmstatCh)
	}

	close(vmstatCh)
}
