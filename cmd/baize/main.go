package main

import "baize/internal/network"

//func main() {
//	cpuinfo := storage.GetController()
//	jsonCPU, err := json.MarshalIndent(cpuinfo, " ", "   ")
//	if err != nil {
//		print(err)
//	}
//	fmt.Println(string(jsonCPU))
//}

func main() {
	c := network.NETWORK{}
	c.Result()
	c.BriefFormat()
	// jsonCPU, err := json.MarshalIndent(c, " ", "   ")
	//
	//	if err != nil {
	//		print(err)
	//	}
	//
	// fmt.Println(string(jsonCPU))
}
