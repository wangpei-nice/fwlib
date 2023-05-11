package main

/*
#cgo CFLAGS: -I./src
#cgo LDFLAGS: -L./../../ -lfwlib32 -Wl,-rpath=./../../
#include <stdlib.h>
#include "../../fwlib32.h"
*/
import "C"

import (
	"fmt"
	"math"
	_ "os"
	_ "reflect"
	"strconv"
	_ "strconv"
	"strings"
	"unsafe"
)

//------------------------------------------------------
type Client struct {
	Address   string
	Port      int
	handle    C.ushort
	Timeout   int32
	Connected bool
	Cnc_type  string
}

func NewClient(address string, port int, timeout int32) Client {
	client := Client{}
	client.Address = address
	client.Connected = false
	client.Port = port
	client.Timeout = timeout
	return client
}

func (client *Client) StartupProcess() (err error) {
	log_level := 0
	log_fname := C.CString("focas.log")
	defer C.free(unsafe.Pointer(log_fname))
	if ret := C.cnc_startupprocess(C.long(log_level), log_fname); ret != C.EW_OK {
		return fmt.Errorf("cnc_startupprocess failed (%d)\n", ret)
	}
	return nil
}

func (client *Client) ExitProcess() (err error) {
	if ret := C.cnc_exitprocess(); ret != C.EW_OK {
		return fmt.Errorf("cnc_exitprocess failed (%d)\n", ret)
	}
	return nil
}

func (client *Client) Connect() (err error) {
	ret := C.cnc_allclibhndl3(C.CString(client.Address), C.ushort(client.Port), C.long(client.Timeout), &client.handle)
	if ret == 0 {
		client.Connected = true
	} else {
		fmt.Println("Connect Error")
		return fmt.Errorf("Connect Error")
	}
	return nil
}

func (client *Client) DisConnect() (err error) {
	C.cnc_freelibhndl(client.handle)
	client.Connected = false
	return nil
}

func (client *Client) CncProducts() (ret C.short, result int32) {
	obj := C.IODBPSD{}
	ret = C.cnc_rdparam(client.handle, 6712, 0, 4+C.MAX_AXIS, &obj)
	u := unsafe.Pointer(uintptr(unsafe.Pointer(&obj)) + 4)
	ldata := *(*int32)(u)
	//fmt.Println(obj)
	fmt.Println("total parts: ", ldata)

	obj2 := C.ODBM{}
	ret = C.cnc_rdmacro(client.handle, 0xf3d, 0x0a, &obj2)
	fmt.Println("当前工件数：", obj2.mcr_val/C.long(math.Pow(10, float64(obj2.dec_val))))
	fmt.Println(obj2)

	obj3 := C.IODBPSD{}
	ret = C.cnc_rdparam(client.handle, 6713, 0, 4+C.MAX_AXIS, &obj3)
	u = unsafe.Pointer(uintptr(unsafe.Pointer(&obj3)) + 4)
	ldata = *(*int32)(u)
	//fmt.Println(obj3)
	fmt.Println("required parts: ", ldata)

	obj4 := C.IODBPSD{}
	ret = C.cnc_rdparam(client.handle, 6711, 0, 4+C.MAX_AXIS, &obj4)
	u = unsafe.Pointer(uintptr(unsafe.Pointer(&obj4)) + 4)
	ldata = *(*int32)(u)
	//fmt.Println(obj4)
	fmt.Println("parts count: ", ldata)
	fmt.Println("--------------------------------")

	return ret, ldata
}

// Series 15/15i
// typedef struct odbst {
//      short  dummy[2];  /* Not used                           */
//      short  aut;       /* AUTOMATIC mode selection           */
//      short  manual;    /* MANUAL mode selection              */
//      short  run;       /* Status of automatic operation      */
//      short  edit;      /* Status of program editing          */
//      short  motion;    /* Status of axis movement,dwell      */
//      short  mstb;      /* Status of M,S,T,B function         */
//      short  emergency; /* Status of emergency                */
//      short  write;     /* Status of writing backupped memory */
//      short  labelskip; /* Status of label skip               */
//      short  alarm;     /* Status of alarm                    */
//      short  warning;   /* Status of warning                  */
//      short  battery;   /* Status of battery                  */
// } ODBST ;

