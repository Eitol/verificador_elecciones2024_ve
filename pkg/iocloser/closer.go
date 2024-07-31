package iocloser

import (
	"fmt"
	"io"
)

func CLose(closers ...io.Closer) {
	for _, c := range closers {
		if c != nil {
			err := c.Close()
			if err != nil {
				fmt.Println("error closing: ", err)
			}
		}
	}
}
