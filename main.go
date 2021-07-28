package main

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func main() {
	router := gin.Default()
	v1 := router.Group("/api/v1/gpu-infos")
	{
		v1.GET("/", GetGpuInfos)
	}
	_ = router.Run()

}

type NodeDevice struct {
	Name        string `json:"name"`
	PowerUsage  string `json:"powerUsage"`
	Temperature int    `json:"temperature"`
	TotalMemory string `json:"totalMemory"`
	UsedMemory  string `json:"usedMemory"`
}

type NodeExtension struct {
	DeviceNum   int          `json:"deviceNum"`
	Devices     []NodeDevice `json:"devices"`
	TotalMemory string       `json:"totalMemory"`
	UsedMemory  string       `json:"usedMemory"`
}

func GetGpuInfos(c *gin.Context) {
	ex := NodeExtension{}
	ex.Devices = make([]NodeDevice, 0)
	sum, err := NumDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "NDevices() error", "data": ""})
		log.Error(err)
	}
	var allUsed, allTotal int64
	contextArray := make([]CUContext, 0)
	for i := 0; i < sum; i++ {
		d, err := GetDevice(i)
		n := NodeDevice{}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "GetDevice() error", "data": ""})
			log.Error(err)
		}
		if &d != nil {
			name, err := d.Name()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "Name() error", "data": ""})
				log.Error(err)
			}
			n.Name = name
		}

		ctx, err := Device(i).MakeContext(SchedAuto)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "MakeContext() error", "data": ""})
			log.Error(err)
		}
		contextArray = append(contextArray, ctx)

		free, total, err := MemInfo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "MemInfo() error", "data": ""})
			log.Error(err)
		}
		n.TotalMemory = IntToString(total)
		n.UsedMemory = IntToString(total - free)
		ex.Devices = append(ex.Devices, n)
		allTotal += total
		allUsed += total - free
	}
	ex.TotalMemory = IntToString(allTotal)
	ex.UsedMemory = IntToString(allUsed)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": ex})
	for _, x := range contextArray {
		err := x.Destroy()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Destroy() error", "data": ""})
			log.Error(err)
		}
	}
}

func IntToString(value int64) string {
	return strconv.FormatInt(value, 10)
}