// Series 16/18/21, 16i/18i/21i, 0i, 30i/31i/32i, Power Mate i, PMi-A
// typedef struct odbst {
// 	short  hdck ;        /* Status of manual handle re-trace */
// 	short  tmmode ;      /* T/M mode selection              */
// 	short  aut ;         /* AUTOMATIC/MANUAL mode selection */
// 	short  run ;         /* Status of automatic operation   */
// 	short  motion ;      /* Status of axis movement,dwell   */
// 	short  mstb ;        /* Status of M,S,T,B function      */
// 	short  emergency ;   /* Status of emergency             */
// 	short  alarm ;       /* Status of alarm                 */
// 	short  edit ;        /* Status of program editing       */
// } ODBST ;
func (client *Client) CncStatInfo() (ret C.short, result int32) {
	obj := C.ODBST{}
	ret = C.cnc_statinfo(client.handle, &obj)

	fmt.Println("stat info: ", obj)
	fmt.Println("hdck: ", obj.hdck)
	fmt.Println("tmmode: ", obj.tmmode)
	fmt.Println("aut: ", obj.aut)
	fmt.Println("run: ", obj.run)
	fmt.Println("motion: ", obj.motion)
	fmt.Println("mstb: ", obj.mstb)
	fmt.Println("emergency: ", obj.emergency)
	fmt.Println("alarm: ", obj.alarm)
	fmt.Println("edit: ", obj.edit)
	fmt.Println("--------------------------------")
	return ret, 0
}

// typedef struct odbsys {
// 	short   addinfo ;    /* Additional information */
// 	short   max_axis ;   /* Max. controlled axes */
// 	char    cnc_type[2] ;/* Kind of CNC (ASCII) */
// 	char    mt_type[2] ; /* Kind of M/T/TT (ASCII) */
// 	char    series[4] ;  /* Series number (ASCII) */
// 	char    version[4] ; /* Version number (ASCII) */
// 	char    axes[2] ;   /* Current controlled axes(ASCII)*/
// } ODBSYS ;
func (client *Client) CncSysInfo() (ret C.short, result int32) {
	obj := C.ODBSYS{}
	ret = C.cnc_sysinfo(client.handle, &obj)

	fmt.Println(obj.max_axis)

	// get cnc type
	r := rune(obj.cnc_type[0])
	str := string(r)
	r = rune(obj.cnc_type[1])
	str += string(r)
	client.Cnc_type = str
	fmt.Println("cnc type: ", str)
	str = ""

	// get mt_type
	r = rune(obj.mt_type[0])
	str += string(r)
	r = rune(obj.mt_type[1])
	str += string(r)
	str = strings.TrimSpace(str)
	fmt.Println("mt_type: ", str)
	str = ""

	// get series
	r = rune(obj.series[0])
	str += string(r)
	r = rune(obj.series[1])
	str += string(r)
	r = rune(obj.series[2])
	str += string(r)
	r = rune(obj.series[3])
	str += string(r)
	fmt.Println("series: ", str)
	str = ""

	// get version
	r = rune(obj.version[0])
	str += string(r)
	r = rune(obj.version[1])
	str += string(r)
	r = rune(obj.version[2])
	str += string(r)
	r = rune(obj.version[3])
	str += string(r)
	fmt.Println("version: ", str)
	str = ""

	// get axes
	r = rune(obj.axes[0])
	str += string(r)
	r = rune(obj.axes[1])
	str += string(r)
	fmt.Println("current controlled axes: ", str)
	fmt.Println("--------------------------------")
	//var data uint16
	data, _ := strconv.Atoi(str)
	fmt.Println(data)

	return ret, 0
}

