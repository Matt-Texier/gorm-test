package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"math"
	"os"
	"sync"
)

type UserSnmpConfig struct {
	gorm.Model
	ManagedRouterID    uint   `json:"-"` // one relationship
	Network            string `json:"snmp_network,omitempty"`
	Address            string `json:"snmp_address,omitempty"`
	Timeout            string `json:"snmp_timeout,omitempty"`
	Retries            string `json:"snmp_retries,omitempty"`
	MessageMaxSize     string `json:"snmp_message_size,omitempty"`
	Version            string `json:"snmp_version,omitempty"`
	V2Community        string `json:"snmpv2_community,omitempty"`
	V3UserName         string `json:"snmpv3_user_name,omitempty"`
	V3SecurityLevel    string `json:"snmpv3_security_level,omitempty"`
	V3AuthPassword     string `json:"snmpv3_auth_passwd,omitempty"`
	V3AuthProtocol     string `json:"snmpv3_auth_proto,omitempty"`
	V3PrivPassword     string `json:"snmpv3_priv_password,omitempty"`
	V3PrivProtocol     string `json:"snmpv3_priv_proto,omitempty"`
	V3SecurityEngineId string `json:"snmpv3_sec_engine_id,omitempty"`
	V3ContextEngineId  string `json:"snmpv3_context_engine_id,omitempty"`
	V3ContextName      string `json:"snmpv3_context_name,omitempty"`
}

// structure to receive interface data
type ManagedInterface struct {
	gorm.Model
	ManagedRouterID uint   `json:"-"` // one to many relationship
	Name            string `json:"name,omitempty" gorm:"type:varchar(150)"`
	Alias           string `json:"description,omitempty" gorm:"type:varchar(250)"`
	HSpeed          uint64 `json:"speed,omitempty"`
	Index           int32  `json:"if-index,omitempty"`
}

// object/structure to store routers information
// FlowSourceIP is PMACCT field: source IP of netflow packets
type ManagedRouter struct {
	gorm.Model
	UniqueName        string              `json:"rtr_unique_name,omitempty" gorm:"type:varchar(150);unique_index"`
	Description       string              `json:"rtr_description,omitempty"`
	UpTime            string              `json:"rtr_up_time,omitempty"`
	Contact           string              `json:"rtr_contact,omitempty" gorm:"type:varchar(150)"`
	Name              string              `json:"rtr_name,omitempty" gorm:"type:varchar(150)"`
	Location          string              `json:"rtr_location,omitempty" gorm:"type:varchar(150)"`
	Lon               float64             `json:"rtr_lon,omitempty"`
	Lat               float64             `json:"rtr_lat,omitempty"`
	BulkMaxRepetition int                 `json:"rtr_bulk_max_rep,omitempty"`
	FlowSourceIP      string              `json:"rtr_flow_src_ip,omitempty" sql:"type:inet;"`
	PollingInterval   string              `json:"rtr_polling_interval,omitempty" gorm:"type:varchar(10)"`
	SnmpConfig        *UserSnmpConfig     `json:"rtr_snmp_config,omitempty" gorm:"foreignkey:ManagedRouterID"` // has one relation
	ManagedInterfaces []*ManagedInterface `json:"rtr_if_table" gorm:"foreignkey:ManagedRouterID"`              // has many relations with interfaces
	WaitWriter        sync.WaitGroup      `json:"-" sql:"-"`                                                   // wait group to avoid race condition
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
		FlowSourceIP:      "10.0.1.1",
		ManagedInterfaces: []*ManagedInterface{
			{Name: "alu-01-int-01"},
			{Name: "alu-01-int-02"},
		},
		SnmpConfig: &UserSnmpConfig{
			Network:     "udp",
			Address:     "10.0.1.1",
			Version:     "v2c",
			V2Community: "public",
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
		FlowSourceIP:      "10.0.1.2",
		ManagedInterfaces: []*ManagedInterface{
			{Name: "alu-02-int-01"},
			{Name: "alu-02-int-02"},
		},
		SnmpConfig: &UserSnmpConfig{
			Network:     "udp",
			Address:     "10.0.1.2",
			Version:     "v2c",
			V2Community: "public",
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
		FlowSourceIP:      "10.0.1.3",
		ManagedInterfaces: []*ManagedInterface{
			{Name: "vrx-01-int-01"},
			{Name: "vrx-01-int-02"},
		},
		SnmpConfig: &UserSnmpConfig{
			Network:     "udp",
			Address:     "10.0.1.3",
			Version:     "v2c",
			V2Community: "public",
		},
	},
}

func main() {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=matthieu dbname=gorm password=matthieu")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	if !db.HasTable(&ManagedRouter{}) {
		db.CreateTable(&ManagedRouter{})
	}
	if !db.HasTable(&ManagedInterface{}) {
		db.AutoMigrate(&ManagedInterface{})
		db.Model(&ManagedRouter{}).Related(&ManagedInterface{})
	}
	if !db.HasTable(&UserSnmpConfig{}) {
		db.AutoMigrate(&UserSnmpConfig{})
		db.Model(&ManagedRouter{}).Related(&UserSnmpConfig{})
	}
	UpdateRouterDbTable(db, someTestRouters)
	SyncMemRoutersWithDb(db, someTestRouters)
	_, routersConfig := LoadRoutersConfigFromDb(db)
	for i, router := range routersConfig {
		fmt.Println("router ", i, ":", router)
		fmt.Println("interfaces: ", router.ManagedInterfaces)
		fmt.Println("SNMP config: ", router.SnmpConfig)
	}
	fmt.Println("#######################################################################")
	_, routersFull := LoadRoutersConfigFromDb(db)
	for i, router := range routersFull {
		fmt.Println("router ", i, ":", router)
		fmt.Println("interfaces: ", router.ManagedInterfaces)
		fmt.Println("SNMP config: ", router.SnmpConfig)
	}
	db.Close()
}

