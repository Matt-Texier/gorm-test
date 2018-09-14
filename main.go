package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/k-sone/snmpgo"
	"math"
	"os"
	"sync"
	"time"
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
type IfCounters struct {
	LocalTimestamp   time.Time `json:"local-time,omitempty"`
	RemoteTimestamp  string    `json:"remote-time,omitempty"`
	InOctets         uint64    `json:"in-octets,omitempty"`
	InUcastPkts      uint64    `json:"in-unicast-pkts,omitempty"`
	InMulticastPkts  uint64    `json:"in-multicast-pkts,omitempty"`
	InBroadcastPkts  uint64    `json:"in-broadcast-pkts,omitempty"`
	OutOctets        uint64    `json:"out-octets,omitempty"`
	OutUcastPkts     uint64    `json:"out-unicast_pkts,omitempty"`
	OutMulticastPkts uint64    `json:"out-multicast-pkts,omitempty"`
	OutBroadcastPkts uint64    `json:"out-broadcast-pkts,omitempty"`
	NextCounters     *IfCounters
}

// structure to receive interface data
type ManagedInterface struct {
	gorm.Model
	ManagedRouterID uint        `json:"-"` // one to many relationship
	Name            string      `json:"name,omitempty" gorm:"type:varchar(255)"`
	Alias           string      `json:"description,omitempty"  gorm:"type:varchar(255)"`
	HSpeed          uint64      `json:"speed,omitempty"`
	Index           int32       `json:"if-index,omitempty"`
	Statistics      *IfCounters `json:"statistics,omitempty" gorm:"-"`
}

type InterfacesIndexType map[int32]*ManagedInterface

// object/structure to store routers information
// FlowSourceIP is PMACCT field: source IP of netflow packets
type ManagedRouter struct {
	gorm.Model
	UniqueName        string                `json:"rtr_unique_name,omitempty" gorm:"type:varchar(150);unique_index"`
	Description       string                `json:"rtr_description,omitempty" gorm:"type:varchar(255)"`
	UpTime            string                `json:"rtr_up_time,omitempty" gorm:"type:varchar(100)"`
	Contact           string                `json:"rtr_contact,omitempty" gorm:"type:varchar(100)"`
	Name              string                `json:"rtr_name,omitempty" gorm:"type:varchar(150)"`
	Location          string                `json:"rtr_location,omitempty" gorm:"type:varchar(150)"`
	Lon               float64               `json:"rtr_lon,omitempty"`
	Lat               float64               `json:"rtr_lat,omitempty"`
	BulkMaxRepetition int                   `json:"rtr_bulk_max_rep,omitempty"`
	FlowSourceIP      string                `json:"rtr_flow_src_ip,omitempty" sql:"type:inet;" `
	PollingInterval   string                `json:"rtr_polling_interval,omitempty" gorm:"type:varchar(20)"`
	SnmpConfig        *UserSnmpConfig       `json:"rtr_snmp_config,omitempty"  gorm:"foreignkey:ManagedRouterID"`
	SnmpArg           *snmpgo.SNMPArguments `json:"-" gorm:"-"`
	InterfacesIndex   InterfacesIndexType   `json:"-" gorm:"-"`
	ManagedInterfaces []*ManagedInterface   `json:"-" gorm:"foreignkey:ManagedRouterID"` // has many relations with interfaces
	WaitWriter        sync.WaitGroup        `json:"-" gorm:"-"`                          // wait group to avoid race condition
	Quit              chan bool             `json:"-" gorm:"-"`                          // channel used to send exit signal to snmp poller from the orchestrator
}

const (
	IF_REMOVE = iota
	IF_RELOAD
	IF_CREATE
	IF_UNTOUCH
)

type InterfaceDiff struct {
	Action          int
	IfMibSliceIndex int
	IfDbSliceIndex  int
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
			someTestInterfaces[0],
			someTestInterfaces[1],
			someTestInterfaces[2],
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
			someTestInterfaces[3],
			someTestInterfaces[4],
			someTestInterfaces[5],
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
			someTestInterfaces[6],
			someTestInterfaces[7],
			someTestInterfaces[8],
		},
		SnmpConfig: &UserSnmpConfig{
			Network:     "udp",
			Address:     "10.0.1.3",
			Version:     "v2c",
			V2Community: "public",
		},
	},
}

