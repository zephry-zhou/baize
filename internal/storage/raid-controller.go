package storage

import (
	"baize/internal/utils"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type RHController struct {
	Cid                       string          `json:"Controller ID,omitempty"` //RAID卡编号
	ProductName               string          `json:"Product Name,omitempty"`  //产品名称
	CacheSize                 string          `json:"Cache Size,omitempty"`    //控制卡缓存大小
	SerialNumber              string          `json:"SN,omitempty"`            //sn
	SasAddress                string          `json:"SAS Address,omitempty"`   //sas地址
	ControllerTime            string          `json:",omitempty"`              //控制器当前时间
	Firmware                  string          `json:",omitempty"`              //固件版本
	BiosVersion               string          `json:",omitempty"`              //bios版本
	FwVersion                 string          `json:",omitempty"`              //fw_version
	CurrentPersonality        string          `json:",omitempty"`              //当前工作模式
	NumberOfRaid              string          `json:",omitempty"`              //当前逻辑硬盘数
	FailedRaid                string          `json:",omitempty"`              //失败的逻辑盘数
	DegradedRaid              string          `json:",omitempty"`              //降级的逻辑盘数
	NumberOfDisk              string          `json:",omitempty"`              //附属硬盘数
	FailedDisk                string          `json:",omitempty"`              //失败硬盘数
	CriticalDisk              string          `json:",omitempty"`              //出现致命错误硬盘数
	ControllerStatus          string          `json:",omitempty"`              //控制器当前状态
	MemoryCorrectableErrors   string          `json:",omitempty"`              //缓存可纠正错误
	MemoryUncorrectableErrors string          `json:",omitempty"`              //缓存不可纠正错误
	ChipRevision              string          `json:",omitempty"`              //修订固件版本
	FrontEndPortCount         string          `json:",omitempty"`              //前背板接口数量
	BackendPortCount          string          `json:",omitempty"`              //后背板接口数量
	NumberOfBackplane         string          `json:",omitempty"`              // 硬盘背板数量
	NVRAMSize                 string          `json:",omitempty"`              //NVRAMSize
	FlashSize                 string          `json:",omitempty"`              //FlashSize
	SupportedDrives           string          `json:",omitempty"`              //支持硬盘类型
	RaidLevelSupported        string          `json:",omitempty"`              //支持raid类型
	EnableJBOD                string          `json:",omitempty"`              //jbod使能
	HostInterface             string          `json:",omitempty"`              //raid卡接口
	DeviceInterface           string          `json:",omitempty"`              //硬盘接口
	Diagnose                  string          `json:",omitempty"`              //raid卡健康诊断
	DiagnoseDetail            string          `json:",omitempty"`              //raid卡诊断详情
	BackPlanes                []backplate     `json:",omitempty"`
	Battery                   bbu             `json:",omitempty"`
	LogicalDriveList          []logicalDrive  `json:",omitempty"`
	PhysicalDriveList         []physicalDrive `json:",omitempty"`
}

type bbu struct {
	Model         string `json:",omitempty"`
	State         string `json:",omitempty"`
	Temp          string `json:",omitempty"`
	RetentionTime string `json:",omitempty"`
	Mode          string `json:",omitempty"`
	MfgDate       string `json:",omitempty"`
}

type backplate struct {
	Location              string `json:",omitempty"` //背板位置
	Eid                   string `json:",omitempty"` //背板id
	State                 string `json:",omitempty"` //背板状态
	Slots                 string `json:",omitempty"` //背板插槽编号
	PhysicalDriveCount    string `json:",omitempty"` //背板硬盘总数
	ConnectorName         string `json:",omitempty"` //背板接口名
	EnclosureType         string `json:",omitempty"` //背板类型
	EnclosureSerialNumber string `json:",omitempty"` //背板sn
	DeviceType            string `json:",omitempty"` //背板设备类型
	Vendor                string `json:",omitempty"` //背板厂商
	ProductIdentification string `json:",omitempty"` //背板产品标识
	ProductRevisionLevel  string `json:",omitempty"`
}

type logicalDrive struct {
	Location              string          `json:"Location,omitempty"`                  //逻辑硬盘位置
	VD                    string          `json:"Virtual Drive,omitempty"`             //逻辑硬盘id
	DG                    string          `json:"Drive Group,omitempty"`               //逻辑硬盘组标识
	Type                  string          `json:"RAID Level,omitempty"`                //逻辑硬盘类型
	SpanDepth             string          `json:"Span Depth,omitempty"`                //逻辑硬盘深度
	Capacity              string          `json:"Capacity,omitempty"`                  //逻辑硬盘容量
	State                 string          `json:"State,omitempty"`                     //逻辑硬盘状态
	Access                string          `json:"Access,omitempty"`                    //逻辑硬盘读写状态
	Consist               string          `json:"Consistent,omitempty"`                //逻辑硬盘一致性状态
	Cache                 string          `json:"Current Cache Policy,omitempty"`      //逻辑硬盘缓存策略
	StripSize             string          `json:"Strip Size,omitempty"`                //逻辑硬盘块大小
	NumberOfBlocks        string          `json:"Number of Block,omitempty"`           //逻辑硬盘块数量
	NumberOfDrivesPerSpan string          `json:"Number of Drives per Span,omitempty"` //逻辑硬盘每层硬盘数量
	MappingFile           string          `json:"Mapping file,omitempty"`              //逻辑硬盘对应系统块设备
	CreateTime            string          `json:"Create Time,omitempty"`               //逻辑硬盘创建时间
	ScsiNaaId             string          `json:"SCSI NAA ID,omitempty"`               //逻辑硬盘scsi编号
	PhysicalDrives        []physicalDrive `json:"Physical Drives,omitempty"`           //逻辑盘包含的物理硬盘
}

type physicalDrive struct {
	Location               string                   `json:",omitempty"` //物理硬盘位置
	EnclosureId            string                   `json:",omitempty"` //物理硬盘背板编号
	SlotId                 string                   `json:",omitempty"` //物理硬盘插槽编号
	DeviceId               string                   `json:",omitempty"` //物理硬盘设备编号
	DG                     string                   `json:"Drive Group,omitempty"`
	Vendor                 string                   `json:",omitempty"` //物理硬盘厂商
	Product                string                   `json:",omitempty"` //物理硬盘产品名称
	Capacity               string                   `json:",omitempty"` //物理硬盘容量
	State                  string                   `json:",omitempty"` //物理硬盘状态
	SN                     string                   `json:",omitempty"` //物理硬盘sn
	Interface              string                   `json:",omitempty"` //物理硬盘接口
	MediumType             string                   `json:",omitempty"` //物理硬盘类型
	DeviceSpeed            string                   `json:",omitempty"` //物理硬盘设备速度
	LinkSpeed              string                   `json:",omitempty"` //物理硬盘链路速度
	RotationRate           string                   `json:",omitempty"` //物理硬盘转速
	FormFactor             string                   `json:",omitempty"` //物理硬盘尺寸
	Firmware               string                   `json:",omitempty"` //物理硬盘固件
	OemVendor              string                   `json:",omitempty"` //物理硬盘oem厂商
	Model                  string                   `json:",omitempty"` //物理硬盘Model
	RebuildInfo            string                   `json:",omitempty"` //物理硬盘重建信息
	WriteCache             string                   `json:",omitempty"` //物理硬盘写缓存
	ReadCache              string                   `json:",omitempty"` //物理硬盘读缓存
	LogicalSectorSize      string                   `json:",omitempty"` //物理硬盘逻辑扇区大小
	PhysicalSectorSize     string                   `json:",omitempty"` //物理硬盘物理扇区大小
	MappingFile            string                   `json:",omitempty"` //物理硬盘映射系统块设备名称
	WWN                    string                   `json:",omitempty"` //物理硬盘WWN
	SmartAttribute         []map[string]interface{} `json:",omitempty"` //Smart属性
	Diagnose               string                   `json:",omitempty"` //物理硬盘健康分析接口
	DiagnoseDetail         string                   `json:",omitempty"` //物理硬盘健康分析详情
	PowerOnTime            string                   `json:",omitempty"` //物理硬盘通电时间
	MediaWearoutIndicator  string                   `json:",omitempty"` //SSD磨损值
	AvailableReservdSpace  string                   `json:",omitempty"` //可用的预留闪存数量
	Temperature            string                   `json:",omitempty"` //物理硬盘温度
	OtherErrorCount        string                   `json:",omitempty"` //物理硬盘其他错误数
	MediaErrorCount        string                   `json:",omitempty"` //物理硬盘物理媒介错误数
	PredictiveFailureCount string                   `json:",omitempty"` //
	SmartHealthStatus      string                   `json:",omitempty"` //物理硬盘SMART状态
	SmartAlert             string                   `json:",omitempty"` //物理硬盘smart警告
	Type                   string                   `json:",omitempty"` // type
}

var run utils.RunSheller = &utils.RunShell{}

func GetController() []RHController {
	byteCard, err := run.Command("sh", "-c", `lspci -Dn | egrep '\010[47]' | grep -v ^$`)
	if err != nil || len(byteCard) == 0 {
		log.Printf("failed to search raid controller through lspci: %v", err)
		return []RHController{}
	}
	ret := []RHController{}
	lines := utils.SplitAndTrim(string(byteCard), "\n")
	vrocFlag := false
	ctrNum := len(lines)

	for i := 0; i < ctrNum; i++ {
		fields := strings.Fields(lines[i])
		if strings.HasPrefix(fields[2], "1000:") {
			ret = append(ret, Storcli(fields[0], ctrNum))
		} else if strings.HasPrefix(fields[2], "103c:") {
			ret = append(ret, Hpssacli(fields[0], ctrNum))
		} else if strings.HasPrefix(fields[2], "8086:") {
			if !vrocFlag {
				ret = append(ret, Mdadm(fields[0], i))
				vrocFlag = true
			}
		} else {
			log.Println("not support raid controller vendor: ", fields[2])
		}
	}
	return ret
}

func rawValue(raw interface{}) float64 {
	var ret float64
	for key, value := range raw.(map[string]interface{}) {
		if key == "value" {
			ret = value.(float64)
		}
	}
	return ret
}

func diskDiagnose(ret *physicalDrive) {
	UnhealthyReason := make([]string, 0)
	if ret.MediaErrorCount != "0" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Media Error Count: %s", ret.MediaErrorCount))
	}
	if ret.PredictiveFailureCount != "0" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Predictive Failure Count: %s", ret.PredictiveFailureCount))
	}
	if ret.OtherErrorCount != "0" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Other Error Count: %s", ret.OtherErrorCount))
	}
	if powerTime, err := strconv.Atoi(ret.PowerOnTime); err == nil && powerTime > 61320 {
		UnhealthyReason = append(UnhealthyReason, "Power On Time more than 7 years")
	}
	if ret.SmartAlert != "No" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Smart Alert: %s", ret.SmartAlert))
	}
	if ret.SmartHealthStatus != "true" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Smart Health Status: %s", ret.SmartHealthStatus))
	}
	checkElements := map[float64]int{1: 0, 5: 0, 184: 0, 197: 0, 198: 0, 199: 0}
	for _, attr := range ret.SmartAttribute {
		if id, ok := attr["id"].(float64); ok {
			if _, ok := checkElements[id]; ok && rawValue(attr["raw"]) != 0 {
				UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Smart Attribute %v: %s", id, attr["value"]))
			}
		} else {
			for key, value := range attr {
				if val, ok := value.(float64); ok && val != 0 {
					UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Smart Attribute %s: %s", key, value))
				}
			}
		}
	}
	if len(UnhealthyReason) != 0 {
		ret.DiagnoseDetail = strings.Join(UnhealthyReason, "; ")
		ret.Diagnose = "Unhealthy"
	} else {
		ret.Diagnose = "Healthy"
	}
}