// This function updates postGRE router table from router table in memory
// Once done it also make sure that memory structure are in sync with data
// stored in the db (primary key and other fields created by the ORM
// It pushes as well all related structures (interfaces and snmp config) and sync it
// as well if routers has never been stored it creates all tables entries
// (this should in theory never happen but could be useful for test purposes)

func UpdateRouterDbTable(db *gorm.DB, myRouters []*ManagedRouter) error {
	var count int
	var routerSearch []ManagedRouter
	db.Model(&ManagedRouter{}).Count(&count)
	// router table is empty we fulfil it with routers in memory
	if count == 0 {
		for _, router := range myRouters {
			db.Create(router)
		}
		// routers has been created in db, now we need to sync them with routers
		// in memory to get all fields created by ORM (primary key ID, date, ...)
		err := SyncMemRoutersWithDb(db, myRouters)
		if err != nil {
			return err
		}
	} else {
		for i, router := range myRouters {
			db.Where(&ManagedRouter{
				UniqueName: router.UniqueName,
			}).Find(&routerSearch)
			if !CompareRouter(routerSearch[0], *router) {
				myRouters[i].Copy(&routerSearch[0])
			}
		}
	}
	return nil
}

func SyncMemRoutersWithDb(db *gorm.DB, myRouters []*ManagedRouter) error {
	var routerSearch []*ManagedRouter
	for i, router := range myRouters {
		db.Where(&ManagedRouter{
			UniqueName: router.UniqueName,
		}).Find(&routerSearch)
		if len(routerSearch) == 1 {
			myRouters[i].Copy(routerSearch[0])
		} else {
			return fmt.Errorf("bad amount of matching unique key router in db")
		}
	}
	return nil
}

func LoadRoutersConfigFromDb(db *gorm.DB) (error, []*ManagedRouter) {
	var count int
	var returnedRouters []*ManagedRouter
	db.Model(&ManagedRouter{}).Count(&count)
	if count == 0 {
		return fmt.Errorf("no router in db"), nil
	}
	db.Model(&ManagedRouter{}).Preload("SnmpConfig").Find(&returnedRouters)
	return nil, returnedRouters
}

func LoadFullRoutersFromDb(db *gorm.DB) (error, []*ManagedRouter) {
	var count int
	var returnedRouters []*ManagedRouter
	db.Model(&ManagedRouter{}).Count(&count)
	if count == 0 {
		return fmt.Errorf("no router in db"), nil
	}
	db.Model(&ManagedRouter{}).Preload("SnmpConfig").Preload("ManagedInterfaces").Find(&returnedRouters)
	return nil, returnedRouters
}

func CompareRouter(router1 ManagedRouter, router2 ManagedRouter) bool {
	return router1.Model.ID == router2.Model.ID &&
		router1.Model.CreatedAt == router2.Model.CreatedAt &&
		router1.Model.UpdatedAt == router2.Model.UpdatedAt &&
		router1.UniqueName == router2.UniqueName &&
		router1.Description == router2.Description &&
		router1.UpTime == router2.UpTime &&
		router1.Contact == router2.Contact &&
		router1.Name == router2.Name &&
		router1.Location == router2.Location &&
		almostEqual(router1.Lon, router2.Lon) &&
		almostEqual(router1.Lat, router2.Lat) &&
		router1.BulkMaxRepetition == router2.BulkMaxRepetition &&
		router1.FlowSourceIP == router2.FlowSourceIP &&
		router1.PollingInterval == router2.PollingInterval
}

func (myRouter *ManagedRouter) Copy(sourceRouter *ManagedRouter) {
	myRouter.ID = sourceRouter.ID
	myRouter.CreatedAt = sourceRouter.CreatedAt
	myRouter.UpdatedAt = sourceRouter.UpdatedAt
	myRouter.UniqueName = sourceRouter.UniqueName
	myRouter.Description = sourceRouter.Description
	myRouter.UpTime = sourceRouter.UpTime
	myRouter.Contact = sourceRouter.Contact
	myRouter.Name = sourceRouter.Name
	myRouter.Location = sourceRouter.Location
	myRouter.Lon = sourceRouter.Lon
	myRouter.Lat = sourceRouter.Lat
	myRouter.BulkMaxRepetition = sourceRouter.BulkMaxRepetition
	myRouter.FlowSourceIP = sourceRouter.FlowSourceIP
	myRouter.PollingInterval = sourceRouter.PollingInterval
}

func almostEqual(a, b float64) bool {
	const float64EqualityThreshold = 1e-9
	return math.Abs(a-b) <= float64EqualityThreshold
}

func (myRouter *ManagedRouter) CreateInterfaceFromDB(db *gorm.DB) error {
	// first let's check this router do have a db primary key
	if myRouter.ID == 0 {
		return fmt.Errorf("router is missing a primary key")
	}
	db.Model(&ManagedInterface{}).Where(&ManagedInterface{
		ManagedRouterID: myRouter.ID,
	}).Find(&myRouter.ManagedInterfaces)
	return nil
}

func (myRouter *ManagedRouter) CreateSnmpUserConfFromDb(db *gorm.DB) error {
	myUserSnmp := NewSnmpStruc()
	if myRouter.ID == 0 {
		return fmt.Errorf("router is missing a primary key")
	}
	db.Model(&UserSnmpConfig{}).Where(&UserSnmpConfig{
		ManagedRouterID: myRouter.ID,
	}).Find(myUserSnmp)
	myRouter.SnmpConfig = myUserSnmp
	return nil
}

func NewSnmpStruc() *UserSnmpConfig {
	return &UserSnmpConfig{}
}
