package raid

import (
	"fmt"
	"log"
	"strconv"

	"github.com/zephry-zhou/baize/internal"
)

func Megacli(pci string, ctrNum int) RHController {
	megacli := `/usr/local/baize/tools/megacli64`
	if !internal.PathExists(megacli) {
		log.Println("exec file not exists:", megacli)
		return RHController{}
	}
	ret := RHController{}
	for i := 0; i <= ctrNum; i++ {
		hasPCI, _ := internal.Run.Command("sh", "-c", fmt.Sprintf("%s -AdpAllInfo -a%s -NoLog", megacli, strconv.Itoa(i)))
		if len(hasPCI) != 0 {
			ret.Cid = strconv.Itoa(i)
			break
		}
	}
	return ret
}
