package heartbeat

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	gonet "net"
	"runtime"
	"strings"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-engine-open/config"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var (
	heartbeat = new(HeartBeat)
	Key       = "RunnerMachineList"
)

func CheckHeartBeat() *HeartBeat {
	heartbeat.Name = GetHostName()
	heartbeat.CpuUsage = GetCpuUsed()
	heartbeat.MemInfo = GetMemInfo()
	heartbeat.CpuLoad = GetCPULoad()
	heartbeat.Networks = GetNetwork()
	heartbeat.MaxGoroutines = config.Conf.Machine.MaxGoroutines
	heartbeat.DiskInfos = GetDiskInfo()
	heartbeat.CreateTime = time.Now().Unix()
	heartbeat.FmtCreateTime = time.Now()
	heartbeat.ServerType = config.Conf.Machine.ServerType
	heartbeat.CurrentGoroutines = runtime.NumGoroutine()
	return heartbeat
}

type HeartBeat struct {
	Name              string        `json:"name"`
	CpuUsage          float64       `json:"cpu_usage"`
	CpuLoad           *load.AvgStat `json:"cpu_load"`
	MemInfo           []MemInfo     `json:"mem_info"`
	Networks          []Network     `json:"networks"`
	DiskInfos         []DiskInfo    `json:"disk_infos"`
	MaxGoroutines     int           `json:"max_goroutines"`
	CurrentGoroutines int           `json:"current_goroutines"`
	ServerType        int           `json:"server_type"`
	CreateTime        int64         `json:"create_time"`
	FmtCreateTime     time.Time     `json:"fmt_create_time"`
}

type MemInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

type DiskInfo struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type Network struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytesSent"`
	BytesRecv   uint64 `json:"bytesRecv"`
	PacketsSent uint64 `json:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv"`
}

// CPU??????

func GetCpuUsed() float64 {
	percent, _ := cpu.Percent(time.Second, false) // false??????CPU???????????????true?????????
	return percent[0]
}

// ????????????

func GetCPULoad() (info *load.AvgStat) {
	info, _ = load.Avg()
	return
}

// ????????????

func GetMemInfo() (memInfoList []MemInfo) {
	memVir := MemInfo{}
	memInfoVir, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	memVir.Total = memInfoVir.Total
	memVir.Free = memInfoVir.Free
	memVir.Used = memInfoVir.Used
	memVir.UsedPercent = memInfoVir.UsedPercent
	memInfoList = append(memInfoList, memVir)
	memInfoSwap, err := mem.SwapMemory()
	if err != nil {
		return
	}
	memVir.Total = memInfoSwap.Total
	memVir.Free = memInfoSwap.Free
	memVir.Used = memInfoSwap.Used
	memVir.UsedPercent = memInfoSwap.UsedPercent
	memInfoList = append(memInfoList, memVir)
	return memInfoList
}

// ????????????

func GetHostName() string {
	hostInfo, _ := host.Info()
	return hostInfo.Hostname
}

// ????????????

func GetDiskInfo() (diskInfoList []DiskInfo) {
	disks, err := disk.Partitions(true)
	if err != nil {
		return
	}
	for _, v := range disks {
		diskInfo := DiskInfo{}
		info, err := disk.Usage(v.Device)
		if err != nil {
			continue
		}
		diskInfo.Total = info.Total
		diskInfo.Free = info.Free
		diskInfo.Used = info.Used
		diskInfo.UsedPercent = info.UsedPercent
		diskInfoList = append(diskInfoList, diskInfo)
	}
	return
}

// ????????????

func GetNetwork() (networkList []Network) {
	netIOs, _ := net.IOCounters(true)
	if netIOs == nil {
		return
	}
	for _, netIO := range netIOs {
		network := Network{}
		network.Name = netIO.Name
		network.BytesSent = netIO.BytesSent
		network.BytesRecv = netIO.BytesRecv
		network.PacketsSent = netIO.PacketsSent
		network.PacketsRecv = netIO.PacketsRecv
		networkList = append(networkList, network)
	}
	return
}

func InitLocalIp() {

	conn, err := gonet.Dial("udp", "8.8.8.8:53")
	if err != nil {
		log.Logger.Error(fmt.Sprintf("udp?????????%s", err.Error()))
		return
	}
	localAddr := conn.LocalAddr().(*gonet.UDPAddr)
	middlewares.LocalIp = strings.Split(localAddr.String(), ":")[0]
	log.Logger.Info("??????ip???", middlewares.LocalIp)
}

func SendHeartBeatRedis(field string, duration int64) {
	for {
		CheckHeartBeat()
		hb, err := json.Marshal(heartbeat)
		if err != nil {
			log.Logger.Debug("err:   ", err)
		}
		err = model.InsertHeartbeat(Key, field, string(hb))
		if err != nil {
			log.Logger.Error(fmt.Sprintf("??????ip:%s, ??????????????????, ??????redis??????:   %s", middlewares.LocalIp, err.Error()))
		}
		time.Sleep(time.Duration(duration) * time.Second)
	}
}

func SendMachineResources(duration int64) {
	key := fmt.Sprintf("MachineMonitor:%s", middlewares.LocalIp)
	for {
		CheckHeartBeat()
		hb, _ := json.Marshal(heartbeat)
		err := model.InsertMachineResources(key, string(hb))
		if err != nil {
			log.Logger.Error(fmt.Sprintf("??????ip:%s, ??????????????????, ??????redis??????:   %s", middlewares.LocalIp, err.Error()))
		}
		time.Sleep(time.Duration(duration) * time.Second)
	}
}