// typedef struct odbnc {
// 	union {
// 	   struct {
// 			short   reg_prg ;   /* Number of registered programs. */
// 			short   unreg_prg ; /* Number of available programs. */
// 			long    used_mem ;  /* Character number of used memory. */
// 			long    unused_mem ;/* Character number of unused memory. */
// 	   } bin ;
// 	   char asc[31] ;           /* Buffer for ASCII format */
// 	} u ;
// } ODBNC ;
func (client *Client) CncReadProgramInfo() (ret C.short, result int32) {
	obj := C.ODBNC{}
	ret = C.cnc_rdproginfo(client.handle, 0, 12, &obj)
	if ret == C.EW_OK {
		fmt.Println(obj.u)
		// refer to https://stackoverflow.com/questions/14581063/golang-cgo-converting-union-field-to-go-type
		var union [32]byte = obj.u // The union

		var addr *byte = &union[0]
		var cast *C.short = (*C.short)(unsafe.Pointer(addr))
		var reg_prg C.short = *cast
		fmt.Println("reg_prg: ", reg_prg)

		var addr2 *byte = &union[2]
		var cast2 *C.short = (*C.short)(unsafe.Pointer(addr2))
		var unreg_prg C.short = *cast2
		fmt.Println("unreg_prg: ", unreg_prg)

		var addr3 *byte = &union[4]
		//var cast3 *C.long = (*C.long)(unsafe.Pointer(addr3))
		//var used_mem C.long = *cast3
		var cast3 *int32 = (*int32)(unsafe.Pointer(addr3))
		var used_mem int32 = *cast3
		fmt.Println("used_mem: ", used_mem)

		var addr4 *byte = &union[8]
		// var cast4 *C.long = (*C.long)(unsafe.Pointer(addr4))
		// var unused_mem C.long = *cast4
		var cast4 *int32 = (*int32)(unsafe.Pointer(addr4))
		var unused_mem int32 = *cast4
		fmt.Println("unused_mem: ", unused_mem)

		fmt.Println("--------------------------------")
	}

	return ret, 0
}

func (client *Client) CncReadExecPrgName() (ret C.short, result int32) {
	obj := C.ODBEXEPRG{}
	ret = C.cnc_exeprgname(client.handle, &obj)
	if ret == C.EW_OK {
		name := C.GoStringN((*C.char)(unsafe.Pointer(&obj.name)), 32)
		fmt.Println("program name: ", name)
		fmt.Println("program num: ", obj.o_num)
		fmt.Println("--------------------------------")
	}

	return ret, 0
}

func (client *Client) CncReadProcTime() (ret C.short, result int32) {
	obj := C.ODBPTIME{}
	ret = C.cnc_rdproctime(client.handle, &obj)
	fmt.Println("The number of processing time stamp data: ", obj.num)
	fmt.Println("--------------------------------")

	return ret, 0
}

func (client *Client) GetAbsolute() {
	obj := C.ODBAXIS{}
	var ret C.short
	if ret = C.cnc_absolute(client.handle, 1, 4+4*C.MAX_AXIS, &obj); ret == C.EW_OK {
		fmt.Println(obj)
		//fmt.Printf("1:%v\n 2:%v\n 3:%v\n", obj.data[0], obj.data[1], obj.data[2])
		fmt.Println("--------------wp-------------------")
		fmt.Println(float64(obj.data[0]) * math.Pow(10, -3))
	}
	fmt.Println(ret)
}

func (client *Client) GetMachine() {
	obj := C.ODBAXIS{}
	C.cnc_machine(client.handle, 1, 4+4*1, &obj)
	fmt.Println(obj)
	fmt.Printf("MACHINE 1:%v\n", obj.data[0])
	fmt.Printf("MACHINE 2:%v\n", obj.data[1])
	fmt.Printf("MACHINE 3:%v\n", obj.data[2])
	fmt.Printf("MACHINE 4:%v\n", obj.data[3])
	fmt.Printf("MACHINE 5:%v\n", obj.data[5])
}