var someTestInterfaces = []*ManagedInterface{
	{
		ManagedRouterID: someTestRouters[0].ID,
		Alias:           "interface 1",
		HSpeed:          10,
		Name:            "1/1/1",
		Index:           1,
		Statistics:      nil,
	},
	{
		ManagedRouterID: someTestRouters[0].ID,
		Alias:           "interface 2",
		HSpeed:          20,
		Name:            "2/2/2",
		Index:           2,
		Statistics:      nil,
	},
	{
		ManagedRouterID: someTestRouters[0].ID,
		Alias:           "interface 3",
		HSpeed:          30,
		Name:            "3/3/3",
		Index:           3,
		Statistics:      nil,
	},
	{
		ManagedRouterID: someTestRouters[1].ID,
		Alias:           "interface 1",
		HSpeed:          10,
		Name:            "1/1/1",
		Index:           1,
		Statistics:      nil,
	},
	{
		ManagedRouterID: someTestRouters[1].ID,
		Alias:           "interface 2",
		HSpeed:          20,
		Name:            "2/2/2",
		Index:           2,
		Statistics:      nil,
	},
	{
		ManagedRouterID: someTestRouters[1].ID,
		Alias:           "interface 3",
		HSpeed:          30,
		Name:            "3/3/3",
		Index:           3,
		Statistics:      nil,
	},
	{
		ManagedRouterID: someTestRouters[2].ID,
		Alias:           "interface 1",
		HSpeed:          10,
		Name:            "1/1/1",
		Index:           1,
		Statistics:      nil,
	},
	{
		ManagedRouterID: someTestRouters[2].ID,
		Alias:           "interface 2",
		HSpeed:          20,
		Name:            "2/2/2",
		Index:           2,
		Statistics:      nil,
	},
	{
		ManagedRouterID: someTestRouters[2].ID,
		Alias:           "interface 3",
		HSpeed:          30,
		Name:            "3/3/3",
		Index:           3,
		Statistics:      nil,
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
	// UpdateRouterDbTable(db, someTestRouters)
	// SyncMemRoutersWithDb(db, someTestRouters)
	PushRouterToDb(db, someTestRouters)
	for i, router := range someTestRouters {
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

func PushRouterToDb(db *gorm.DB, myRouters []*ManagedRouter) error {
	var count int
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

func (myRouter *ManagedRouter) PushUpdateRouterInterface(db *gorm.DB) error {
	var myRouterDbInterfaces []*ManagedInterface
	db.Model(&ManagedInterface{}).Where(&ManagedInterface{
		ManagedRouterID: myRouter.ID,
	}).Find(&myRouterDbInterfaces)
	interfaceDiff, err := DiffIfMibFromDbInterface(myRouter.ManagedInterfaces, myRouterDbInterfaces)
	if err != nil {
		fmt.Printf("Diff failed")
	}
	for _, myDiff := range interfaceDiff {
		switch myDiff.Action {
		case IF_CREATE:
			{
				db.Model(&ManagedInterface{}).Create(myRouter.ManagedInterfaces[myDiff.IfMibSliceIndex])
			}
		case IF_RELOAD:
			{
				db.Model(&ManagedInterface{}).Save(myRouter.ManagedInterfaces[myDiff.IfMibSliceIndex])
			}
		case IF_UNTOUCH:
			{
				fmt.Println("skip interface", myRouter.ManagedInterfaces[myDiff.IfMibSliceIndex].Name)
			}
		case IF_REMOVE:
			{
				db.Model(&ManagedInterface{}).Delete(myRouterDbInterfaces[myDiff.IfDbSliceIndex])
			}
		default:
			{
				fmt.Println("wrong diff action")
			}
		}
	}
	return nil
}

func DiffIfMibFromDbInterface(IfsMib []*ManagedInterface,
	IfsDb []*ManagedInterface) ([]*InterfaceDiff, error) {
	var returnDiff []*InterfaceDiff
	// simple situation where interface router db is empty
	if len(IfsDb) == 0 {
		for i, _ := range IfsMib {
			returnDiff = append(returnDiff, NewDiffInterface(IF_CREATE, i, -1))
		}
		return returnDiff, nil
	}
	for i, myIfMib := range IfsMib {
		indexIfDb := findInterfaceInSlice(IfsDb, myIfMib)
		if indexIfDb != -1 { // we found the same If in the db
			// we check if data in db diff that data in mib
			isTheSame := compareInterface(myIfMib, IfsDb[indexIfDb])
			if isTheSame { // interface hasn't change
				returnDiff = append(returnDiff, NewDiffInterface(IF_UNTOUCH, i, indexIfDb))
			} else {
				returnDiff = append(returnDiff, NewDiffInterface(IF_RELOAD, i, indexIfDb))
			}

		} else { // interface in mib is not in interface in db
			returnDiff = append(returnDiff, NewDiffInterface(IF_CREATE, i, indexIfDb))
		}
	}
	// we went through the whole mib if and elaborate diff with db if
	// Now we need to check if mib interfaces has been removed since last db update
	if len(returnDiff) < len(IfsDb) {
		stillPresentIfs := make([]bool, len(IfsDb))
		for i := 0; i < len(stillPresentIfs); i++ {
			stillPresentIfs[i] = false
		}
		for i, myDbIf := range IfsDb {
			indexIfDb := findInterfaceInSlice(IfsMib, myDbIf)
			if indexIfDb == -1 {
				stillPresentIfs[i] = false
			} else {
				stillPresentIfs[i] = true
			}
		}
		for i := 0; i < len(stillPresentIfs); i++ {
			if stillPresentIfs[i] == false {
				returnDiff = append(returnDiff, NewDiffInterface(IF_REMOVE, -1, i))
			}
		}
	}
	return returnDiff, nil
}

func compareInterface(myInterfaceA *ManagedInterface, myInterfaceB *ManagedInterface) bool {
	return myInterfaceA.ManagedRouterID == myInterfaceB.ManagedRouterID &&
		myInterfaceA.HSpeed == myInterfaceB.HSpeed &&
		myInterfaceA.Alias == myInterfaceB.Alias &&
		myInterfaceA.Name == myInterfaceB.Name &&
		myInterfaceA.Index == myInterfaceB.Index
}

func findInterfaceInSlice(myIfSlice []*ManagedInterface, myInterface *ManagedInterface) int {
	for i, myInterfaceFromSlice := range myIfSlice {
		if compareInterface(myInterfaceFromSlice, myInterface) {
			return i
		}
	}
	return -1
}

func NewDiffInterface(myAction int, myIfMibSliceIndex int, myIfDbSliceIndex int) *InterfaceDiff {
	return &InterfaceDiff{
		Action:          myAction,
		IfDbSliceIndex:  myIfDbSliceIndex,
		IfMibSliceIndex: myIfMibSliceIndex,
	}
}
