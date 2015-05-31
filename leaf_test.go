package leaf

import (
	"fmt"
	"leaf/scanner"
	"os"
	"strconv"
	"testing"
)

func TestScanner(*testing.T) {
	for i := 0; i <= 4; i++ {
		if f, err := os.Open("test" + strconv.Itoa(i) + ".leaf"); err == nil {
			s := scanner.New()
			s.Init(f)
			for {
				sym := s.Get()
				fmt.Println(sym)
				if s.Error() != nil || s.Eot() {
					break
				}
			}
		}
	}
}