func (client *Client) GetPosition() {
	obj := C.ODBPOS{}
	var num C.short = C.MAX_AXIS
	C.cnc_rdposition(client.handle, 0, &num, &obj)
	// fmt.Println(obj.abs.name)
	// fmt.Println(float64(obj.abs.data) * math.Pow(10, -int(obj.abs.dec)))

	// //fmt.Println("当前工件数：", obj2.mcr_val/C.long(math.Pow(10, float64(obj2.dec_val))))
	// fmt.Println(obj)
	// fmt.Println(obj.abs)
}

const (
	CNC_POWER_ON_TIME = iota
	CNC_OPERATING_TIME
	CNC_CUTTING_TIME
	CNC_CYCLE_TIME
)

func (client *Client) CncReadTimer() (ret C.short, result int32) {
	obj := C.IODBTIME{}
	ret = C.cnc_rdtimer(client.handle, CNC_POWER_ON_TIME, &obj)
	fmt.Println("minute: ", obj.minute)
	fmt.Println("msec: ", obj.msec)
	fmt.Println("-------------------------")

	return ret, 0
}

// The unit of result is mins
func (client *Client) CncReadPowerOnTime() (ret C.short, result uint32) {
	obj := C.IODBPSD{}
	if ret = C.cnc_rdparam(client.handle, 6750, 0, 4+C.MAX_AXIS, &obj); ret == C.EW_OK {
		u := unsafe.Pointer(uintptr(unsafe.Pointer(&obj)) + 4)
		ldata := *(*uint32)(u)

		// var union [512]byte = obj.u // The union

		// var addr *byte = &union[0]
		// var cast *C.long = (*C.long)(unsafe.Pointer(addr))
		// var ldata C.long = *cast

		fmt.Println("power on time: ", ldata)
		return ret, ldata
	}
	return ret, 0
}

func (client *Client) CncReadOperatingTime() (ret C.short, result uint32) {
	obj := C.IODBPSD{}

	if ret = C.cnc_rdparam(client.handle, 6752, 0, 4+C.MAX_AXIS, &obj); ret == C.EW_OK {
		u := unsafe.Pointer(uintptr(unsafe.Pointer(&obj)) + 4)
		ldata := *(*uint32)(u)

		// var union [512]byte = obj.u // The union

		// var addr *byte = &union[0]
		// var cast *C.long = (*C.long)(unsafe.Pointer(addr))
		// var ldata C.long = *cast
		fmt.Println("operating time: ", ldata)
		return ret, ldata
	}
	return ret, 0
}

func (client *Client) CncReadCuttingTime() (ret C.short, result uint32) {
	obj := C.IODBPSD{}

	if ret = C.cnc_rdparam(client.handle, 6754, 0, 4+C.MAX_AXIS, &obj); ret == C.EW_OK {
		u := unsafe.Pointer(uintptr(unsafe.Pointer(&obj)) + 4)
		ldata := *(*uint32)(u)

		// var union [512]byte = obj.u // The union

		// var addr *byte = &union[0]
		// var cast *C.long = (*C.long)(unsafe.Pointer(addr))
		// var ldata C.long = *cast
		fmt.Println("cutting time: ", ldata)
		return ret, ldata
	}
	return ret, 0
}

func (client *Client) CncReadCycleTime() (ret C.short, result uint32) {
	obj := C.IODBPSD{}

	if ret = C.cnc_rdparam(client.handle, 6758, 0, 4+C.MAX_AXIS, &obj); ret == C.EW_OK {
		u := unsafe.Pointer(uintptr(unsafe.Pointer(&obj)) + 4)
		ldata := *(*uint32)(u)

		// var union [512]byte = obj.u // The union

		// var addr *byte = &union[0]
		// var cast *C.long = (*C.long)(unsafe.Pointer(addr))
		// var ldata C.long = *cast
		fmt.Println("cycle time: ", ldata)
		return ret, ldata
	}
	return ret, 0
}

