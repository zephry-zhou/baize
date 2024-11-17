package storage

import (
	"baize/internal/utils"
	"fmt"
	"log"
	"strconv"
)

func Megacli(pci string, ctrNum int) RHController {
	megacli := `/usr/local/baize/tools/megacli64`
	if !utils.PathExists(megacli) {
		log.Println("exec file not exists:", megacli)
		return RHController{}
	}
	ret := RHController{}
	for i := 0; i <= ctrNum; i++ {
		hasPCI, _ := run.Command("sh", "-c", fmt.Sprintf("%s -AdpAllInfo -a%s -NoLog", megacli, strconv.Itoa(i)))
		if len(hasPCI) != 0 {
			ret.Cid = strconv.Itoa(i)
			break
		}
	}
	return ret
}
