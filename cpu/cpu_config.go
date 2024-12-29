package cpu

type CPU struct {
	ModelName      string   `json:"Product Name,omitempty"`          // CPU型号名称
	Vendor         string   `json:"Vendor,omitempty"`                // 厂商
	Architecture   string   `json:"Architecture,omitempty"`          // CPU架构
	Hyper          string   `json:"Hyper Threading,omitempty"`       // 是否支持超线程
	Power          string   `json:"Power State,omitempty"`           // 电源状态：performance or powerSave
	OpMode         string   `json:"CPU op-mode,omitempty"`           // CPU模式
	ByteOrder      string   `json:"Byte Order,omitempty"`            // 字节序
	AdrSize        string   `json:"Address Size,omitempty"`          // 地址大小
	Threads        string   `json:"Number Of Threads,omitempty"`     // 线程数
	OnLineThreads  string   `json:"Online CPUs,omitempty"`           // 在线CPU
	ThrPerCore     string   `json:"Threads per Core,omitempty"`      // 每核心线程数
	CorePerSocket  string   `json:"Cores per Socket,omitempty"`      // 每Socket核心数
	Socket         string   `json:"Sockets,omitempty"`               // 插槽数
	NUMANode       string   `json:"NUMA Node,omitempty"`             // NUMA节点
	Family         string   `json:"Family,omitempty"`                // CPU系列
	Model          string   `json:"Model,omitempty"`                 // CPU型号
	Step           string   `json:"Stepping,omitempty"`              // 步进
	BogoMIPS       string   `json:"BogoMIPS,omitempty"`              // BogoMIPS
	Virtualization string   `json:"Virtualization,omitempty"`        // 虚拟化
	MinFreq        string   `json:"Minimum Frequency,omitempty"`     // 最小频率
	MaxFreq        string   `json:"Maximum Frequency,omitempty"`     // 最大频率
	Temp           string   `json:"Temperature,omitempty"`           // 温度
	Watt           string   `json:"Power Consumption,omitempty"`     // 电源消耗
	L1d            string   `json:"L1d Cache,omitempty"`             // L1d缓存
	L1i            string   `json:"L1i Cache,omitempty"`             // L1i缓存
	L2             string   `json:"L2 Cache,omitempty"`              // L2缓存
	L3             string   `json:"L3 Cache,omitempty"`              // L3缓存
	Flags          []string `json:"Flags,omitempty"`                 // CPU特性
	PhyCPU         []phyCPU `json:"Physical CPU Entities,omitempty"` // 物理CPU列表
}

type phyCPU struct {
	SocketID     string   `json:"Socket ID,omitempty"`         // 插槽ID
	Family       string   `json:"Family,omitempty"`            // CPU系列
	Manufacturer string   `json:"Vendor,omitempty"`            // 厂商
	Signature    string   `json:"Signature,omitempty"`         // CPU标识
	Version      string   `json:"Prodcut Name,omitempty"`      // CPU型号名称
	Voltage      string   `json:"Voltage,omitempty"`           // 电压
	ExClock      string   `json:"External Speed,omitempty"`    // 外部时钟
	MaxSpeed     string   `json:"Max Speed,omitempty"`         // 最大频率
	CurSpeed     string   `json:"Based Speed,omitempty"`       // 当前频率,基础频率
	Status       string   `json:"State,omitempty"`             // 状态
	Cores        string   `json:"Cores,omitempty"`             // 核心数
	CoreEnable   string   `json:"Core Enabled,omitempty"`      // 启用核心数
	Threads      string   `json:"Threads,omitempty"`           // 线程数
	Temp         string   `json:"Temperature,omitempty"`       // 温度
	Watt         string   `json:"Power Consumption,omitempty"` // 电源消耗
	ThreadList   []thread `json:"Thread List,omitempty"`       // 线程列表
}

type thread struct {
	Processor string `json:"Processor,omitempty"`      // 线程ID
	Freq      string `json:"Core Frequency,omitempty"` // 线程运行频率
	PhyID     string `json:"Physical ID,omitempty"`    // 线程所属物理ID
	CoreID    string `json:"Core ID,omitempty"`        // 线程所属核心ID
	Temp      string `json:"Temperature,omitempty"`    // 线程温度
}

var vendorMap = map[string]string{
	"GenuineIntel": "Intel",
	"AuthenticAMD": "AMD",
}