func (client *Client) PmcRdPmcRange() (ret C.short, result int32) {
	obj := C.IODBPMC{}
	ret = C.pmc_rdpmcrng(client.handle, 0, 1, 30, 31, 8+1*2, &obj)
	u := unsafe.Pointer(uintptr(unsafe.Pointer(&obj)) + 8)
	ldata := *(*uint16)(u)
	fmt.Println("主轴倍率: ", ldata)

	obj2 := C.IODBPMC{}
	ret = C.pmc_rdpmcrng(client.handle, 0, 1, 12, 13, 8+1*2, &obj2)
	u = unsafe.Pointer(uintptr(unsafe.Pointer(&obj2)) + 8)
	data := *(*uint16)(u)
	// 经测试可以读取寄存器G0012的值，读取到的数值与实际倍率存在如下关系     255-G0012=实际倍率？？？？
	fmt.Println("进给倍率: ", data)
	fmt.Println("xxxxxxxxxxxx:", obj2)

	obj3 := C.ODBACT{}
	ret = C.cnc_actf(client.handle, &obj3)
	fmt.Println("实际进给速率: ", obj3.data)
	fmt.Println(unsafe.Sizeof(obj3.data))

	obj4 := C.ODBACT{}
	ret = C.cnc_acts(client.handle, &obj4)
	fmt.Println("实际主轴速率： ", obj4.data)
	fmt.Println("--------------------------------")

	obj5 := C.ODBM{}
	ret5 := C.cnc_rdmacro(client.handle, 4320, 10, &obj5)
	fmt.Println(ret5)
	if ret5 == C.EW_OK {
		fmt.Println("宏变量：", obj5)
	}

	return ret, 0
}

func (client *Client) CncReadCncID() (ret C.short, result int32) {
	var cnc_ids [4]uint32
	if ret := C.cnc_rdcncid(client.handle, (*C.ulong)(unsafe.Pointer(&cnc_ids[0]))); ret != C.EW_OK {
		fmt.Printf("cnc_rdcncid failed (%d)\n", ret)
		return ret, 0
	}
	machine_id := fmt.Sprintf("%08x-%08x-%08x-%08x", cnc_ids[0], cnc_ids[1], cnc_ids[2], cnc_ids[3])

	fmt.Printf("machine_id: %s\n", machine_id)

	return ret, 0
}

func (client *Client) test() {
	fmt.Println("----------test-start--------------")
	obj1 := C.ODBTLIFE2{}
	res := C.cnc_rdngrp(client.handle, &obj1)
	fmt.Println("res:", res)
	fmt.Println("刀具组数:", obj1)

	// obj := C.ODBSPN{}
	// C.cnc_rdspload(client.handle, 9, &obj)

	// obj := C.IODBTD2{}
	// C.cnc_rd1tlifedat2(client.handle, 0, 0, &obj)
	// fmt.Println(obj)

	var sv [C.MAX_AXIS]C.ODBSVLOAD
	var num C.short = 3 //C.MAX_AXIS
	ret := C.cnc_rdsvmeter(client.handle, &num, &sv[0])
	if ret == C.EW_OK {
		fmt.Println(sv)
		var i C.short = 0
		for i = 0; i < num; i++ {
			fmt.Printf("%v = %v\n", sv[i].svload.name, sv[i].svload.data)
		}
	}

	fmt.Println("----------test-end--------------")
}

func (client *Client) test2() {
	obj := C.IODBTD2{}
	ret := C.cnc_rd1tlifedat2(client.handle, 0, 0, &obj)
	fmt.Println("ret=", ret)
	fmt.Printf("--wp--debug--data:%v\n", obj)
}

//------------------------------------------------------
func main() {
	client := NewClient("192.168.11.129", 8193, 10)
	client.StartupProcess()
	defer client.ExitProcess()
	client.Connect()
	defer client.DisConnect()

	if client.Connected {
		client.CncProducts()
		client.CncStatInfo()
		client.CncSysInfo()
		client.CncReadTimer()
		client.CncReadProcTime()
		client.CncReadProgramInfo()
		client.CncReadExecPrgName()
		client.PmcRdPmcRange()
		//client.CncReadCncID()
		client.CncReadPowerOnTime()
		client.CncReadOperatingTime()
		client.CncReadCuttingTime()
		client.CncReadCycleTime()
		client.test()
		client.test2()
		// client.GetAbsolute()
		// client.GetMachine()
		// client.GetPosition()
	}
}
