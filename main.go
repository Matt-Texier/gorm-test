package main

import (
	"fmt"
	"github.com/Pragma-innovation/gorm"
	"github.com/Pragma-innovation/gorm/dialects/postgres"
	_ "github.com/Pragma-innovation/gorm/dialects/postgres"
	"net"
	"os"
	"sync"
)

// structure to receive interface data
type ManagedInterface struct {
	gorm.Model
	ManagedRouterID uint   `json:"-"` // one to many relationship
	Name            string `json:"name,omitempty"`
	Alias           string `json:"description,omitempty"`
	HSpeed          uint64 `json:"speed,omitempty"`
	Index           int32  `json:"if-index,omitempty"`
}

// object/structure to store routers information
// FlowSourceIP is PMACCT field: source IP of netflow packets
type ManagedRouter struct {
	gorm.Model
	UniqueName        string              `json:"rtr_unique_name,omitempty"`
	Description       string              `json:"rtr_description,omitempty"`
	UpTime            string              `json:"rtr_up_time,omitempty"`
	Contact           string              `json:"rtr_contact,omitempty"`
	Name              string              `json:"rtr_name,omitempty"`
	Location          string              `json:"rtr_location,omitempty"`
	Lon               float64             `json:"rtr_lon,omitempty"`
	Lat               float64             `json:"rtr_lat,omitempty"`
	BulkMaxRepetition int                 `json:"rtr_bulk_max_rep,omitempty"`
	FlowSourceIP      postgres.Inet       `json:"rtr_flow_src_ip,omitempty"`
	PollingInterval   string              `json:"rtr_polling_interval,omitempty"`
	ManagedInterfaces []*ManagedInterface `json:"rtr_if_table"` // has many relations with interfaces
	WaitWriter        sync.WaitGroup      `json:"-" sql:"-"`    // wait group to avoid race condition
}

var someTestRouters = []*ManagedRouter{
	{
		UniqueName:        "alu-01",
		Description:       "",
		UpTime:            "",
		Contact:           "",
		Name:              "",
		Location:          "",
		BulkMaxRepetition: 10,
		FlowSourceIP: postgres.Inet{
			IP: net.ParseIP("10.0.1.1"),
		},
		ManagedInterfaces: []*ManagedInterface{
			{Name: "alu-01-int-01"},
			{Name: "alu-01-int-02"},
		},
	},
	{
		UniqueName:        "alu-02",
		Description:       "",
		UpTime:            "",
		Contact:           "",
		Name:              "",
		Location:          "",
		BulkMaxRepetition: 10,
		FlowSourceIP: postgres.Inet{
			IP: net.ParseIP("10.0.1.2"),
		},
		ManagedInterfaces: []*ManagedInterface{
			{Name: "alu-02-int-01"},
			{Name: "alu-02-int-02"},
		},
	},
	{
		UniqueName:        "vxr-01",
		Description:       "",
		UpTime:            "",
		Contact:           "",
		Name:              "",
		Location:          "",
		BulkMaxRepetition: 10,
		FlowSourceIP: postgres.Inet{
			IP: net.ParseIP("10.0.1.3"),
		},
		ManagedInterfaces: []*ManagedInterface{
			{Name: "vrx-01-int-01"},
			{Name: "vrx-01-int-02"},
		},
	},
}

func main() {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=matthieu dbname=gorm password=matthieu")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	defer db.Close()
	db.AutoMigrate(&ManagedRouter{})
	db.AutoMigrate(&ManagedInterface{})
	db.Model(&ManagedRouter{}).Related(&ManagedInterface{})
	for _, router := range someTestRouters {
		db.Create(router)
	}

}
