package target

import (
	"encoding/json"
	"fmt"
)

import (
	"leaf/ir"
)

func Do(mod *ir.Module) {
	data, _ := json.Marshal(mod)
	fmt.Println(string(data))
}
